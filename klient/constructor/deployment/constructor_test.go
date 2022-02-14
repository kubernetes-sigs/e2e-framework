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

package deployment

import (
	"reflect"
	"testing"

	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/constructor/container"
	"sigs.k8s.io/e2e-framework/klient/constructor/meta"
	"sigs.k8s.io/e2e-framework/klient/constructor/pod"
)

func TestConstructor(t *testing.T) {
	tests := map[string]struct {
		constructor Constructor
		expected    appsV1.Deployment
	}{
		"simple": {
			constructor: Deployment(
				meta.Object("test-deployment").Namespace(meta.DefaultNamespace),
				Replicas(2),
				meta.MatchLabels(map[string]string{"server-type": "web"}),
				StrategyDefault,
				pod.Template(meta.ObjectMetaNone, pod.Spec(container.Name("server").Image("nginx").Commands("/start"))),
			),
			expected: appsV1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployment",
					Namespace: "default",
				},
				Spec: appsV1.DeploymentSpec{
					Replicas: Replicas(2),
					Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"server-type": "web"}},
					Strategy: appsV1.DeploymentStrategy{
						Type: appsV1.RecreateDeploymentStrategyType,
					},
					Template: coreV1.PodTemplateSpec{
						Spec: coreV1.PodSpec{
							Containers: []coreV1.Container{{
								Name:    "server",
								Image:   "nginx",
								Command: []string{"/start"},
							}},
						},
					},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			if !reflect.DeepEqual(test.constructor.Build(), test.expected) {
				t.Error("object not equal")
			}
		})
	}
}
