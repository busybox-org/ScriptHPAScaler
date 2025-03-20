package manager

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	watchtools "k8s.io/client-go/tools/watch"
	log "k8s.io/klog/v2"
)

func (e *sExecutor) waitPodReady(gvr schema.GroupVersionResource, name string, eventFn func(message string)) {
	log.Infof("%s/%s waiting for the expansion to complete", e.namespace, name)
	defer log.Infof("%s/%s expansion completed", e.namespace, name)
	eventFn("Waiting for the expansion to complete")
	var fn func(obj runtime.Unstructured) (string, bool, error)
	if gvr.Resource == "deployments" {
		fn = e.watchDeployment
	} else if gvr.Resource == "statefulsets" {
		fn = e.watchStatefulSet
	} else {
		return
	}
	fieldSelector := fields.OneTermEqualSelector("metadata.name", name).String()
	lw := &cache.ListWatch{
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			options.FieldSelector = fieldSelector
			return e.DynamicClient().Resource(gvr).Namespace(e.namespace).List(context.Background(), options)
		},
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			options.FieldSelector = fieldSelector
			return e.DynamicClient().Resource(gvr).Namespace(e.namespace).Watch(context.Background(), options)
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	_, err := watchtools.UntilWithSync(ctx, lw, &unstructured.Unstructured{}, nil, func(event watch.Event) (bool, error) {
		switch t := event.Type; t {
		case watch.Added, watch.Modified:
			status, done, err := fn(event.Object.(runtime.Unstructured))
			if err != nil {
				return false, err
			}
			eventFn(status)
			if done {
				return true, nil
			}
			return false, nil
		case watch.Deleted:
			// We need to abort to avoid cases of recreation and not to silently watch the wrong (new) object
			return true, fmt.Errorf("object has been deleted")

		default:
			return true, fmt.Errorf("internal error: Unexpected event %#v", event)
		}
	})
	if err != nil {
		log.Errorln(err)
		return
	}
	eventFn("Expansion completed")
}

func (e *sExecutor) watchDeployment(obj runtime.Unstructured) (string, bool, error) {
	deployment := &appsv1.Deployment{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), deployment)
	if err != nil {
		return "", false, fmt.Errorf("failed to convert %T to %T: %v", obj, deployment, err)
	}

	if deployment.Generation <= deployment.Status.ObservedGeneration {
		cond := e.getResourceCondition(deployment.Status, appsv1.DeploymentProgressing)
		if cond != nil && cond.Reason == "ProgressDeadlineExceeded" {
			return "", false, fmt.Errorf("deployment %s/%s exceeded its progress deadline",
				deployment.Namespace, deployment.Name)
		}
		if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			return fmt.Sprintf("waiting for deployment %s/%s rollout to finish: %d out of %d new replicas have been updated...",
				deployment.Namespace, deployment.Name, deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas), false, nil
		}
		if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			return fmt.Sprintf("waiting for deployment %s/%s rollout to finish: %d old replicas are pending termination...",
				deployment.Namespace, deployment.Name, deployment.Status.Replicas-deployment.Status.UpdatedReplicas), false, nil
		}
		if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			return fmt.Sprintf("waiting for deployment %s/%s rollout to finish: %d of %d updated replicas are available...",
				deployment.Namespace, deployment.Name, deployment.Status.UpdatedReplicas, deployment.Status.AvailableReplicas), false, nil
		}
		return fmt.Sprintf("deployment %s/%s successfully rolled out",
			deployment.Namespace, deployment.Name), true, nil
	}
	return fmt.Sprintf("waiting for deployment %s/%s spec update to be observed...",
		deployment.Namespace, deployment.Name), false, nil
}

func (e *sExecutor) watchStatefulSet(obj runtime.Unstructured) (string, bool, error) {
	sts := &appsv1.StatefulSet{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), sts)
	if err != nil {
		return "", false, fmt.Errorf("failed to convert %T to %T: %v", obj, sts, err)
	}
	if sts.Spec.UpdateStrategy.Type != appsv1.RollingUpdateStatefulSetStrategyType {
		return "", true, fmt.Errorf("rollout status is only available for %s strategy type", appsv1.RollingUpdateStatefulSetStrategyType)
	}
	if sts.Status.ObservedGeneration == 0 || sts.Generation > sts.Status.ObservedGeneration {
		return fmt.Sprintf("waiting for stateful set %s/%s spec update to be observed...",
			sts.Namespace, sts.Name), false, nil
	}
	if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
		return fmt.Sprintf("waiting for stateful set %s/%s %d pods to be ready...",
			sts.Namespace, sts.Name, *sts.Spec.Replicas-sts.Status.ReadyReplicas), false, nil
	}
	if (sts.Spec.Replicas == nil || *sts.Spec.Replicas == 0) && sts.Status.UpdatedReplicas > 0 {
		return fmt.Sprintf("waiting for stateful set %s/%s %d pods to be delete...",
			sts.Namespace, sts.Name, sts.Status.UpdatedReplicas), false, nil
	}
	if sts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil {
		if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
				return fmt.Sprintf("waiting for stateful set %s/%s  partitioned roll out to finish: %d out of %d new pods have been updated...",
					sts.Namespace, sts.Name, sts.Status.UpdatedReplicas, *sts.Spec.Replicas-*sts.Spec.UpdateStrategy.RollingUpdate.Partition), false, nil
			}
		}
		return fmt.Sprintf("stateful set %s/%s partitioned roll out complete: %d new pods have been updated...",
			sts.Namespace, sts.Name, sts.Status.UpdatedReplicas), true, nil
	}
	if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
		return fmt.Sprintf("waiting for stateful set %s/%s rolling update to complete %d pods at revision %s...",
			sts.Namespace, sts.Name, sts.Status.UpdatedReplicas, sts.Status.UpdateRevision), false, nil
	}
	return fmt.Sprintf("stateful set %s/%s rolling update complete %d pods at revision %s...",
		sts.Namespace, sts.Name, sts.Status.CurrentReplicas, sts.Status.CurrentRevision), true, nil
}

func (e *sExecutor) getResourceCondition(status appsv1.DeploymentStatus, condType appsv1.DeploymentConditionType) *appsv1.DeploymentCondition {
	for i := range status.Conditions {
		c := status.Conditions[i]
		if c.Type == condType {
			return &c
		}
	}
	return nil
}
