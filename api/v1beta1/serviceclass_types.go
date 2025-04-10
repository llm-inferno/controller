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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceClassSpec defines the desired state of ServiceClass.
type ServiceClassSpec struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"` // service class name

	// +kubebuilder:default=0
	// +kubebuilder:validation:Minimum=0
	Priority int `json:"priority,omitempty"` // (non-negative) priority (lower value is higher priority)

	Data []ServiceClassModelData `json:"data,omitempty"` // model SLO data
}

// Specifications of SLO data for a model
type ServiceClassModelData struct {
	// +kubebuilder:validation:MinLength=1
	Model string `json:"model"` // model name

	SLO_ITL float32 `json:"slo-itl,omitempty"` // inter-token latency (msec)
	SLO_TTW float32 `json:"slo-ttw,omitempty"` // request waiting time (msec)
	SLO_TPS float32 `json:"slo-tps,omitempty"` // throughput (tokens/sec)
}

// Specifications of SLO data for a combination of a service class and a model
type ServiceClassDataItem struct {
	Name     string `json:"name"`               // service class name
	Priority int    `json:"priority,omitempty"` // (non-negative) priority (lower value is higher priority)
	ServiceClassModelData
}

// ServiceClassStatus defines the observed state of ServiceClass.
type ServiceClassStatus struct {
	Active bool `json:"active"` // processed by the optimizer
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ServiceClass is the Schema for the serviceclasses API.
type ServiceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceClassSpec   `json:"spec,omitempty"`
	Status ServiceClassStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceClassList contains a list of ServiceClass.
type ServiceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceClass{}, &ServiceClassList{})
}
