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

// Package policy provides offline evaluation of ValidatingAdmissionPolicy-
// shaped CEL validations. It lets test authors unit-test the CEL rules in
// a policy against fixture objects without standing up a live API server.
package policy

import (
	"errors"
	"fmt"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/cel"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

// Policy is a named collection of CEL validations, mirroring the shape of
// ValidatingAdmissionPolicy.spec.validations.
type Policy struct {
	Name        string       `json:"name"`
	Validations []Validation `json:"validations"`
}

// Validation is one CEL rule within a Policy.
type Validation struct {
	// Expression is the CEL expression evaluated against the target object.
	// It must return a bool; a true result admits, false rejects.
	Expression string `json:"expression"`
	// Message is a human-readable failure explanation surfaced when the
	// expression evaluates to false. Defaults to Expression when empty.
	Message string `json:"message,omitempty"`
	// Reason mirrors ValidatingAdmissionPolicy.spec.validations[*].reason.
	Reason metav1.StatusReason `json:"reason,omitempty"`
}

// Failure captures a single failed Validation within a Result.
type Failure struct {
	Validation Validation `json:"validation"`
	Err        error      `json:"error,omitempty"`
}

// Result is the report of a Policy.Check.
type Result struct {
	PolicyName string    `json:"policyName"`
	Failures   []Failure `json:"failures,omitempty"`
}

// Passed reports whether every validation admitted the object.
func (r Result) Passed() bool { return len(r.Failures) == 0 }

// Err returns a joined error covering every failure, or nil if the result
// passed.
func (r Result) Err() error {
	if r.Passed() {
		return nil
	}
	errs := make([]error, 0, len(r.Failures))
	for _, f := range r.Failures {
		msg := f.Validation.Message
		if msg == "" {
			msg = f.Validation.Expression
		}
		errs = append(errs, fmt.Errorf("%s: %s: %w", r.PolicyName, msg, f.Err))
	}
	return errors.Join(errs...)
}

// Check runs every Validation against obj and returns a Result. All
// validations are evaluated (no short-circuit on first failure), matching
// the admission path's accumulation of failures when failurePolicy is Fail.
func (p Policy) Check(ev *cel.Evaluator, obj k8s.Object) Result {
	res := Result{PolicyName: p.Name}
	bindings := cel.ObjectBinding(obj)
	for _, v := range p.Validations {
		if err := ev.Assert(v.Expression, bindings); err != nil {
			res.Failures = append(res.Failures, Failure{Validation: v, Err: err})
		}
	}
	return res
}

// FromVAP converts a ValidatingAdmissionPolicy manifest into a testable
// Policy. Only the Expression, Message, and Reason fields are carried over;
// paramKind, matchConstraints, and matchConditions are not evaluated (this
// is an offline validation of the CEL rules, not a full admission
// simulation).
func FromVAP(vap *admissionregistrationv1.ValidatingAdmissionPolicy) Policy {
	if vap == nil {
		return Policy{}
	}
	p := Policy{
		Name:        vap.Name,
		Validations: make([]Validation, 0, len(vap.Spec.Validations)),
	}
	for _, v := range vap.Spec.Validations {
		conv := Validation{
			Expression: v.Expression,
			Message:    v.Message,
		}
		if v.Reason != nil {
			conv.Reason = *v.Reason
		}
		p.Validations = append(p.Validations, conv)
	}
	return p
}
