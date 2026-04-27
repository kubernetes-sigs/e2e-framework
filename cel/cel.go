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

// Package cel provides CEL-based assertion utilities for testing Kubernetes
// objects. It wraps cel-go with the variables and library set that the
// Kubernetes API server registers for admission CEL, so the same expressions
// that appear in a ValidatingAdmissionPolicy can appear in a test assertion.
package cel

import (
	"fmt"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// Env selects which CEL variable shape the Evaluator exposes.
type Env int

const (
	// EnvAdmission exposes object, oldObject, request, params,
	// namespaceObject, and authorizer (matching ValidatingAdmissionPolicy).
	EnvAdmission Env = iota
	// EnvCRD exposes self and oldSelf (matching x-kubernetes-validations).
	EnvCRD
)

// DefaultCostLimit matches the API server's per-expression CEL cost limit.
const DefaultCostLimit uint64 = 10_000_000

// Evaluator compiles and evaluates CEL expressions against variable bindings.
// An Evaluator is safe for concurrent use; compiled programs are cached so
// the same expression is not re-compiled across assertions.
type Evaluator struct {
	env   *cel.Env
	cache sync.Map // expression -> cel.Program
	opts  options
}

type options struct {
	env        Env
	libSet     librarySet
	costLimit  uint64
	customVars []cel.EnvOption
}

// Option configures an Evaluator.
type Option func(*options)

// WithEnvironment selects the CEL variable shape. Defaults to EnvAdmission.
func WithEnvironment(env Env) Option {
	return func(o *options) { o.env = env }
}

// WithLibraries restricts the Kubernetes CEL libraries wired into the
// evaluator. By default every library registered for admission CEL is wired
// in; passing WithLibraries replaces that set with the provided subset.
func WithLibraries(libs ...Library) Option {
	return func(o *options) {
		o.libSet = newLibrarySet(libs)
	}
}

// WithCostLimit overrides the per-expression CEL cost limit. The default is
// DefaultCostLimit, matching the API server. Passing 0 disables the limit.
func WithCostLimit(limit uint64) Option {
	return func(o *options) { o.costLimit = limit }
}

// WithVariables is an escape hatch for advanced use: the supplied EnvOptions
// are appended after the default variables and libraries are registered.
func WithVariables(decls ...cel.EnvOption) Option {
	return func(o *options) { o.customVars = append(o.customVars, decls...) }
}

// NewEvaluator returns an Evaluator configured for the admission CEL
// environment with every standard Kubernetes CEL library wired in.
func NewEvaluator(opts ...Option) (*Evaluator, error) {
	o := options{
		env:       EnvAdmission,
		libSet:    allLibraries(),
		costLimit: DefaultCostLimit,
	}
	for _, opt := range opts {
		opt(&o)
	}

	envOpts := make([]cel.EnvOption, 0, 16)
	envOpts = append(envOpts, variableDecls(o.env)...)
	envOpts = append(envOpts, o.libSet.envOptions()...)
	envOpts = append(envOpts, o.customVars...)

	env, err := cel.NewEnv(envOpts...)
	if err != nil {
		return nil, fmt.Errorf("cel: build env: %w", err)
	}
	return &Evaluator{env: env, opts: o}, nil
}

// Eval compiles expr and evaluates it against b, returning the raw result.
// Compiled programs are cached per-expression inside the Evaluator.
func (e *Evaluator) Eval(expr string, b Bindings) (ref.Val, error) {
	prg, err := e.programFor(expr)
	if err != nil {
		return nil, err
	}
	out, _, err := prg.Eval(map[string]any(b))
	if err != nil {
		return nil, fmt.Errorf("cel: eval %q: %w", expr, err)
	}
	return out, nil
}

// Assert returns nil iff expr evaluates to boolean true. Non-boolean results
// and compile/evaluation errors are surfaced as errors.
func (e *Evaluator) Assert(expr string, b Bindings) error {
	out, err := e.Eval(expr, b)
	if err != nil {
		return err
	}
	bv, ok := out.(types.Bool)
	if !ok {
		return fmt.Errorf("cel: expression %q returned %T, want bool", expr, out)
	}
	if !bool(bv) {
		return fmt.Errorf("cel: assertion failed: %s", expr)
	}
	return nil
}

// programFor returns a cached compiled program for expr, compiling on first
// use. Multiple goroutines racing on the same expression may each compile,
// but only one result is retained; cel.Program is concurrency-safe.
func (e *Evaluator) programFor(expr string) (cel.Program, error) {
	if p, ok := e.cache.Load(expr); ok {
		prg, ok := p.(cel.Program)
		if !ok {
			return nil, fmt.Errorf("cel: cached value for %q is %T, want cel.Program", expr, p)
		}
		return prg, nil
	}
	ast, iss := e.env.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("cel: compile %q: %w", expr, iss.Err())
	}
	progOpts := []cel.ProgramOption{cel.EvalOptions(cel.OptOptimize)}
	if e.opts.costLimit > 0 {
		progOpts = append(progOpts, cel.CostLimit(e.opts.costLimit))
	}
	prg, err := e.env.Program(ast, progOpts...)
	if err != nil {
		return nil, fmt.Errorf("cel: program %q: %w", expr, err)
	}
	actual, _ := e.cache.LoadOrStore(expr, prg)
	stored, ok := actual.(cel.Program)
	if !ok {
		return nil, fmt.Errorf("cel: cached value for %q is %T, want cel.Program", expr, actual)
	}
	return stored, nil
}
