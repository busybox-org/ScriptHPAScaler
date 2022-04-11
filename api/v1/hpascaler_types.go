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

package v1

import (
	"encoding/json"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Status string

const (
	Succeed   Status = "Succeed"
	Failed    Status = "Failed"
	Submitted Status = "Submitted"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HPAScalerSpec defines the desired state of HPAScaler
type HPAScalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ScaleTargetRef               ScaleTargetRef `json:"scaleTargetRef"`
	Freq                         string         `json:"freq"`
	Plugin                       Plugin         `json:"plugin"`
	MaxReplicas                  int32          `json:"maxReplicas,omitempty"`
	MinReplicas                  int32          `json:"minReplicas,omitempty"`
	ScaleUp                      Scale          `json:"scaleUp,omitempty"`
	ScaleDown                    Scale          `json:"scaleDown,omitempty"`
	DownscaleStabilisationWindow string         `json:"downscaleStabilisationWindow,omitempty" default:"3m"`
}

func (hs *HPAScalerSpec) ToString() string {
	bs, err := json.Marshal(hs)
	if err != nil {
		return ""
	}
	return string(bs)
}

type ScaleTargetRef struct {
	Kind string `json:"kind"` // "Deployment" or "StatefulSet" Or "Node"
	Name string `json:"name"`
}

type Plugin struct {
	Type   string `json:"type"`
	Url    string `json:"url"`
	Config Config `json:"config,omitempty"`
}

type Config map[string]string

type Scale struct {
	Threshold int64 `json:"threshold,omitempty"` // 触发临界值
	Amount    int32 `json:"amount,omitempty"`    // 增加或减少的数量
}

// HPAScalerStatus defines the observed state of HPAScaler
type HPAScalerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// +kubebuilder:validation:Required
	Condition Condition `json:"condition,omitempty"`
}

type Condition struct {
	UID             string      `json:"uid"`
	Status          Status      `json:"status"`
	LastProbeTime   metav1.Time `json:"lastProbeTime"`
	DesiredReplicas int32       `json:"desiredReplicas"`
	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HPAScaler is the Schema for the hpascalers API
type HPAScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HPAScalerSpec   `json:"spec,omitempty"`
	Status HPAScalerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HPAScalerList contains a list of HPAScaler
type HPAScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HPAScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HPAScaler{}, &HPAScalerList{})
}
