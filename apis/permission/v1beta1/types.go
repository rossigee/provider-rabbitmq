/*
Copyright 2025 The Crossplane Authors.

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

	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
)

// PermissionParameters define the desired state of a RabbitMQ Permission
type PermissionParameters struct {
	// User is the name of the user
	// +kubebuilder:validation:Required
	User string `json:"user"`

	// VHost is the name of the virtual host
	// +kubebuilder:validation:Required
	VHost string `json:"vhost"`

	// Configure is the configure permission regex
	Configure string `json:"configure,omitempty"`

	// Write is the write permission regex
	Write string `json:"write,omitempty"`

	// Read is the read permission regex
	Read string `json:"read,omitempty"`
}

// PermissionObservation reflects the observed state of a RabbitMQ Permission
type PermissionObservation struct {
	User string `json:"user,omitempty"`
	VHost string `json:"vhost,omitempty"`
	Configure string `json:"configure,omitempty"`
	Write string `json:"write,omitempty"`
	Read string `json:"read,omitempty"`
}

// A PermissionSpec defines the desired state of a Permission.
type PermissionSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider             PermissionParameters `json:"forProvider"`
}

// A PermissionStatus represents the observed state of a Permission.
type PermissionStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider            PermissionObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Permission is a namespaced managed resource that represents a RabbitMQ Permission.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,rabbitmq}
// +kubebuilder:storageversion
type Permission struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PermissionSpec   `json:"spec"`
	Status PermissionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PermissionList contains a list of Permission
type PermissionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Permission `json:"items"`
}
