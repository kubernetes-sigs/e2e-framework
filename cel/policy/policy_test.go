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

package policy

import (
	"strings"
	"testing"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/cel"
)

func dep(replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "default"},
		Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
	}
}

var replicasPolicy = Policy{
	Name: "deployment-replicas",
	Validations: []Validation{
		{Expression: "object.spec.replicas >= 1", Message: "replicas must be at least 1"},
		{Expression: "object.spec.replicas <= 100", Message: "replicas must not exceed 100"},
	},
}

func TestPolicy_admits(t *testing.T) {
	ev := newEv(t)
	res := replicasPolicy.Check(ev, dep(3))
	if !res.Passed() {
		t.Fatalf("expected pass, got failures: %v", res.Err())
	}
	if res.Err() != nil {
		t.Fatalf("passed result should have nil Err, got %v", res.Err())
	}
}

func TestPolicy_rejects_belowMin(t *testing.T) {
	ev := newEv(t)
	res := replicasPolicy.Check(ev, dep(0))
	if res.Passed() {
		t.Fatal("expected failure")
	}
	if len(res.Failures) != 1 {
		t.Fatalf("want 1 failure, got %d", len(res.Failures))
	}
	if !strings.Contains(res.Err().Error(), "at least 1") {
		t.Fatalf("want min-replicas message, got %v", res.Err())
	}
}

func TestPolicy_accumulatesAllFailures(t *testing.T) {
	multi := Policy{
		Name: "multi",
		Validations: []Validation{
			{Expression: "object.spec.replicas < 0", Message: "must be negative"},
			{Expression: "object.spec.replicas > 1000", Message: "must exceed 1000"},
		},
	}
	ev := newEv(t)
	res := multi.Check(ev, dep(5))
	if len(res.Failures) != 2 {
		t.Fatalf("want 2 failures, got %d", len(res.Failures))
	}
}

func TestPolicy_emptyMessageFallsBackToExpression(t *testing.T) {
	p := Policy{
		Name: "bare",
		Validations: []Validation{
			{Expression: "object.spec.replicas < 0"},
		},
	}
	ev := newEv(t)
	res := p.Check(ev, dep(5))
	if !strings.Contains(res.Err().Error(), "object.spec.replicas < 0") {
		t.Fatalf("want expression in message, got %v", res.Err())
	}
}

func TestFromVAP_preservesValidations(t *testing.T) {
	reasonInvalid := metav1.StatusReasonInvalid
	vap := &admissionregistrationv1.ValidatingAdmissionPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "require-replicas"},
		Spec: admissionregistrationv1.ValidatingAdmissionPolicySpec{
			Validations: []admissionregistrationv1.Validation{
				{
					Expression: "object.spec.replicas >= 1",
					Message:    "replicas >= 1",
					Reason:     &reasonInvalid,
				},
				{
					Expression: "object.spec.replicas <= 100",
					Message:    "replicas <= 100",
				},
			},
		},
	}
	p := FromVAP(vap)
	if p.Name != "require-replicas" {
		t.Errorf("Name = %q, want require-replicas", p.Name)
	}
	if len(p.Validations) != 2 {
		t.Fatalf("want 2 validations, got %d", len(p.Validations))
	}
	if p.Validations[0].Reason != metav1.StatusReasonInvalid {
		t.Errorf("Reason not preserved: got %q", p.Validations[0].Reason)
	}

	// Round-trip: run the converted Policy and confirm it admits a valid obj.
	ev := newEv(t)
	if !p.Check(ev, dep(3)).Passed() {
		t.Fatal("converted policy should admit dep(3)")
	}
	if p.Check(ev, dep(0)).Passed() {
		t.Fatal("converted policy should reject dep(0)")
	}
}

func TestFromVAP_nilReturnsEmpty(t *testing.T) {
	p := FromVAP(nil)
	if p.Name != "" || len(p.Validations) != 0 {
		t.Fatalf("nil VAP should produce empty Policy, got %+v", p)
	}
}

func newEv(t *testing.T) *cel.Evaluator {
	t.Helper()
	ev, err := cel.NewEvaluator()
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	return ev
}
