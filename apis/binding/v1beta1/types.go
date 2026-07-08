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

// BindingParameters define the desired state of a RabbitMQ Binding
type BindingParameters struct {
	// Source is the source exchange name
	// +kubebuilder:validation:Required
	Source string `json:"source"`

	// Destination is the destination queue or exchange name
	// +kubebuilder:validation:Required
	Destination string `json:"destination"`

	// DestinationType is the type of destination (queue or exchange)
	// +kubebuilder:validation:Enum=queue;exchange
	// +kubebuilder:default="queue"
	DestinationType string `json:"destinationType,omitempty"`

	// VHost is the name of the virtual host
	// +kubebuilder:validation:Required
	VHost string `json:"vhost"`

	// RoutingKey is the routing key for the binding
	// +kubebuilder:validation:Required
	RoutingKey string `json:"routingKey"`

	// Arguments is a map of additional arguments for the binding
	Arguments map[string]string `json:"arguments,omitempty"`
}

// BindingObservation reflects the observed state of a RabbitMQ Binding
type BindingObservation struct {
	Source          string            `json:"source,omitempty"`
	Destination     string            `json:"destination,omitempty"`
	DestinationType string            `json:"destinationType,omitempty"`
	VHost           string            `json:"vhost,omitempty"`
	RoutingKey      string            `json:"routingKey,omitempty"`
	Arguments       map[string]string `json:"arguments,omitempty"`
}

// A BindingSpec defines the desired state of a Binding.
type BindingSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              BindingParameters `json:"forProvider"`
}

// A BindingStatus represents the observed state of a Binding.
type BindingStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider             BindingObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Binding is a namespaced managed resource that represents a RabbitMQ Binding.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,rabbitmq}
// +kubebuilder:storageversion
type Binding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BindingSpec   `json:"spec"`
	Status BindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BindingList contains a list of Binding
type BindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Binding `json:"items"`
}
