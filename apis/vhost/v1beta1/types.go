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

// VhostParameters define the desired state of a RabbitMQ Vhost
type VhostParameters struct {
	// Name is the name of the vhost to create
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Description is an optional description for the vhost
	// +kubebuilder:validation:MaxLength=2000
	Description string `json:"description,omitempty"`

	// Tags is an optional set of tags for the vhost
	Tags []string `json:"tags,omitempty"`
}

// VhostObservation reflects the observed state of a RabbitMQ Vhost
type VhostObservation struct {
	// Name is the vhost name
	Name string `json:"name,omitempty"`

	// Description is the vhost description
	Description string `json:"description,omitempty"`

	// Tags associated with the vhost
	Tags []string `json:"tags,omitempty"`

	// TracerPort is the port used for tracing
	TracerPort int `json:"tracerPort,omitempty"`
}

// A VhostSpec defines the desired state of a Vhost.
type VhostSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider             VhostParameters `json:"forProvider"`
}

// A VhostStatus represents the observed state of a Vhost.
type VhostStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider            VhostObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Vhost is a managed resource that represents a RabbitMQ Vhost.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,rabbitmq}
// +kubebuilder:storageversion
type Vhost struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VhostSpec   `json:"spec"`
	Status VhostStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VhostList contains a list of Vhost
type VhostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Vhost `json:"items"`
}
