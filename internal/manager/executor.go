package manager

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"

	busyboxorgv1alpha1 "github.com/busybox-org/scripthpascaler/api/v1alpha1"
	"github.com/busybox-org/scripthpascaler/internal/yaegi"
)

const (
	maxStabWindowTime     = 15 * time.Minute
	minStabWindowTime     = 15 * time.Second
	defaultStabWindowTime = 3 * time.Minute
)

type sExecutor struct {
	ISelfManager
	namespace string
	name      string
	running   int32
}

func (e *sExecutor) NamespacedName() string {
	return fmt.Sprintf("%s/%s", e.namespace, e.name)
}

func (e *sExecutor) SchedulePlan() string {
	// 每三秒检查
	return "@every 1s"
}

func (e *sExecutor) Run() {
	if !atomic.CompareAndSwapInt32(&e.running, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&e.running, 0)

	instance := &busyboxorgv1alpha1.ScriptHPAScaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      e.name,
			Namespace: e.namespace,
		},
	}
	err := e.Client().Get(context.TODO(), types.NamespacedName{
		Name:      e.name,
		Namespace: e.namespace,
	}, instance)
	if err != nil {
		log.Errorln(e.NamespacedName(), "get instance failed, err", err)
		e.EventRecorder().Event(instance, v1.EventTypeWarning, "ScalingReplicaSet", err.Error())
		return
	}

	desiredReplicas, _err := e.scaleReplicas(instance)
	if desiredReplicas == -1 {
		return
	}
	var status busyboxorgv1alpha1.ScriptHPAScalerStatus
	status.DesiredReplicas = desiredReplicas
	status.State = busyboxorgv1alpha1.ScaleStateSuccess
	status.Message = fmt.Sprintf("scaling replicas successful. current replicas is %d", desiredReplicas)
	if _err != nil {
		log.Errorln(e.NamespacedName(), "scale replicas failed, err", _err)
		status.State = busyboxorgv1alpha1.ScaleStateFailure
		status.Message = _err.Error()
	}
	err = e.UpdateStatus(instance, status)
	if err == nil {
		return
	}
	log.Errorln(e.NamespacedName(), "update instance status failed, err", err)
	e.EventRecorder().Event(instance, v1.EventTypeWarning, "ScalingReplicaSet", fmt.Sprintf("Can't update HorizontalPodAutoscaler status: %v", err))
}

func (e *sExecutor) scaleReplicas(item *busyboxorgv1alpha1.ScriptHPAScaler) (int32, error) {
	var name = item.Spec.ScaleTargetRef.Name
	var gvr, err = e.generateGroupVersionResource(item.Spec.ScaleTargetRef.APIVersion, item.Spec.ScaleTargetRef.Kind)
	if err != nil {
		return -1, err
	}
	// 解析稳定窗口，默认 3 分钟
	// 如果稳定窗口超过 15 分钟，则取上限 15 分钟
	timDuration, err := time.ParseDuration(item.Spec.StabilisationWindow)
	if err != nil {
		timDuration = defaultStabWindowTime
	}
	timDuration = max(minStabWindowTime, min(timDuration, maxStabWindowTime))

	current, err := e.getReplicas(gvr.GroupResource(), name)
	if err != nil {
		log.Errorln(e.NamespacedName(), "get replicas failed, err", err)
		return 0, err
	}
	if current == 0 {
		return -1, nil
	}
	metric, err := yaegi.Eval(context.TODO(), item.Spec.Script, map[string]any{
		"Replicas":    current,
		"MinReplicas": item.Spec.MinReplicas,
		"MaxReplicas": item.Spec.MaxReplicas,
	})
	if err != nil {
		return -1, err
	}
	// 最大稳定建议值
	metric = e.StabilizeMember(e.NamespacedName(), metric, timDuration)
	target := max(item.Spec.MinReplicas, min(int32(metric), item.Spec.MaxReplicas))
	var action string
	switch {
	case target < current && target >= item.Spec.MinReplicas:
		action = "down"
	case target > current && target <= item.Spec.MaxReplicas:
		action = "up"
	default:
		return -1, nil
	}
	log.Infof("%s replicas is %d, desired is %d, scaling %s", e.NamespacedName(), current, target, action)
	err = e.updateReplicas(gvr.GroupResource(), name, target)
	if err != nil {
		return -1, fmt.Errorf("failed to scale %v", err)
	}

	e.waitPodReady(gvr, name, func(message string) {
		e.EventRecorder().Event(item, v1.EventTypeNormal, "ScalingReplicaSet", message)
	})
	return target, nil
}

func (e *sExecutor) updateReplicas(gr schema.GroupResource, name string, replicas int32) error {
	newScale, err := e.ScaleClient().Scales(e.namespace).
		Update(context.TODO(), gr, &autoscalingv1.Scale{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: e.namespace,
			},
			Spec: autoscalingv1.ScaleSpec{
				Replicas: replicas,
			},
		}, metav1.UpdateOptions{})
	if err != nil {
		log.Errorf("%s update replicas failed, err %v", e.NamespacedName(), err)
		return err
	}
	log.Infof("%s update replicas success, new replicas: %d", e.NamespacedName(), newScale.Spec.Replicas)
	return nil
}

func (e *sExecutor) getReplicas(gr schema.GroupResource, name string) (replicas int32, err error) {
	_scale, err := e.ScaleClient().Scales(e.namespace).
		Get(context.TODO(), gr, name, metav1.GetOptions{})
	if err == nil {
		return _scale.Spec.Replicas, nil
	}
	log.Errorf("%s get replicas failed, err %v", e.NamespacedName(), err)
	return
}

func (e *sExecutor) generateGroupVersionResource(gv, kind string) (schema.GroupVersionResource, error) {
	_gv, err := schema.ParseGroupVersion(gv)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	switch strings.ToLower(kind) {
	case "deployment", "deployments":
		return _gv.WithResource("deployments"), nil
	case "statefulset", "statefulsets":
		return _gv.WithResource("statefulsets"), nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("unknown kind %s", kind)
	}
}
