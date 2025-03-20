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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ScaleState string

const (
	ScaleStateSuccess ScaleState = "success"
	ScaleStateFailure ScaleState = "failure"
	ScaleStatePending ScaleState = "pending"
	ScaleStateUnknown ScaleState = "unknown"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScriptHPAScalerSpec defines the desired state of ScriptHPAScaler.
type ScriptHPAScalerSpec struct {
	StabilisationWindow string         `json:"stabilisationWindow,omitempty" default:"3m"`
	MaxReplicas         int32          `json:"maxReplicas"`
	MinReplicas         int32          `json:"minReplicas"`
	ScaleTargetRef      ScaleTargetRef `json:"scaleTargetRef"`
	Script              string         `json:"script"`
}

type ScaleTargetRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
}

// ScriptHPAScalerStatus defines the observed state of ScriptHPAScaler.
type ScriptHPAScalerStatus struct {
	State           ScaleState  `json:"state"`
	LastProbeTime   metav1.Time `json:"lastProbeTime"`
	DesiredReplicas int32       `json:"desiredReplicas,omitempty"`
	Message         string      `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ScriptHPAScaler is the Schema for the scripthpascalers API.
type ScriptHPAScaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScriptHPAScalerSpec   `json:"spec"`
	Status ScriptHPAScalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScriptHPAScalerList contains a list of ScriptHPAScaler.
type ScriptHPAScalerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScriptHPAScaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScriptHPAScaler{}, &ScriptHPAScalerList{})
}
