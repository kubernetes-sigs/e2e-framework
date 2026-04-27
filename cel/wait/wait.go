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

// Package wait provides CEL-based conditions for use with klient/wait.For.
// Each condition refetches the target on every poll and evaluates a CEL
// expression against it, succeeding when the expression returns true.
package wait

import (
	"context"
	"strings"

	apimachinerywait "k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/e2e-framework/cel"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

// Match returns a ConditionWithContextFunc that refetches target from r on
// every poll and evaluates expr against it. The condition is satisfied when
// expr evaluates to true.
//
// On each poll target is Get-ted by name in namespace and its fields are
// populated in place. A Get error (including NotFound) is treated as "not
// yet" rather than a terminal failure so the poll keeps retrying until the
// overall wait.For timeout elapses. A CEL compile or type error IS
// terminal and halts the poll immediately.
func Match(r *resources.Resources, ev *cel.Evaluator, target k8s.Object, name, namespace, expr string) apimachinerywait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		if err := r.Get(ctx, name, namespace, target); err != nil {
			// Treat fetch errors as transient: the object may not exist
			// yet. wait.For will stop polling on timeout.
			return false, nil
		}
		if err := ev.Assert(expr, cel.ObjectBinding(target)); err != nil {
			if isTerminal(err) {
				return false, err
			}
			return false, nil
		}
		return true, nil
	}
}

// MatchAny is like Match but accepts caller-supplied bindings in addition
// to `object`. Use when the expression references more than the fetched
// target (for example a request binding or a second object).
//
// target is still fetched on every poll and bound to `object`; the extra
// bindings are merged in afterwards via cel.Bind (so they can override
// `object` if a caller really wants to).
func MatchAny(r *resources.Resources, ev *cel.Evaluator, target k8s.Object, name, namespace, expr string, extra cel.Bindings) apimachinerywait.ConditionWithContextFunc {
	return func(ctx context.Context) (bool, error) {
		if err := r.Get(ctx, name, namespace, target); err != nil {
			return false, nil
		}
		bindings := cel.Bind(cel.ObjectBinding(target), extra)
		if err := ev.Assert(expr, bindings); err != nil {
			if isTerminal(err) {
				return false, err
			}
			return false, nil
		}
		return true, nil
	}
}

// isTerminal reports whether a CEL error should stop polling outright
// rather than being retried. Compile errors and type errors cannot be
// fixed by retrying, so they halt immediately; a simple "assertion failed"
// just means the expression returned false and we should poll again.
func isTerminal(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "compile ") || strings.Contains(msg, "want bool")
}
