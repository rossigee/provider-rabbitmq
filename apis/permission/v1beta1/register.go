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
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "rabbitmq.provider.crossplane.io"
	Version = "v1beta1"
)

var (
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Permission type metadata.
var (
	PermissionKind             = reflect.TypeOf(Permission{}).Name()
	PermissionGroupKind        = schema.GroupKind{Group: Group, Kind: PermissionKind}
	PermissionKindAPIVersion   = PermissionKind + "." + SchemeGroupVersion.String()
	PermissionGroupVersionKind = SchemeGroupVersion.WithKind(PermissionKind)
)

func init() {
	SchemeBuilder.Register(&Permission{}, &PermissionList{})
}

// AddToScheme adds all types of this group into the given scheme.
func AddToScheme(s *runtime.Scheme) error {
	return SchemeBuilder.AddToScheme(s)
}
