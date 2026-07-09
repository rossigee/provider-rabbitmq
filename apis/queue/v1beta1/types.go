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
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// QueueParameters define the desired state of a RabbitMQ Queue
type QueueParameters struct {
	// Name is the name of the queue
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// VHost is the name of the virtual host to create the queue in
	// +kubebuilder:validation:Required
	VHost string `json:"vhost"`

	// Durable survives broker restart
	// +kubebuilder:default=true
	Durable bool `json:"durable,omitempty"`

	// AutoDelete deletes the queue when it has no consumers
	AutoDelete bool `json:"autoDelete,omitempty"`

	// Exclusive is only accessible to the connection that declared it
	Exclusive bool `json:"exclusive,omitempty"`

	// Arguments is a map of additional arguments for the queue
	// Note: x-message-ttl should be passed as string (e.g., "3600000") for CRD compatibility
	Arguments map[string]string `json:"arguments,omitempty"`

	// MessageTTL sets the TTL for messages in the queue (in milliseconds)
	MessageTTL int `json:"messageTTL,omitempty"`

	// Expires sets the time (in milliseconds) after which a queue can be deleted if no consumers
	Expires int `json:"expires,omitempty"`

	// MaxLength sets the maximum number of messages in the queue
	MaxLength int `json:"maxLength,omitempty"`

	// OverflowBehavior sets the behavior when the queue reaches its max length
	// +kubebuilder:validation:Enum=drop-head;reject-publish;reject-publish-dlx
	OverflowBehavior string `json:"overflowBehavior,omitempty"`
}

// QueueObservation reflects the observed state of a RabbitMQ Queue
type QueueObservation struct {
	Name             string            `json:"name,omitempty"`
	VHost            string            `json:"vhost,omitempty"`
	Durable          bool              `json:"durable,omitempty"`
	AutoDelete       bool              `json:"autoDelete,omitempty"`
	Exclusive        bool              `json:"exclusive,omitempty"`
	Arguments        map[string]string `json:"arguments,omitempty"`
	MessageTTL       int               `json:"messageTTL,omitempty"`
	Expires          int               `json:"expires,omitempty"`
	MaxLength        int               `json:"maxLength,omitempty"`
	OverflowBehavior string            `json:"overflowBehavior,omitempty"`
	Messages         int               `json:"messages,omitempty"`
	Consumers        int               `json:"consumers,omitempty"`
}

// A QueueSpec defines the desired state of a Queue.
type QueueSpec struct {
	xpv1.ManagedResourceSpec `json:",inline"`
	ForProvider              QueueParameters `json:"forProvider"`
}

// A QueueStatus represents the observed state of a Queue.
type QueueStatus struct {
	xpv1.ConditionedStatus `json:",inline"`
	AtProvider             QueueObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Queue is a namespaced managed resource that represents a RabbitMQ Queue.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,rabbitmq}
// +kubebuilder:storageversion
type Queue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QueueSpec   `json:"spec"`
	Status QueueStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QueueList contains a list of Queue
type QueueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Queue `json:"items"`
}
