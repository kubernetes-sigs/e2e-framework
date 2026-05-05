/*
Copyright 2026 The Kubernetes Authors.

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

// Package cel demonstrates the cel and cel/policy packages without a live
// cluster: every assertion is evaluated offline against a fixture object,
// the same way a unit test would exercise a ValidatingAdmissionPolicy
// without standing up an API server.
package cel

import (
	"strings"
	"testing"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	celpkg "sigs.k8s.io/e2e-framework/cel"
	"sigs.k8s.io/e2e-framework/cel/policy"
	"sigs.k8s.io/e2e-framework/klient/decoder"
)

// vapYAML is the same shape an operator would ship as a
// ValidatingAdmissionPolicy. policy.FromVAP turns it into a Policy that
// can be run offline against fixture objects.
const vapYAML = `
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: deployment-replicas
spec:
  matchConstraints:
    resourceRules:
    - apiGroups:   ["apps"]
      apiVersions: ["v1"]
      operations:  ["CREATE","UPDATE"]
      resources:   ["deployments"]
  validations:
    - expression: "object.spec.replicas >= 1"
      message: "replicas must be at least 1"
    - expression: "object.spec.replicas <= 100"
      message: "replicas must not exceed 100"
`

func TestEvaluatorAssert(t *testing.T) {
	ev, err := celpkg.NewEvaluator()
	if err != nil {
		t.Fatal(err)
	}
	dep := newDeployment("cel-demo", 2)
	if err := ev.Assert("object.spec.replicas >= 1", celpkg.ObjectBinding(dep)); err != nil {
		t.Fatal(err)
	}
}

func TestPolicyCheck(t *testing.T) {
	ev, err := celpkg.NewEvaluator()
	if err != nil {
		t.Fatal(err)
	}
	pol := policy.Policy{
		Name: "deployment-replicas",
		Validations: []policy.Validation{
			{Expression: "object.spec.replicas >= 1", Message: "replicas must be at least 1"},
			{Expression: "object.spec.replicas <= 100", Message: "replicas must not exceed 100"},
		},
	}
	if res := pol.Check(ev, newDeployment("cel-demo", 2)); !res.Passed() {
		t.Fatal(res.Err())
	}
	if res := pol.Check(ev, newDeployment("cel-demo", 0)); res.Passed() {
		t.Fatal("expected validation failure for replicas=0")
	}
}

func TestPolicyFromVAP(t *testing.T) {
	ev, err := celpkg.NewEvaluator()
	if err != nil {
		t.Fatal(err)
	}
	var vap admissionregistrationv1.ValidatingAdmissionPolicy
	if err := decoder.Decode(strings.NewReader(vapYAML), &vap); err != nil {
		t.Fatal(err)
	}
	pol := policy.FromVAP(&vap)
	if res := pol.Check(ev, newDeployment("cel-demo", 2)); !res.Passed() {
		t.Fatal(res.Err())
	}
}

func newDeployment(name string, replicas int32) *appsv1.Deployment {
	labels := map[string]string{"app": "cel-example"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "nginx", Image: "nginx"}},
				},
			},
		},
	}
}
