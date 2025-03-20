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

package manager

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/record"
	log "k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	busyboxorgv1alpha1 "busybox.org/scripthpascaler/api/v1alpha1"
)

type IManager interface {
	ISelfManager
	Run(stopChan chan struct{})
	AddExecutor(namespace, name string) error
	Clean(namespacedName string)
	FilterItem(item busyboxorgv1alpha1.ScriptHPAScalerSpec) bool
}

type ISelfManager interface {
	Client() client.Client
	ScaleClient() scale.ScalesGetter
	DynamicClient() *dynamic.DynamicClient
	EventRecorder() record.EventRecorder
	StabilizeMember(key string, value int64, timDuration time.Duration) int64
	UpdateStatus(item *busyboxorgv1alpha1.ScriptHPAScaler, status busyboxorgv1alpha1.ScriptHPAScalerStatus) error
}

type sManager struct {
	sync.Mutex
	cron          *cron.Cron
	cfg           *rest.Config
	client        client.Client
	scaleClient   scale.ScalesGetter
	eventRecorder record.EventRecorder
	dynamicClient *dynamic.DynamicClient
	executorCache sync.Map
	members       map[string][]memberValue
}

type memberValue struct {
	timestamp time.Time
	value     int64
}

func New(mgr ctrl.Manager) IManager {
	m := &sManager{
		cron: cron.New(
			cron.WithSeconds(),
			cron.WithLocation(time.Now().Location()),
		),
		cfg:           mgr.GetConfig(),
		client:        mgr.GetClient(),
		eventRecorder: mgr.GetEventRecorderFor("HPAScalerController"),
		dynamicClient: dynamic.NewForConfigOrDie(mgr.GetConfig()),
		members:       make(map[string][]memberValue),
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(m.cfg)
	if err != nil {
		log.Fatalf("Failed to create discovery client: %v", err)
	}
	scaleKindResolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)
	m.scaleClient, err = scale.NewForConfig(m.cfg, mgr.GetRESTMapper(), dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)
	if err != nil {
		log.Fatalf("Failed to create scale client: %v", err)
	}
	go m.cron.Run()
	return m
}

// StabilizeMember :
// - replaces old value with the newest value,
// - returns max of values that are not older than stabilisationWindow.
func (m *sManager) StabilizeMember(key string, value int64, timDuration time.Duration) (res int64) {
	m.Lock()
	defer m.Unlock()

	res = value
	foundOldSample := false
	oldSampleIndex := 0

	cutoff := time.Now().Add(-timDuration)
	for i, r := range m.members[key] {
		if r.timestamp.Before(cutoff) {
			if !foundOldSample {
				oldSampleIndex = i
				foundOldSample = true
			}
		} else if r.value > res {
			res = r.value
		}
	}
	if foundOldSample {
		// If we found an old sample, we'll use the most recent recommendation
		m.members[key][oldSampleIndex] = memberValue{
			value:     value,
			timestamp: time.Now(),
		}
	} else {
		// If we didn't find an old sample, we'll use the current recommendation
		m.members[key] = append(m.members[key], memberValue{
			value:     value,
			timestamp: time.Now(),
		})
	}
	return res
}

func (m *sManager) Client() client.Client {
	return m.client
}

func (m *sManager) ScaleClient() scale.ScalesGetter {
	return m.scaleClient
}

func (m *sManager) DynamicClient() *dynamic.DynamicClient {
	return m.dynamicClient
}

func (m *sManager) EventRecorder() record.EventRecorder {
	return m.eventRecorder
}

func (m *sManager) Clean(namespacedName string) {
	m.Lock()
	defer m.Unlock()
	var id cron.EntryID
	m.executorCache.Range(func(key, value interface{}) bool {
		if key.(string) == namespacedName {
			id = value.(cron.EntryID)
			return false
		}
		return true
	})
	if id == 0 {
		// not found, skip
		return
	}
	m.executorCache.Delete(namespacedName)
	delete(m.members, namespacedName)
	// remove executor
	m.cron.Remove(id)
	log.Infof("Remove HorizontalPodAutoscaler executor %s", namespacedName)
}

func (m *sManager) initLoad() {
	listOptions := &client.ListOptions{}
	list := &busyboxorgv1alpha1.ScriptHPAScalerList{}
	if err := m.client.List(context.TODO(), list, listOptions); err != nil {
		log.Fatalf("init autoscaler manager failed, err: %v", err)
	}
	list.Items = m.filter(list.Items)
	for _, item := range list.Items {
		err := m.AddExecutor(item.Namespace, item.Name)
		if err != nil {
			log.Fatalf("init autoscaler manager failed, err: %v", err)
		}
	}
}

func (m *sManager) filter(items []busyboxorgv1alpha1.ScriptHPAScaler) []busyboxorgv1alpha1.ScriptHPAScaler {
	var result []busyboxorgv1alpha1.ScriptHPAScaler
	for _, item := range items {
		if m.FilterItem(item.Spec) {
			result = append(result, item)
		}
	}
	return result
}

func (m *sManager) FilterItem(item busyboxorgv1alpha1.ScriptHPAScalerSpec) bool {
	switch item.ScaleTargetRef.Kind {
	case "Deployment":
		return true
	case "StatefulSet":
		return true
	}
	return false
}

func (m *sManager) Run(stopChan chan struct{}) {
	m.initLoad()
	<-stopChan
	m.cron.Stop()
}

func (m *sManager) CheckExist(namespacedName string) bool {
	m.Lock()
	defer m.Unlock()
	_, found := m.executorCache.Load(namespacedName)
	return found
}

func (m *sManager) AddExecutor(namespace, name string) error {
	m.Lock()
	defer m.Unlock()

	e := &sExecutor{
		ISelfManager: m,
		namespace:    namespace,
		name:         name,
	}
	e.watchFnMap = map[string]watchFn{
		"Deployment":  e.watchDeployment,
		"StatefulSet": e.watchStatefulSet,
	}

	_, found := m.executorCache.Load(e.NamespacedName())
	if found {
		return errors.New("executor already exist")
	}
	// executor not found
	id, err := m.cron.AddJob(e.SchedulePlan(), e)
	if err != nil {
		log.Errorln(err)
		return err
	}
	m.executorCache.LoadOrStore(e.NamespacedName(), id)
	log.Infof("HorizontalPodAutoscaler %s is created", e.NamespacedName())

	return nil
}

func (m *sManager) UpdateStatus(item *busyboxorgv1alpha1.ScriptHPAScaler, status busyboxorgv1alpha1.ScriptHPAScalerStatus) error {
	m.Lock()
	defer m.Unlock()

	patch := client.MergeFrom(item.DeepCopy())
	item.Status = status
	item.Status.LastProbeTime = metav1.Now()
	if err := m.client.Status().Patch(context.TODO(), item, patch); err != nil {
		log.Errorf("UpdateStatus failed, err: %v", err)
		return err
	}
	return nil
}
