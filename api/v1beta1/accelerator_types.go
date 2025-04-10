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
	apiv1beta1 "github.ibm.com/inferno/api/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AcceleratorSpec defines the desired state of Accelerator.
type AcceleratorSpec apiv1beta1.AcceleratorSpec

// AcceleratorStatus defines the observed state of Accelerator.
type AcceleratorStatus struct {
	Active bool `json:"active"` // processed by the optimizer
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Accelerator is the Schema for the accelerators API.
type Accelerator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AcceleratorSpec   `json:"spec,omitempty"`
	Status AcceleratorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AcceleratorList contains a list of Accelerator.
type AcceleratorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Accelerator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Accelerator{}, &AcceleratorList{})
}
