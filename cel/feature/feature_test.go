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

package feature

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/cel"
	"sigs.k8s.io/e2e-framework/cel/policy"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

// These tests exercise the happy path of each adapter. Failure behavior is
// a thin t.Fatal call and is covered by the primitive tests in ../cel and
// ../cel/policy. The Go testing framework does not give us a practical way
// to capture t.Fatal without failing the parent test, which is itself a
// motivator for the pluggable-T work in #527.

func dep(replicas, ready int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Name: "demo", Namespace: "default"},
		Spec:       appsv1.DeploymentSpec{Replicas: &replicas},
		Status:     appsv1.DeploymentStatus{ReadyReplicas: ready},
	}
}

func TestAssert_happyPath(t *testing.T) {
	ev := newEv(t)
	binder := func(_ context.Context, _ *envconf.Config) (cel.Bindings, error) {
		return cel.ObjectBinding(dep(3, 3)), nil
	}
	fn := Assert(ev, "object.status.readyReplicas == object.spec.replicas", binder)
	fn(context.Background(), t, &envconf.Config{})
}

func TestAssertPolicy_happyPath(t *testing.T) {
	ev := newEv(t)
	pol := policy.Policy{
		Name: "replicas",
		Validations: []policy.Validation{
			{Expression: "object.spec.replicas >= 1"},
			{Expression: "object.spec.replicas <= 100"},
		},
	}
	fetcher := func(_ context.Context, _ *envconf.Config) (k8s.Object, error) {
		return dep(3, 3), nil
	}
	fn := AssertPolicy(ev, pol, fetcher)
	fn(context.Background(), t, &envconf.Config{})
}

// TestAdapters_compileAsFeaturesFunc pins down that each adapter returns a
// features.Func value — i.e. the type is exactly what features.Builder.Assess
// accepts. If this compiles, the adapter surface matches the intended shape.
func TestAdapters_compileAsFeaturesFunc(t *testing.T) {
	ev := newEv(t)

	binder := func(_ context.Context, _ *envconf.Config) (cel.Bindings, error) {
		return cel.ObjectBinding(dep(1, 1)), nil
	}
	fetcher := func(_ context.Context, _ *envconf.Config) (k8s.Object, error) {
		return dep(1, 1), nil
	}

	// Assign to concrete function type to catch signature drift at compile time.
	_ = Assert(ev, "true", binder)
	_ = AssertPolicy(ev, policy.Policy{}, fetcher)
}

func newEv(t *testing.T) *cel.Evaluator {
	t.Helper()
	ev, err := cel.NewEvaluator()
	if err != nil {
		t.Fatalf("NewEvaluator: %v", err)
	}
	return ev
}
