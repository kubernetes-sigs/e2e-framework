/*
Copyright 2022 The Kubernetes Authors.

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

package meta

import (
	"reflect"
	"testing"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestObjectMetaConstructor (t *testing.T) {
	tests := map[string]struct {
		constructor ObjectMetaConstructor
		expected metaV1.ObjectMeta
	}{
		"empty object":{
			constructor: ObjectMetaConstructor{obj:metaV1.ObjectMeta{}},
			expected: metaV1.ObjectMeta{},
		},
		"name only":{
			constructor: Object("simple-name"),
			expected: metaV1.ObjectMeta{Name: "simple-name"},
		},
		"name and namespace":{
			constructor: Object("simple-name").Namespace("my-namespace"),
			expected: metaV1.ObjectMeta{Name: "simple-name", Namespace: "my-namespace"},
		},
		"name and namespace and labels":{
			constructor: Object("simple-name").Namespace("my-namespace").Labels(map[string]string{"tier":"web"}),
			expected: metaV1.ObjectMeta{Name: "simple-name", Namespace: "my-namespace", Labels: map[string]string{"tier":"web"}},
		},
		"all fields":{
			constructor: Object("simple-name").Namespace("my-namespace").Labels(map[string]string{"tier":"web"}).ClusterName("test-cluster"),
			expected: metaV1.ObjectMeta{Name: "simple-name", Namespace: "my-namespace", Labels: map[string]string{"tier":"web"}, ClusterName: "test-cluster"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T){
			if !reflect.DeepEqual(test.constructor.Build(), test.expected) {
				t.Error("object not equal")
			}
		})
	}
}