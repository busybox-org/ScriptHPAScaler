/*
Copyright 2022 xmapst@gmail.com.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/ringtail/go-cron"
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/record"
	log "k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sync"
)

const MaxRetryTimes = 3

type NoNeedUpdate struct{}

func (n NoNeedUpdate) Error() string {
	return "NoNeedUpdate"
}

type ScalerManager struct {
	sync.Mutex
	cfg           *rest.Config
	client        client.Client
	jobQueue      map[string]ScalerJob // jobID -> ScalerJob
	executor      ScalerExecutor
	scaleClient   scale.ScalesGetter
	eventRecorder record.EventRecorder
}

func NewScalerManager(mgr ctrl.Manager) *ScalerManager {
	sm := &ScalerManager{
		cfg:           mgr.GetConfig(),
		client:        mgr.GetClient(),
		jobQueue:      make(map[string]ScalerJob),
		eventRecorder: mgr.GetEventRecorderFor("hpascaler-controller"),
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(sm.cfg)
	if err != nil {
		log.Fatalf("Failed to create discovery client: %v", err)
	}
	scaleKindResolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	sm.scaleClient, err = scale.NewForConfig(sm.cfg, mgr.GetRESTMapper(), dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)
	if err != nil {
		log.Fatalf("Failed to create scale client: %v", err)
	}
	sm.executor = NewScalerHPAExecutor(nil, sm.resultHandler)
	return sm
}

func (sm *ScalerManager) getJobByNamespacedName(namespacedName string) (ScalerJob, bool) {
	sm.Lock()
	defer sm.Unlock()
	for _, job := range sm.jobQueue {
		if job.Namespace()+"/"+job.Name() == namespacedName {
			return job, true
		}
	}
	return nil, false
}

func (sm *ScalerManager) GC(namespacedName string) {
	job, ok := sm.getJobByNamespacedName(namespacedName)
	if !ok {
		// job not found
		return
	}
	sm.Lock()
	defer sm.Unlock()
	// remove job from queue
	defer delete(sm.jobQueue, job.ID())
	// stop job
	err := sm.executor.RemoveJob(job)
	if err != nil {
		log.Errorf("Failed to remove job %s: %v", namespacedName, err)
	}
	log.Infof("Remove HPAScaler job %s", namespacedName)
}

func (sm *ScalerManager) init() {
	listOptions := &client.ListOptions{}
	list := &k8sq1comv1.HPAScalerList{}
	err := sm.client.List(context.TODO(), list, listOptions)
	if err != nil {
		log.Fatalf("init autoscaler manager failed, err: %v", err)
	}
	list.Items = sm.filter(list.Items)
	for _, v := range list.Items {
		job := ScalerJobFactory(v, sm.scaleClient, sm.client)
		sm.jobQueue[job.ID()] = job
		err = sm.executor.AddJob(job)
		if err != nil {
			log.Errorf("Failed to add job %s: %v", v.Namespace+"/"+v.Name, err)
		}
		log.Infof("Add HPAScaler job %s", v.Namespace+"/"+v.Name)
	}
}

func (sm *ScalerManager) filter(items []k8sq1comv1.HPAScaler) []k8sq1comv1.HPAScaler {
	var result []k8sq1comv1.HPAScaler
	for _, item := range items {
		if sm.FilterItem(item.Spec) {
			result = append(result, item)
		}
	}
	return result
}

func (sm *ScalerManager) FilterItem(item k8sq1comv1.HPAScalerSpec) bool {
	switch item.ScaleTargetRef.Kind {
	case "Deployment":
		return true
	case "StatefulSet":
		return true
	case "ReplicaSet":
		return true
	}
	return false
}

func (sm *ScalerManager) Start() {
	sm.executor.Run()
}

func (sm *ScalerManager) Run(stopChan chan struct{}) {
	sm.init()
	<-stopChan
	sm.executor.Stop()
}

func (sm *ScalerManager) CreateOrUpdate(job ScalerJob) error {
	sm.Lock()
	defer sm.Unlock()
	if _, ok := sm.jobQueue[job.ID()]; !ok {
		// job not found
		err := sm.executor.AddJob(job)
		if err != nil {
			return fmt.Errorf("failed to add AutoScaler job %s/%s: %v", job.Namespace(), job.Name(), err)
		}
		sm.jobQueue[job.ID()] = job
		log.Infof("AutoScaler %s in %s namespace is created, %d active jobs exist", job.Name(), job.Namespace(), len(sm.jobQueue))
	} else {
		j := sm.jobQueue[job.ID()]
		if !j.Equals(job) {
			err := sm.executor.Update(job)
			if err != nil {
				return fmt.Errorf("failed to update AutoScaler job %s/%s: %v", job.Namespace(), job.Name(), err)
			}
			sm.jobQueue[job.ID()] = job
		} else {
			return &NoNeedUpdate{}
		}
	}
	return nil
}

// ResultHandler is a function that is called when a cron job's execution
func (sm *ScalerManager) resultHandler(js *cron.JobResult) {
	for i := 0; i < MaxRetryTimes; i++ {
		job, ok := js.Ref.(*ScalerJobHPA)
		if !ok {
			log.Errorf("job result handler failed")
			break
		}
		instance := &k8sq1comv1.HPAScaler{}
		err := sm.client.Get(context.TODO(), types.NamespacedName{
			Namespace: job.Namespace(),
			Name:      job.Name(),
		}, instance)
		if err != nil {
			log.Errorf("job result handler failed, get instance failed, err: %v", err)
			continue
		}
		instance.Status.Condition.Status = k8sq1comv1.Succeed
		instance.Status.Condition.Message = js.Msg

		err = js.Error
		if err != nil {
			if _, ok := err.(*NoNeedUpdate); ok {
				break
			}
			instance.Status.Condition.Status = k8sq1comv1.Failed
			instance.Status.Condition.Message = err.Error()
		} else {
			if instance.Status.Condition.DesiredReplicas == job.DesiredReplicas() {
				break
			}
			instance.Status.Condition.DesiredReplicas = job.DesiredReplicas()
		}
		err = sm.updateStatus(instance)
		if err == nil {
			break
		}
		log.Errorf("job result handler failed, update instance status failed, err: %v", err)
		sm.eventRecorder.Event(instance, v1.EventTypeWarning, "Warning", fmt.Sprintf("Can't update HPAScaler status: %v", err))
	}
}

func (sm *ScalerManager) updateStatus(item *k8sq1comv1.HPAScaler) error {
	item.Status.Condition.LastProbeTime = metav1.Now()
	err := sm.client.Status().Update(context.TODO(), item)
	if err != nil {
		return err
	}
	return nil
}
