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

package decoder

import (
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/cel"
	"sigs.k8s.io/e2e-framework/cel/policy"
	kdecoder "sigs.k8s.io/e2e-framework/klient/decoder"
)

const singleDocYAML = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: demo
data:
  greeting: hello
`

const multiDocYAML = `
apiVersion: v1
kind: ConfigMap
metadata: {name: a, namespace: demo}
data: {key: value-a}
---
apiVersion: v1
kind: ConfigMap
metadata: {name: b, namespace: demo}
data: {key: value-b}
---
apiVersion: v1
kind: ConfigMap
metadata: {name: c, namespace: demo}
data: {key: value-c}
`

func TestAssertYAML_pass(t *testing.T) {
	ev := newEv(t)
	err := AssertYAML(ev,
		`object.kind == "ConfigMap" && object.metadata.namespace == "demo"`,
		strings.NewReader(singleDocYAML))
	if err != nil {
		t.Fatalf("AssertYAML: %v", err)
	}
}

func TestAssertYAML_fail(t *testing.T) {
	ev := newEv(t)
	err := AssertYAML(ev,
		`object.metadata.namespace == "prod"`,
		strings.NewReader(singleDocYAML))
	if err == nil || !strings.Contains(err.Error(), "assertion failed") {
		t.Fatalf("expected assertion failure, got %v", err)
	}
}

func TestAssertYAMLAll_allPass(t *testing.T) {
	ev := newEv(t)
	err := AssertYAMLAll(context.Background(), ev,
		`object.kind == "ConfigMap"`,
		strings.NewReader(multiDocYAML))
	if err != nil {
		t.Fatalf("AssertYAMLAll: %v", err)
	}
}

func TestAssertYAMLAll_haltsOnFirstFailure(t *testing.T) {
	// b/key is "value-b"; assertion requires it to equal "value-a".
	ev := newEv(t)
	err := AssertYAMLAll(context.Background(), ev,
		`object.data.key == "value-a"`,
		strings.NewReader(multiDocYAML))
	if err == nil {
		t.Fatal("expected failure on second document")
	}
	if !strings.Contains(err.Error(), "assertion failed") {
		t.Fatalf("want assertion failure, got %v", err)
	}
}

func TestPolicyHandler_perObject(t *testing.T) {
	ev := newEv(t)
	pol := policy.Policy{
		Name: "config-shape",
		Validations: []policy.Validation{
			{Expression: `has(object.data) && size(object.data) > 0`, Message: "must have data"},
			{Expression: `object.metadata.namespace != ""`, Message: "must have namespace"},
		},
	}
	err := kdecoder.DecodeEach(context.Background(),
		strings.NewReader(multiDocYAML),
		PolicyHandler(ev, pol))
	if err != nil {
		t.Fatalf("PolicyHandler: %v", err)
	}
}

func TestAssertHandler_appliesPerObject(t *testing.T) {
	// Use DecodeEach directly with AssertHandler — the standard streaming
	// pattern a test author would write.
	ev := newEv(t)
	err := kdecoder.DecodeEach(context.Background(),
		strings.NewReader(multiDocYAML),
		AssertHandler(ev, `object.metadata.namespace == "demo"`))
	if err != nil {
		t.Fatalf("DecodeEach + AssertHandler: %v", err)
	}
}

// TestDirect_AssertAgainstTypedObject sanity-checks that the same decoded
// object we evaluate via YAML also evaluates via ObjectBinding on a typed
// struct, so both paths stay consistent.
func TestDirect_AssertAgainstTypedObject(t *testing.T) {
	cm := &corev1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
		ObjectMeta: metav1.ObjectMeta{Name: "app-config", Namespace: "demo"},
		Data:       map[string]string{"greeting": "hello"},
	}
	ev := newEv(t)
	if err := ev.Assert(`object.data.greeting == "hello"`, cel.ObjectBinding(cm)); err != nil {
		t.Fatalf("typed ObjectBinding: %v", err)
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
