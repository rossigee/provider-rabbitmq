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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"reflect"
)

// Package type metadata.
const (
	Group   = "rabbitmq.provider.crossplane.io"
	Version = "v1beta1"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
)

// Vhost type metadata.
var (
	VhostKind             = reflect.TypeOf(Vhost{}).Name()
	VhostGroupKind        = schema.GroupKind{Group: Group, Kind: VhostKind}
	VhostKindAPIVersion   = VhostKind + "." + SchemeGroupVersion.String()
	VhostGroupVersionKind = SchemeGroupVersion.WithKind(VhostKind)
)

// AddToScheme adds all types of this group into the given scheme.
func addKnownTypes(s *runtime.Scheme) error {
	return nil
}

func AddToScheme(s *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(s)
}
