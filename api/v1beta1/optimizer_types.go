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

package v1beta1

import (
	apiv1beta1 "github.com/llm-inferno/api/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OptimizerSpec defines the desired state of Optimizer.
type OptimizerSpec struct {
	Optimize bool          `json:"optimize"` // request to invoke optimizer
	Data     OptimizerData `json:"data"`     // parameter data for optimizer
}

type OptimizerData apiv1beta1.OptimizerData
type AllocationSolution apiv1beta1.AllocationSolution

// OptimizerStatus defines the observed state of Optimizer.
type OptimizerStatus struct {
	Done bool `json:"done"` // processed by the optimizer
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Optimizer is the Schema for the optimizers API.
type Optimizer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OptimizerSpec   `json:"spec,omitempty"`
	Status OptimizerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OptimizerList contains a list of Optimizer.
type OptimizerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Optimizer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Optimizer{}, &OptimizerList{})
}
