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

package clients

import (
	"bytes"
	"encoding/json"

	queuev1beta1 "github.com/rossigee/provider-rabbitmq/apis/queue/v1beta1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// ArgsMapsEqual reports whether two RabbitMQ "arguments" maps are equivalent.
// Values are stored as raw JSON (apiextensionsv1.JSON), so equivalence is
// byte-wise on the encoded value. Keys are compared rather than marshalled, so
// key order is irrelevant.
func ArgsMapsEqual(a, b map[string]apiextensionsv1.JSON) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if !bytes.Equal(av.Raw, bv.Raw) {
			return false
		}
	}
	return true
}

// jsonArgsToInterface converts raw-JSON arguments into Go values suitable for
// marshalling into a Management API request body.
func jsonArgsToInterface(args map[string]apiextensionsv1.JSON) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range args {
		var value interface{}
		if err := json.Unmarshal(v.Raw, &value); err != nil {
			value = string(v.Raw)
		}
		out[k] = value
	}
	return out
}

// QueueArgumentsJSON returns the RabbitMQ queue "arguments" object derived
// from the typed convenience fields and the free-form Arguments map, encoded
// as raw JSON. Typed fields take precedence over matching entries in the map.
func QueueArgumentsJSON(spec *queuev1beta1.QueueParameters) map[string]apiextensionsv1.JSON {
	args := map[string]apiextensionsv1.JSON{}
	for k, v := range spec.Arguments {
		args[k] = v
	}
	if spec.MessageTTL != 0 {
		args["x-message-ttl"] = jsonValue(spec.MessageTTL)
	}
	if spec.Expires != 0 {
		args["x-expires"] = jsonValue(spec.Expires)
	}
	if spec.MaxLength != 0 {
		args["x-max-length"] = jsonValue(spec.MaxLength)
	}
	if spec.OverflowBehavior != "" {
		args["x-overflow"] = jsonValue(spec.OverflowBehavior)
	}
	return args
}

func jsonValue(v interface{}) apiextensionsv1.JSON {
	raw, _ := json.Marshal(v)
	return apiextensionsv1.JSON{Raw: raw}
}

// StringSetEqual reports whether two string slices contain the same elements
// regardless of order. Used for tag comparisons, where the broker may return
// tags in a different order than the spec declared them.
func StringSetEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, s := range a {
		m[s]++
	}
	for _, s := range b {
		if m[s] == 0 {
			return false
		}
		m[s]--
	}
	return true
}
