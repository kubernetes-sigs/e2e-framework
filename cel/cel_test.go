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

package cel

import (
	"strings"
	"sync"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deployment(replicas, ready int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
			Labels:    map[string]string{"app": "demo"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
		},
		Status: appsv1.DeploymentStatus{
			ReadyReplicas: ready,
		},
	}
}

func TestNewEvaluator_defaults(t *testing.T) {
	ev, err := NewEvaluator()
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	if ev.opts.env != EnvAdmission {
		t.Errorf("default env = %v, want EnvAdmission", ev.opts.env)
	}
	if ev.opts.costLimit != DefaultCostLimit {
		t.Errorf("default costLimit = %d, want %d", ev.opts.costLimit, DefaultCostLimit)
	}
}

func TestAssert_pass(t *testing.T) {
	ev := newEv(t)
	if err := ev.Assert("object.spec.replicas == 3",
		ObjectBinding(deployment(3, 3))); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
}

func TestAssert_fail(t *testing.T) {
	ev := newEv(t)
	err := ev.Assert("object.status.readyReplicas == object.spec.replicas",
		ObjectBinding(deployment(3, 1)))
	if err == nil {
		t.Fatal("expected failure, got nil")
	}
	if !strings.Contains(err.Error(), "assertion failed") {
		t.Fatalf("want assertion failure, got %v", err)
	}
}

func TestAssert_nonBoolResultIsError(t *testing.T) {
	ev := newEv(t)
	err := ev.Assert("1 + 1", ObjectBinding(deployment(1, 1)))
	if err == nil || !strings.Contains(err.Error(), "want bool") {
		t.Fatalf("expected non-bool error, got %v", err)
	}
}

func TestAssert_compileError(t *testing.T) {
	ev := newEv(t)
	err := ev.Assert("object.spec.replicas >", ObjectBinding(deployment(1, 1)))
	if err == nil || !strings.Contains(err.Error(), "compile") {
		t.Fatalf("expected compile error, got %v", err)
	}
}

func TestAssert_unknownVariableIsCompileError(t *testing.T) {
	ev := newEv(t)
	// Admission env exposes `object` but not `self`.
	err := ev.Assert("self.spec.replicas > 0", ObjectBinding(deployment(1, 1)))
	if err == nil || !strings.Contains(err.Error(), "compile") {
		t.Fatalf("expected compile error for unknown var, got %v", err)
	}
}

func TestEval_returnsRawValue(t *testing.T) {
	ev := newEv(t)
	out, err := ev.Eval("object.spec.replicas", ObjectBinding(deployment(5, 5)))
	if err != nil {
		t.Fatal(err)
	}
	got := out.Value()
	// ref.Val returns int64 for integer fields under DynType.
	if got != int64(5) {
		t.Errorf("got %v (%T), want int64(5)", got, got)
	}
}

func TestProgramCache_reusedAcrossEval(t *testing.T) {
	ev := newEv(t)
	expr := "object.spec.replicas == 3"
	dep := ObjectBinding(deployment(3, 3))

	if err := ev.Assert(expr, dep); err != nil {
		t.Fatal(err)
	}
	if _, cached := ev.cache.Load(expr); !cached {
		t.Fatal("program not cached after first Assert")
	}
	// Second call must hit the cache without recompiling.
	if err := ev.Assert(expr, dep); err != nil {
		t.Fatal(err)
	}
}

func TestProgramCache_concurrent(t *testing.T) {
	ev := newEv(t)
	expr := "object.spec.replicas >= 1"
	dep := ObjectBinding(deployment(3, 3))

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := ev.Assert(expr, dep); err != nil {
				t.Errorf("concurrent Assert: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestWithEnvironment_CRD(t *testing.T) {
	ev, err := NewEvaluator(WithEnvironment(EnvCRD))
	if err != nil {
		t.Fatal(err)
	}
	if err := ev.Assert(`self.metadata.name == "demo"`,
		SelfBinding(deployment(1, 1))); err != nil {
		t.Fatalf("CRD-env Assert failed: %v", err)
	}
	// `object` must not resolve in the CRD environment.
	err = ev.Assert("object.spec.replicas == 1", ObjectBinding(deployment(1, 1)))
	if err == nil || !strings.Contains(err.Error(), "compile") {
		t.Fatalf("expected compile error for admission var in CRD env, got %v", err)
	}
}

func TestWithLibraries_narrowsSet(t *testing.T) {
	// Only wire Quantity — Regex must not be available.
	ev, err := NewEvaluator(WithLibraries(LibQuantity))
	if err != nil {
		t.Fatal(err)
	}
	// Quantity works.
	if err := ev.Assert(`quantity("100Mi").isGreaterThan(quantity("50Mi"))`, nil); err != nil {
		t.Fatalf("quantity assertion failed: %v", err)
	}
	// Regex-based call should compile-fail.
	err = ev.Assert(`"abc".find("a.")  != ""`, nil)
	if err == nil || !strings.Contains(err.Error(), "compile") {
		t.Fatalf("expected regex to be unavailable, got %v", err)
	}
}

func TestDefaultLibraries_wireStandardK8sSet(t *testing.T) {
	ev := newEv(t)
	cases := []string{
		`quantity("100Mi").isGreaterThan(quantity("50Mi"))`,
		`ip("10.0.0.1").family() == 4`,
		`cidr("10.0.0.0/8").containsIP("10.1.2.3")`,
		`"deployment-demo".find("[a-z]+")  == "deployment"`,
		`semver("1.2.3").isLessThan(semver("2.0.0"))`,
	}
	for _, expr := range cases {
		if err := ev.Assert(expr, nil); err != nil {
			t.Errorf("expected library call %q to pass, got %v", expr, err)
		}
	}
}

func TestBind_laterOverridesEarlier(t *testing.T) {
	a := Bindings{"x": 1, "y": 2}
	b := Bindings{"y": 3, "z": 4}
	merged := Bind(a, b)
	if merged["x"] != 1 || merged["y"] != 3 || merged["z"] != 4 {
		t.Fatalf("Bind merge wrong: %v", merged)
	}
}

func TestObjectBinding_nilBecomesNil(t *testing.T) {
	b := ObjectBinding(nil)
	if b["object"] != nil {
		t.Fatalf("nil object should bind as nil, got %v", b["object"])
	}
}

func TestOldObjectBinding_boundKey(t *testing.T) {
	b := OldObjectBinding(deployment(1, 1))
	if _, ok := b["oldObject"]; !ok {
		t.Fatal("OldObjectBinding should populate oldObject")
	}
}

func TestRequestBinding_surfacesOperation(t *testing.T) {
	ev := newEv(t)
	b := Bind(
		ObjectBinding(deployment(3, 3)),
		RequestBinding(&AdmissionRequest{Operation: "CREATE", Namespace: "demo"}),
	)
	if err := ev.Assert(`request.operation == "CREATE" && request.namespace == "demo"`, b); err != nil {
		t.Fatalf("request binding: %v", err)
	}
}

func TestNamespaceBinding(t *testing.T) {
	ev := newEv(t)
	ns := &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Labels: map[string]string{"tier": "gold"}},
	}
	b := Bind(
		ObjectBinding(deployment(1, 1)),
		NamespaceBinding(ns),
	)
	if err := ev.Assert(`namespaceObject.metadata.labels["tier"] == "gold"`, b); err != nil {
		t.Fatalf("namespace binding: %v", err)
	}
}

func newEv(t *testing.T) *Evaluator {
	t.Helper()
	ev, err := NewEvaluator()
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	return ev
}
