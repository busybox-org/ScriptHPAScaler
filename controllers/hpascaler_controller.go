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
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	log "k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HPAScalerReconciler reconciles a HPAScaler object
type HPAScalerReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	ScalerManager *ScalerManager
}

//+kubebuilder:rbac:groups=k8s.q1.com,resources=hpascalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.q1.com,resources=hpascalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.q1.com,resources=hpascalers/finalizers,verbs=update
//+kubebuilder:rbac:groups=*,resources=*/scale,verbs=get;list;update;patch
//+kubebuilder:rbac:groups=extensions,resources=*,verbs=get;list;watch;create;update
//+kubebuilder:rbac:groups=apps,resources=*,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HPAScaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *HPAScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &k8sq1comv1.HPAScaler{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("HPAScaler %s in %s namespace is deleted", req.Name, req.Namespace)
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			go r.ScalerManager.GC(req.NamespacedName.String())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !r.ScalerManager.FilterItem(instance.Spec) {
		return ctrl.Result{}, nil
	}

	var condition k8sq1comv1.Condition
	needUpdateStatus := true
	job := ScalerJobFactory(*instance, r.ScalerManager.scaleClient, r.Client)
	condition.UID = job.ID()
	condition.Status = k8sq1comv1.Submitted
	condition.Message = fmt.Sprintf("%s is submitted", job.ID())
	err = r.ScalerManager.CreateOrUpdate(job)
	if err != nil {
		if _, ok := err.(*NoNeedUpdate); !ok {
			condition.Status = k8sq1comv1.Failed
			condition.Message = fmt.Sprintf("failed to submitted HPAScaler job %s,because of %v", job.Name(), err)
		} else {
			needUpdateStatus = false
		}
	}

	// update the status
	if needUpdateStatus {
		instance.Status.Condition = condition
		err = r.ScalerManager.updateStatusWithRetry(instance)
		if err != nil {
			log.Errorf("failed to update status of HPAScaler %s in %s namespace,because of %v", req.Name, req.Namespace, err)
			return ctrl.Result{}, err
		}
		log.Infof("HPAScaler %s in %s namespace is updated", req.Name, req.Namespace)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HPAScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var stopCh chan struct{}
	sm := NewScalerManager(mgr)
	go func(sm *ScalerManager, stopCh chan struct{}) {
		sm.Start()
		// wait for the cache to be synced
		cache := mgr.GetCache()
		if !cache.WaitForCacheSync(context.Background()) {
			log.Fatalln("failed to wait for caches to sync")
		}
		sm.Run(stopCh)
		<-stopCh
	}(sm, stopCh)
	r.ScalerManager = sm
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sq1comv1.HPAScaler{}).
		Complete(r)
}
