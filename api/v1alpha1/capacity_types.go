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
	apiv1alpha1 "github.com/llm-inferno/api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CapacitySpec defines the desired state of Capacity.
type CapacitySpec apiv1alpha1.CapacityData

// CapacityStatus defines the observed state of Capacity.
type CapacityStatus struct {
	Active bool `json:"active"` // processed by the optimizer
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Capacity is the Schema for the capacities API.
type Capacity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CapacitySpec   `json:"spec,omitempty"`
	Status CapacityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CapacityList contains a list of Capacity.
type CapacityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Capacity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Capacity{}, &CapacityList{})
}
