# CEL Assertions

This document proposes the design for a set of CEL ([Common Expression Language](https://github.com/google/cel-spec)) utilities in a new top-level package, `cel`, intended to let test developers write declarative assertions against Kubernetes objects in the same language the Kubernetes API server uses for admission and CRD validation. The goal of these utilities is to make test assertions terse, readable, and consistent with the CEL expressions a project already ships in its admission policies and CRDs.

## Table of Contents

1. [Motivation](#Motivation)
2. [Supported CEL environments](#supported-cel-environments)
3. [Goals](#Goals)
4. [Non-Goals](#Non-Goals)
5. [Design Components](#Design-Components)
    * [Evaluator](#Evaluator)
    * [Variable Bindings](#Variable-Bindings)
    * [CEL Library Composition](#CEL-Library-Composition)
    * [Policy](#Policy)
    * [Feature Helpers](#Feature-Helpers)
6. [CEL Proposal](#CEL-Proposal)
    * [Pre-defined Evaluators and Bindings](#pre-defined-evaluators-and-bindings)
    * [Pre-defined Helpers](#Pre-defined-Helpers)
    * [Wait Integration](#Wait-Integration)
    * [Decoder Integration](#Decoder-Integration)

## Motivation

When developing tests for Kubernetes components, it is common to fetch an object using the `klient` client and then assert some property of it — that a Deployment is fully rolled out, that a Service has the expected selector, that a ConfigMap contains a specific key. These assertions are typically written as Go field accessors, which gets verbose quickly and reads differently from the CRD validation rules or admission policies the same project ships.

Kubernetes uses CEL for CRD `x-kubernetes-validations`, `ValidatingAdmissionPolicy`, and `MutatingAdmissionPolicy`. A test author wanting to exercise the same invariant a policy enforces has to translate the CEL expression into Go, which is easy to get wrong and adds drift between the test and the policy.

A CEL utility in `cel` removes the translation step. The same expression that appears in a `ValidatingAdmissionPolicy` can appear in a test assertion, bound to the same variable names. Tests that need to check several invariants at once can group them into a reusable `Policy` value or chain multiple `Assess(...)` calls on a `features.Feature`, rather than inline every accessor.

## Supported CEL environments

- Admission (`object`, `oldObject`, `request`, `params`, `namespaceObject`, `authorizer`, `variables`)
- CRD validation (`self`, `oldSelf`)

The admission environment is the default. It matches what a `ValidatingAdmissionPolicy` author sees, and is what most tests will want. The CRD environment is available via `WithEnvironment(EnvCRD)` for tests that assert on values scoped to a CRD field path.

## Goals

- Provide a simple way to assert that a Kubernetes object satisfies a CEL expression, returning a Go `error`.
- Support the variables and library functions the Kubernetes API server registers for admission CEL (`quantity`, `url`, `ip`, `cidr`, `regex`, `lists`, `format`, `semver`, `authz`, `jsonpatch`).
- Support unit-testing `ValidatingAdmissionPolicy` validations offline against fixture objects, without a live API server.
- Integrate with `pkg/features` so a CEL assertion reads as a one-line `Assess(...)`.
- Cache compiled CEL programs to avoid repeated compile cost across assertions.

## Non-Goals

- Replacing integration tests against a live admission webhook. The offline `Policy` check evaluates CEL against fixture objects; it does not install a policy on a cluster.
- Server-side-apply merge semantics for `MutatingAdmissionPolicy` `ApplyConfiguration` patches.
- A policy management layer. Applying or syncing policies to a cluster is the responsibility of `klient/decoder` and `klient/k8s/resources`.
- CEL authoring tooling (syntax highlighting, completion, lint).

## Design Components

### **Evaluator**

```go
type Evaluator struct {
    // unexported
}

func NewEvaluator(opts ...Option) (*Evaluator, error)
```

The `Evaluator` owns a `*cel.Env`, a compiled-program cache, and a default set of variable bindings. It is the single primitive every other component uses. The zero-option constructor returns an evaluator configured for the admission environment with all standard Kubernetes CEL libraries wired in.

An `Evaluator` is safe for concurrent use, so the same instance can be shared across parallel features.

```go
type Option func(*options)
```

`Option` values are used to narrow or widen the default configuration.

Example pre-defined Options:

```go
// select the CEL environment (admission or CRD)
func WithEnvironment(env Env) Option
// restrict the wired-in Kubernetes CEL libraries
func WithLibraries(libs ...Library) Option
// bind the `authorizer` variable to a live *envconf.Config for SAR checks
func WithAuthorizer(cfg *envconf.Config) Option
// override the per-expression CEL cost limit (default matches API server)
func WithCostLimit(limit uint64) Option
```

### **Variable Bindings**

```go
type Bindings map[string]any
```

`Bindings` is the variable map passed to CEL at evaluation time. Pre-defined binding helpers convert `k8s.Object` and related types into a `Bindings` value using the variable names the admission CEL environment expects.

```go
func ObjectBinding(obj k8s.Object) Bindings
func OldObjectBinding(obj k8s.Object) Bindings
func ParamsBinding(params k8s.Object) Bindings
func NamespaceBinding(ns *corev1.Namespace) Bindings
func RequestBinding(req *AdmissionRequest) Bindings

// CRD environment bindings — only valid when WithEnvironment(EnvCRD) is set
func SelfBinding(self k8s.Object) Bindings
func OldSelfBinding(oldSelf k8s.Object) Bindings

// Compose multiple bindings; later keys override earlier ones.
func Bind(parts ...Bindings) Bindings
```

Binding helpers convert objects via `runtime.DefaultUnstructuredConverter` so CEL sees the same field shape the API server does.

### **CEL Library Composition**

```go
type Library int

const (
    LibAuthz Library = iota
    LibAuthzSelectors
    LibQuantity
    LibURLs
    LibIP
    LibCIDR
    LibRegex
    LibLists
    LibFormat
    LibSemver
    LibJSONPatch
)
```

By default, `NewEvaluator` wires in every CEL library the Kubernetes API server registers for admission. Users can select a narrower set with `WithLibraries`.

The `Authz` library is the one exception worth calling out: its function bodies issue `SubjectAccessReview` calls, so tests that use `authorizer.check(...)` need an `Evaluator` built with `WithAuthorizer(cfg)` pointing at a live `*envconf.Config`.

### **Policy**

```go
type Policy struct {
    Name        string
    Validations []Validation
}

type Validation struct {
    Expression string
    Message    string
    Reason     metav1.StatusReason
}

type Result struct {
    PolicyName string
    Failures   []Failure
}

type Failure struct {
    Validation Validation
    Err        error
}
```

`Policy` is the offline analogue of `ValidatingAdmissionPolicy.spec.validations`. `Policy.Check` runs every validation against a candidate object and accumulates failures rather than short-circuiting, so tests can report the full admission story in one pass.

```go
func (p Policy) Check(ev *Evaluator, obj k8s.Object) Result

func (r Result) Passed() bool
func (r Result) Err() error
```

A companion helper converts a real `admissionregistrationv1.ValidatingAdmissionPolicy` into a testable `Policy`:

```go
func FromVAP(vap *admissionregistrationv1.ValidatingAdmissionPolicy) Policy
```

Paired with `klient/decoder`, a test can decode the same `ValidatingAdmissionPolicy` manifest the operator ships and check its validations against a fixture object — no re-expressing the rules in Go.

### **Feature Helpers**

A thin `cel/feature` sub-package adapts primitives into `features.Func` values, so CEL primitives stay test-framework agnostic.

```go
type BinderFunc  func(context.Context, *envconf.Config) (Bindings, error)
type FetcherFunc func(context.Context, *envconf.Config) (k8s.Object, error)

func Assert(ev *Evaluator, expr string, binder BinderFunc) features.Func
func AssertPolicy(ev *Evaluator, pol Policy, fetcher FetcherFunc) features.Func
```

The two primitives above take a `BinderFunc` or `FetcherFunc` that the
caller supplies. For the 80% case — fetch a single named resource in the
test namespace and assert on it — a pair of shortcut helpers collapses
fetch + bind + assert into one call:

```go
// AssertObject fetches target by name in cfg.Namespace() and asserts expr.
func AssertObject(ev *Evaluator, expr string, target k8s.Object, name string) features.Func
// AssertObjectIn is like AssertObject but uses the given namespace.
func AssertObjectIn(ev *Evaluator, expr string, target k8s.Object, name, namespace string) features.Func

// AssertPolicyOnObject runs pol against target fetched by name in cfg.Namespace().
func AssertPolicyOnObject(ev *Evaluator, pol Policy, target k8s.Object, name string) features.Func
func AssertPolicyOnObjectIn(ev *Evaluator, pol Policy, target k8s.Object, name, namespace string) features.Func
```

The low-level fetch helpers stay exported for callers composing multi-binding assertions:

```go
// Fetch returns a FetcherFunc that Gets target by name from namespace.
// If namespace is empty, cfg.Namespace() is used.
func Fetch(target k8s.Object, name, namespace string) FetcherFunc
// AsBinder adapts a FetcherFunc into a BinderFunc binding `object`.
func AsBinder(f FetcherFunc) BinderFunc
```

Each helper wraps the underlying CEL call in a `features.Func`, calls `t.Fatal` on failure, and returns the context unchanged.

## CEL Proposal

Proposal on the function signatures:

```go
// NewEvaluator returns an Evaluator configured for the admission CEL
// environment with all standard Kubernetes CEL libraries wired in.
// Options narrow or widen the default.
func NewEvaluator(opts ...Option) (*Evaluator, error)

// Eval compiles expr and evaluates it against b, returning the raw result.
// Compiled programs are cached per-expression inside the Evaluator.
func (e *Evaluator) Eval(expr string, b Bindings) (ref.Val, error)

// Assert is the common case: returns nil iff expr evaluates to boolean true.
func (e *Evaluator) Assert(expr string, b Bindings) error
```

Usage:

```go
ev, _ := cel.NewEvaluator()

var dep appsv1.Deployment
_ = cfg.Client().Resources().Get(ctx, "demo", "default", &dep)

err := ev.Assert(
    "object.status.readyReplicas == object.spec.replicas",
    cel.ObjectBinding(&dep),
)
```

With a narrower library set:

```go
ev, _ := cel.NewEvaluator(
    cel.WithLibraries(cel.LibQuantity, cel.LibRegex, cel.LibFormat),
)
```

With the CRD environment:

```go
ev, _ := cel.NewEvaluator(cel.WithEnvironment(cel.EnvCRD))
err := ev.Assert("self.size() > 0", cel.SelfBinding(&cr))
```

### Pre-defined Evaluators and Bindings

Building on the proposal, the following helpers would be included to reduce boilerplate for common setups.

```go
// ObjectBinding wraps an object in a Bindings value using the admission
// variable name `object`.
func ObjectBinding(obj k8s.Object) Bindings
// OldObjectBinding binds `oldObject` — useful for UPDATE-path policy tests.
func OldObjectBinding(obj k8s.Object) Bindings
// ParamsBinding binds `params` — the param resource a VAP references.
func ParamsBinding(params k8s.Object) Bindings
// NamespaceBinding binds `namespaceObject`.
func NamespaceBinding(ns *corev1.Namespace) Bindings
// RequestBinding binds `request` — a test-constructed AdmissionRequest subset.
func RequestBinding(req *AdmissionRequest) Bindings

// Bind composes multiple bindings. Later keys override earlier ones.
func Bind(parts ...Bindings) Bindings
```

Usage, composing several bindings for an UPDATE-path policy test:

```go
b := cel.Bind(
    cel.ObjectBinding(newDep),
    cel.OldObjectBinding(oldDep),
    cel.RequestBinding(&cel.AdmissionRequest{Operation: "UPDATE"}),
)
err := ev.Assert(
    "object.spec.replicas >= oldObject.spec.replicas",
    b,
)
```

### Pre-defined Helpers

```go
// FromVAP converts a live ValidatingAdmissionPolicy into a testable Policy.
func FromVAP(vap *admissionregistrationv1.ValidatingAdmissionPolicy) Policy
```

Usage:

```go
var vap admissionregistrationv1.ValidatingAdmissionPolicy
_ = decoder.Decode(strings.NewReader(policyYAML), &vap)

pol := cel.FromVAP(&vap)
res := pol.Check(ev, &dep)
if !res.Passed() {
    t.Fatal(res.Err())
}
```

`features.Func` adapters:

```go
// Assert evaluates expr against the bindings produced by binder.
func Assert(ev *Evaluator, expr string, binder BinderFunc) features.Func
// AssertPolicy fetches an object and runs every validation in pol against it.
func AssertPolicy(ev *Evaluator, pol Policy, fetcher FetcherFunc) features.Func
```

Usage in a feature:

```go
ev, _ := cel.NewEvaluator()

f := features.New("deployment is fully ready").
    Assess("replicas match", feature.Assert(
        ev,
        "object.status.readyReplicas == object.spec.replicas",
        fetchDeployment("demo", "default"),
    )).
    Feature()
```

Chaining multiple invariants on the same feature:

```go
f := features.New("baseline conformance").
    Assess("replicas match",
        feature.AssertObject(ev,
            "object.status.readyReplicas == object.spec.replicas",
            &appsv1.Deployment{}, "demo")).
    Assess("at least one replica",
        feature.AssertObject(ev,
            "object.spec.replicas >= 1",
            &appsv1.Deployment{}, "demo")).
    Feature()
```

### Wait Integration

`klient/wait.For` drives polling against a `ConditionWithContextFunc`. A
companion `cel/wait` package supplies conditions whose predicate is
expressed in CEL, so the same polling machinery that powers `wait.For` can
terminate on any CEL invariant without a hand-written matcher.

```go
// Match returns a ConditionWithContextFunc that refetches target on every
// poll and evaluates expr against it. Succeeds when expr returns true.
func Match(r *resources.Resources, ev *cel.Evaluator, target k8s.Object,
    name, namespace, expr string) apimachinerywait.ConditionWithContextFunc

// MatchAny is like Match but accepts additional bindings (for example a
// `request` binding) that are merged in alongside `object`.
func MatchAny(r *resources.Resources, ev *cel.Evaluator, target k8s.Object,
    name, namespace, expr string, extra cel.Bindings) apimachinerywait.ConditionWithContextFunc
```

Fetch errors (including `NotFound`) are treated as transient so the poll
keeps retrying until the overall `wait.For` timeout elapses. CEL compile
and type errors are terminal and halt the poll immediately.

Usage, waiting for a Deployment to roll out:

```go
err := wait.For(
    celwait.Match(client.Resources(), ev, &appsv1.Deployment{}, "cel-demo", cfg.Namespace(),
        "object.status.readyReplicas == object.spec.replicas"),
    wait.WithTimeout(2*time.Minute),
)
```

Waiting with an extra binding:

```go
err := wait.For(
    celwait.MatchAny(client.Resources(), ev, &corev1.Pod{}, "demo", cfg.Namespace(),
        `object.status.phase == "Running" && request.dryRun == false`,
        cel.RequestBinding(&cel.AdmissionRequest{DryRun: false}),
    ),
    wait.WithTimeout(time.Minute),
)
```

### Decoder Integration

`klient/decoder` already reads YAML/JSON into `k8s.Object` values (single
or multi-document, file, string, or URL). A `cel/decoder` sub-package
layers CEL assertions on top, with two shapes: `HandlerFunc` factories that
plug into streaming decoders, and one-shot helpers for the common cases.

```go
// AssertHandler returns a decoder.HandlerFunc that asserts expr against
// every decoded object.
func AssertHandler(ev *cel.Evaluator, expr string) decoder.HandlerFunc

// PolicyHandler returns a decoder.HandlerFunc that runs pol against every
// decoded object.
func PolicyHandler(ev *cel.Evaluator, pol policy.Policy) decoder.HandlerFunc

// AssertYAML decodes a single-document manifest and asserts expr.
func AssertYAML(ev *cel.Evaluator, expr string, manifest io.Reader) error

// AssertYAMLAll decodes a multi-document stream and asserts expr against
// every object, halting on the first failure.
func AssertYAMLAll(ctx context.Context, ev *cel.Evaluator, expr string, manifest io.Reader) error
```

Usage, asserting every object in a multi-document manifest has a namespace:

```go
err := celdecoder.AssertYAMLAll(ctx, ev,
    `object.metadata.namespace != ""`,
    strings.NewReader(manifestYAML))
```

Usage with an existing decoder stream:

```go
err := decoder.DecodeEachFile(ctx, os.DirFS("testdata"), "*.yaml",
    celdecoder.AssertHandler(ev, `object.kind != "Pod"`))
```

Usage running a full policy per decoded object:

```go
pol := policy.Policy{
    Name: "shape",
    Validations: []policy.Validation{
        {Expression: `has(object.metadata.namespace)`, Message: "must set namespace"},
        {Expression: `object.metadata.name != ""`, Message: "must set name"},
    },
}
err := decoder.DecodeEach(ctx, strings.NewReader(manifest),
    celdecoder.PolicyHandler(ev, pol))
```
