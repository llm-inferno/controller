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

// ModelSpec defines the desired state of Model.
type ModelSpec struct {
	// +kubebuilder:validation:MinLength=1
	Name string                `json:"name"`           // name of model
	Data []AcceleratorPerfData `json:"data,omitempty"` // performance data
}

// Specifications for accelerator data
type AcceleratorPerfData struct {
	Acc          string  `json:"acc"`          // accelerator name
	AccCount     int     `json:"accCount"`     // number of accelerator units used by model
	Alpha        float32 `json:"alpha"`        // alpha parameter of ITL
	Beta         float32 `json:"beta"`         // beta parameter of ITL
	MaxBatchSize int     `json:"maxBatchSize"` // max batch size based on average number of tokens per request
	AtTokens     int     `json:"atTokens"`     // average number of tokens per request assumed in max batch size calculation
}

// Specifications for a combination of a model and accelerator data
type ModelAcceleratorPerfData struct {
	Name string `json:"name"` // model name
	AcceleratorPerfData
}

// ModelStatus defines the observed state of Model.
type ModelStatus struct {
	Active bool `json:"active"` // processed by the optimizer
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Model is the Schema for the models API.
type Model struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSpec   `json:"spec,omitempty"`
	Status ModelStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ModelList contains a list of Model.
type ModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Model `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Model{}, &ModelList{})
}
