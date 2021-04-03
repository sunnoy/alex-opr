/*
Copyright 2021 lirui.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MacBookSpec defines the desired state of MacBook
type MacBookSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// DisPlay is an example field of MacBook. Edit macbook_types.go to remove/update
	// todo code 添加spec的字段
	DisPlay string `json:"display,omitempty"`
}

// MacBookStatus defines the observed state of MacBook
type MacBookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// todo code 添加status的字段
	Mod string `json:"mod,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MacBook is the Schema for the macbooks API
type MacBook struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MacBookSpec   `json:"spec,omitempty"`
	Status MacBookStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MacBookList contains a list of MacBook
type MacBookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MacBook `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MacBook{}, &MacBookList{})
}
