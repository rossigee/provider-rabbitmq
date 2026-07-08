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
	"github.com/crossplane/crossplane/apis/v2/core/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExchangeParameters define the desired state of a RabbitMQ Exchange
type ExchangeParameters struct {
	// Name is the name of the exchange
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// VHost is the name of the virtual host to create the exchange in
	// +kubebuilder:validation:Required
	VHost string `json:"vhost"`

	// Type is the exchange type
	// +kubebuilder:validation:Enum=direct;topic;fanout;headers;x-delayed-message;x-durable-delayed-message
	// +kubebuilder:default="topic"
	Type string `json:"type,omitempty"`

	// AutoDelete deletes the exchange when it has no subscribers
	AutoDelete bool `json:"autoDelete,omitempty"`

	// Durable survives broker restart
	// +kubebuilder:default=true
	Durable bool `json:"durable,omitempty"`

	// Internal is true if the exchange is internal (cannot be directly published to)
	Internal bool `json:"internal,omitempty"`

	// Arguments is a map of additional arguments for the exchange
	Arguments map[string]string `json:"arguments,omitempty"`
}

// ExchangeObservation reflects the observed state of a RabbitMQ Exchange
type ExchangeObservation struct {
	Name       string            `json:"name,omitempty"`
	VHost      string            `json:"vhost,omitempty"`
	Type       string            `json:"type,omitempty"`
	AutoDelete bool              `json:"autoDelete,omitempty"`
	Durable    bool              `json:"durable,omitempty"`
	Internal   bool              `json:"internal,omitempty"`
	Arguments  map[string]string `json:"arguments,omitempty"`
}

// A ExchangeSpec defines the desired state of a Exchange.
type ExchangeSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              ExchangeParameters `json:"forProvider"`
}

// A ExchangeStatus represents the observed state of a Exchange.
type ExchangeStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider             ExchangeObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Exchange is a namespaced managed resource that represents a RabbitMQ Exchange.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,rabbitmq}
// +kubebuilder:storageversion
type Exchange struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExchangeSpec   `json:"spec"`
	Status ExchangeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExchangeList contains a list of Exchange
type ExchangeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Exchange `json:"items"`
}
