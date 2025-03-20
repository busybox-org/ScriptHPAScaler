/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	log "k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	busyboxorgv1alpha1 "busybox.org/scripthpascaler/api/v1alpha1"
	"busybox.org/scripthpascaler/internal/manager"
)

// ScriptHPAScalerReconciler reconciles a ScriptHPAScaler object
type ScriptHPAScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	manager.IManager
}

// +kubebuilder:rbac:groups=busybox.org,resources=scripthpascalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=busybox.org,resources=scripthpascalers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=busybox.org,resources=scripthpascalers/finalizers,verbs=update
// +kubebuilder:rbac:groups=*,resources=*/scale,verbs=get;list;update;patch
// +kubebuilder:rbac:groups=extensions,resources=*,verbs=get;list;watch;create;update
// +kubebuilder:rbac:groups=apps,resources=*,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=validatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScriptHPAScaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *ScriptHPAScalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &busyboxorgv1alpha1.ScriptHPAScaler{}
	err := r.Client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Infof("HorizontalPodAutoscaler %s in %s namespace is deleted", req.Name, req.Namespace)
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			go r.Clean(req.NamespacedName.String())
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !r.FilterItem(instance.Spec) {
		return ctrl.Result{}, nil
	}
	var status busyboxorgv1alpha1.ScriptHPAScalerStatus
	status.State = busyboxorgv1alpha1.ScaleStatePending
	status.Message = fmt.Sprintf("%s is submitted", req.NamespacedName.String())
	err = r.AddExecutor(instance.Namespace, instance.Name)
	if err != nil {
		if err.Error() == "executor already exist" {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// update the status
	err = r.UpdateStatus(instance, status)
	if err != nil {
		log.Errorf("failed to update status of HorizontalPodAutoscaler %s,because of %v", req.NamespacedName, err)
		return ctrl.Result{}, err
	}
	log.Infof("HorizontalPodAutoscaler %s is updated", req.NamespacedName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScriptHPAScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var stopCh chan struct{}
	r.IManager = manager.New(mgr)
	go func(sm manager.IManager, stopCh chan struct{}) {
		// wait for the cache to be synced
		cache := mgr.GetCache()
		if !cache.WaitForCacheSync(context.Background()) {
			log.Fatalln("failed to wait for caches to sync")
		}
		sm.Run(stopCh)
		<-stopCh
	}(r.IManager, stopCh)
	return ctrl.NewControllerManagedBy(mgr).
		For(&busyboxorgv1alpha1.ScriptHPAScaler{}).
		Named("scripthpascaler").
		Complete(r)
}
