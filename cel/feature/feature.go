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

// Package feature adapts the cel primitives into features.Func values
// so CEL assertions read as a one-line Assess in a test.
package feature

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/cel"
	"sigs.k8s.io/e2e-framework/cel/policy"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// BinderFunc produces a Bindings map for an assertion. The test environment
// is supplied so binders can fetch live objects from the cluster.
type BinderFunc func(context.Context, *envconf.Config) (cel.Bindings, error)

// FetcherFunc returns a Kubernetes object to be bound as `object`. It is
// the common simpler case of BinderFunc for policy assertions that only
// look at one object.
type FetcherFunc func(context.Context, *envconf.Config) (k8s.Object, error)

// Assert returns a features.Func that evaluates expr against the bindings
// produced by binder and fails the assessment via t.Fatal if the expression
// does not evaluate to true.
func Assert(ev *cel.Evaluator, expr string, binder BinderFunc) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		t.Helper()
		bindings, err := binder(ctx, cfg)
		if err != nil {
			t.Fatalf("cel: bind: %v", err)
		}
		if err := ev.Assert(expr, bindings); err != nil {
			t.Fatal(err)
		}
		return ctx
	}
}

// AssertPolicy fetches an object and runs every Validation in pol against
// it. Any failure ends the assessment; t.Fatal reports every failure so the
// test output shows the full admission story.
func AssertPolicy(ev *cel.Evaluator, pol policy.Policy, fetcher FetcherFunc) features.Func {
	return func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		t.Helper()
		obj, err := fetcher(ctx, cfg)
		if err != nil {
			t.Fatalf("cel: fetch: %v", err)
		}
		if res := pol.Check(ev, obj); !res.Passed() {
			t.Fatal(res.Err())
		}
		return ctx
	}
}

// AssertObject is a one-line shortcut for the common case: fetch target by
// name in the test's namespace (cfg.Namespace()), bind it to `object`, and
// assert expr against it. For a specific namespace use AssertObjectIn.
func AssertObject(ev *cel.Evaluator, expr string, target k8s.Object, name string) features.Func {
	return Assert(ev, expr, AsBinder(Fetch(target, name, "")))
}

// AssertObjectIn is like AssertObject but fetches from the given namespace
// rather than cfg.Namespace(). Use for cluster-scoped resources (pass "")
// or to cross-reference a resource in another namespace.
func AssertObjectIn(ev *cel.Evaluator, expr string, target k8s.Object, name, namespace string) features.Func {
	return Assert(ev, expr, AsBinder(Fetch(target, name, namespace)))
}

// AssertPolicyOnObject is a shortcut for running pol against target fetched
// by name in cfg.Namespace(). For a specific namespace use
// AssertPolicyOnObjectIn.
func AssertPolicyOnObject(ev *cel.Evaluator, pol policy.Policy, target k8s.Object, name string) features.Func {
	return AssertPolicy(ev, pol, Fetch(target, name, ""))
}

// AssertPolicyOnObjectIn is like AssertPolicyOnObject but fetches from the
// given namespace.
func AssertPolicyOnObjectIn(ev *cel.Evaluator, pol policy.Policy, target k8s.Object, name, namespace string) features.Func {
	return AssertPolicy(ev, pol, Fetch(target, name, namespace))
}
