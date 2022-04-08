package controllers

import (
	"context"
	"fmt"
	k8sq1comv1 "github.com/xmapst/supersetscalers/api/v1"
	"github.com/xmapst/supersetscalers/plugins"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

const (
	updateRetryInterval = 3 * time.Second
	maxRetryTimeout     = 10 * time.Second
)

type ScalerJob interface {
	ID() string
	Name() string
	Namespace() string
	SetID(id string)
	Equals(Job ScalerJob) bool
	SchedulePlan() string
	DesiredReplicas() int32
	Run() (string, error)
}

type ScalerJobHPA struct {
	hpaSpec         k8sq1comv1.HPAScalerSpec
	namespace       string
	id              string
	name            string
	scaleClient     scale.ScalesGetter
	desiredReplicas int32
	client          client.Client
	recommendations []timestampedRecommendation
}

type timestampedRecommendation struct {
	recommendation int32
	timestamp      time.Time
}

func (sh *ScalerJobHPA) SetID(id string) {
	sh.id = id
}

func (sh *ScalerJobHPA) Name() string {
	return sh.name
}

func (sh *ScalerJobHPA) Namespace() string {
	return sh.namespace
}

func (sh *ScalerJobHPA) ID() string {
	return sh.id
}

func (sh *ScalerJobHPA) DesiredReplicas() int32 {
	return sh.desiredReplicas
}

func (sh *ScalerJobHPA) Equals(j ScalerJob) bool {
	// update will create a new uuid
	if sh.id == j.ID() && sh.hpaSpec.ToString() == j.(*ScalerJobHPA).hpaSpec.ToString() {
		return true
	}
	return false
}

func (sh *ScalerJobHPA) SchedulePlan() string {
	return fmt.Sprintf("@every %s", sh.hpaSpec.Freq)
}

func (sh *ScalerJobHPA) Run() (msg string, err error) {
	startTime := time.Now()
	times := 0
	for {
		now := time.Now()
		// timeout and exit
		if startTime.Add(maxRetryTimeout).Before(now) {
			return "", fmt.Errorf("failed to scale %s in %s namespace after retrying %d times and exit,because of %v", sh.name, sh.namespace, times, err)
		}
		msg, err = sh.Scale()
		if err == nil {
			break
		}
		if _, ok := err.(*NoNeedUpdate); ok {
			break
		}
		time.Sleep(updateRetryInterval)
		times = times + 1
	}
	return msg, err
}

func (sh *ScalerJobHPA) Scale() (msg string, err error) {
	ready, err := plugins.Plugins[sh.hpaSpec.Plugin.Type]().Run(&sh.hpaSpec.Plugin)
	if err != nil {
		return "", fmt.Errorf("failed to run plugin %s, because of %v", sh.hpaSpec.Plugin.Type, err)
	}
	replicas, err := sh.getReplicas()
	if err != nil || replicas == 0 {
		return "replicas is 0, skip", err
	}

	desired := sh.targetReplicas(ready, replicas)
	// If the number of instances calculated this time is greater than the current number of instances,
	// the expansion will be triggered immediately,but the scaling will not be triggered
	// immediately if the number of instances is lower than the current number of instances.
	// In order to avoid the regular jitter of the Pod resource utilization,
	// the expansion and contraction will be performed frequently.
	if desired > replicas {
		msg = fmt.Sprintf("replicas is %d, desired is %d, scaling up", replicas, desired)
	} else {
		// downscaleStabilisationWindow is the time for which the replica count needs to be stable before it is eligible for downscaling.
		// It's also the time we give a chance to the controller manager to distribute the update when we are scaling down.
		desired = sh.stabilizeRecommendation(desired)
		if desired < replicas {
			msg = fmt.Sprintf("replicas is %d, desired is %d, scaling down", replicas, desired)
		} else {
			return "noting to do", &NoNeedUpdate{}
		}
	}
	sh.desiredReplicas = desired
	err = sh.updateReplicas()
	if err != nil {
		return "", fmt.Errorf("failed to scale %s in %s namespace, because of %v", sh.name, sh.namespace, err)
	}
	return msg, nil
}

func (sh *ScalerJobHPA) stabilizeRecommendation(prenormalizedDesiredReplicas int32) int32 {
	maxRecommendation := prenormalizedDesiredReplicas
	foundOldSample := false
	oldSampleIndex := 0
	// default stabilization window is 3 minutes.
	downscaleStabilisationWindow, err := time.ParseDuration(sh.hpaSpec.DownscaleStabilisationWindow)
	if err != nil {
		downscaleStabilisationWindow = 3 * time.Minute
	}
	// if downscaleStabilisationWindow more than 15 time.Minute, use default value instead.
	if downscaleStabilisationWindow > 15*time.Minute {
		log.Info("downscaleStabilisationWindow is too long, forcibly set it to 3 minutes")
		downscaleStabilisationWindow = 15 * time.Minute
	}
	cutoff := time.Now().Add(-downscaleStabilisationWindow)
	// if we find a recommendation higher than the current max recommendation,
	// set the current max recommendation to the current recommendation
	// and set the index of the recommendation to the current index
	// and set the foundOldSample to false
	// so that we can find the next higher recommendation
	// and set the oldSampleIndex to the current index
	// and set the foundOldSample to true
	for i, r := range sh.recommendations {
		if r.timestamp.Before(cutoff) {
			if !foundOldSample {
				oldSampleIndex = i
				foundOldSample = true
			}
		} else if r.recommendation > maxRecommendation {
			maxRecommendation = r.recommendation
		}
	}
	if foundOldSample {
		// If we found an old sample, we'll use the most recent recommendation
		sh.recommendations[oldSampleIndex] = timestampedRecommendation{
			recommendation: prenormalizedDesiredReplicas,
			timestamp:      time.Now(),
		}
	} else {
		// If we didn't find an old sample, we'll use the current recommendation
		sh.recommendations = append(sh.recommendations, timestampedRecommendation{
			recommendation: prenormalizedDesiredReplicas,
			timestamp:      time.Now(),
		})
	}
	return maxRecommendation
}

func (sh *ScalerJobHPA) updateReplicas() error {
	s := &autoscalingv1.Scale{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sh.hpaSpec.ScaleTargetRef.Name,
			Namespace: sh.Namespace(),
		},
		Spec: autoscalingv1.ScaleSpec{
			Replicas: sh.desiredReplicas,
		},
	}

	newScale, err := sh.scaleClient.Scales(sh.Namespace()).
		Update(context.TODO(), sh.generateGroupResource(), s, metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	log.Infof("update replicas success, new replicas: %d", newScale.Spec.Replicas)
	return nil
}

func (sh *ScalerJobHPA) getReplicas() (replicas int32, err error) {
	_scale, err := sh.scaleClient.Scales(sh.Namespace()).
		Get(context.TODO(), sh.generateGroupResource(), sh.hpaSpec.ScaleTargetRef.Name, metav1.GetOptions{})
	if err == nil {
		return _scale.Spec.Replicas, nil
	}
	return
}

func (sh *ScalerJobHPA) generateGroupResource() (groupResource schema.GroupResource) {
	switch sh.hpaSpec.ScaleTargetRef.Kind {
	case "Deployment":
		groupResource = schema.GroupResource{Group: appsv1.GroupName, Resource: "deployments"}
	case "StatefulSet":
		groupResource = schema.GroupResource{Group: appsv1.GroupName, Resource: "statefulsets"}
	case "Replicaset":
		groupResource = schema.GroupResource{Group: appsv1.GroupName, Resource: "replicasets"}
	}
	return
}

// targetReplicas returns the desired replicas of a given pod,
// based on its current size.
func (sh *ScalerJobHPA) targetReplicas(ready int64, replicas int32) (desired int32) {
	if ready >= sh.hpaSpec.ScaleUp.Threshold && replicas < sh.hpaSpec.MaxReplicas {
		// calculate the number of expansions required, try to process them at once.
		max := ready / int64(sh.hpaSpec.ScaleUp.Amount)
		if ready%int64(sh.hpaSpec.ScaleUp.Amount) > 0 {
			max = max + 1
		}
		desired = replicas + int32(max)*sh.hpaSpec.ScaleUp.Amount
	} else if ready <= sh.hpaSpec.ScaleDown.Threshold && replicas > sh.hpaSpec.MinReplicas {
		desired = replicas - sh.hpaSpec.ScaleDown.Amount
	} else {
		desired = replicas
	}
	// check if desired is valid, desired should be between min and max
	if desired > sh.hpaSpec.MaxReplicas {
		desired = sh.hpaSpec.MaxReplicas
	}
	if desired < sh.hpaSpec.MinReplicas {
		desired = sh.hpaSpec.MinReplicas
	}
	return
}

func ScalerJobFactory(instance k8sq1comv1.HPAScaler, scaleClient scale.ScalesGetter, client client.Client) ScalerJob {
	return &ScalerJobHPA{
		id:              fmt.Sprintf("%v", instance.UID),
		hpaSpec:         instance.Spec,
		name:            instance.Name,
		namespace:       instance.Namespace,
		scaleClient:     scaleClient,
		client:          client,
		recommendations: []timestampedRecommendation{},
	}
}
