# Decoding Resources

This document proposes the design for a set of decoding functions a new package, `klient/decoder`, intended to provide utilities for creating `k8s.Object` types from common sources of input in Go programs: files, strings, or any type that satisfies the [io.Reader](https://pkg.go.dev/io#Reader) interface. The goal of these decoding functions is to provide an easy way for test developers to interact with Kubernetes objects in their Go tests.

## Table of Contents

- [Decoding Resources](#decoding-resources)
  - [Table of Contents](#table-of-contents)
  - [Motivation](#motivation)
  - [Supported object formats](#supported-object-formats)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
  - [Design Components](#design-components)
    - [**Decoding Options**](#decoding-options)
    - [**Handlers**](#handlers)
    - [Decoding a single-document YAML/JSON input](#decoding-a-single-document-yamljson-input)
    - [Decoding a multi-document YAML/JSON input](#decoding-a-multi-document-yamljson-input)
      - [**Decoding without knowing the object type**](#decoding-without-knowing-the-object-type)
  - [Decode Proposal](#decode-proposal)
    - [Pre-defined Decoders](#pre-defined-decoders)
    - [Pre-defined Helpers](#pre-defined-helpers)

## Motivation

When developing tests that are meant to utilize Kubernetes APIs, it is expected that you will construct a `k8s.Object` type in order to use many functions defined in the `e2e-framework` packages (as in the `klient` package).

This may be accomplished by defining Go structs, importing stdlib or third-party code.

When developing many tests, the verbosity and complexity of defining many types and understanding which packages to import may add a burden to test developers. Managing these resources as YAML or JSON has obvious benefits in regard to maintainability (and even extensibility), as they are how these resource types are traditionally represented in documentation and utilized in actual deployments.

In Go, `testdata` is a special directory that can be used to store such test fixtures, and using such a testdata directory as a source of easy-to-manage files is a common pattern associated with table-driven testing.

Finally, to help develop feature tests, it is common to need to have a set of resources created before a feature assessment begins. Similarly, deleting a set of resources may be required in a teardown step.

## Supported object formats

- YAML
- JSON

## Goals

- Support decoding [single-document](https://yaml.org/spec/1.2.2/#91-documents) YAML/JSON input
- Support decoding a [multi-document](https://yaml.org/spec/1.2.2/#92-streams) YAML/JSON stream input
- Accept io.Reader interface as input

## Non-Goals

- Encoding Objects

## Design Components

### **Decoding Options**

```go
type DecodeOption struct {
    DefaultGVK  *schema.GroupVersionKind
    MutateFuncs []MutateFunc
}

type DecodeOption func(*DecodeOption)

type MutateFunc func(k8s.Object) error
```

All decoding functions accept a `options ...DecodeOption` argument. Options may be used to apply "patches", or post-decoding mutations to Objects to inject data
after decoding is completed. Additionally, the Group Version Kind may be specified to instruct the decoding process on the type to use.

If a MutateFunc returns an error, decoding is halted.

This may be done to inject dynamic data that may not be known until runtime or that may be sensitive like a locally valid credential.

Example pre-defined MutateFuncs, wrapped as DecodeOptions:

```go
// apply an override set of labels to a decoded object
func MutateLabels(overrides map[string]string) DecodeOption
// apply an override set of annotations to a decoded object
func MutateAnnotations(overrides map[string]string) DecodeOption
// apply an owner annotation to a decoded object
func MutateOwnerAnnotations(owner k8s.Object) DecodeOption
```

### **Handlers**

Some decoding functions accept a HandlerFunc, a function that is executed after decoding and the optional patches are completed per each object.

If a HandlerFunc returns an error, decoding is halted.

```go
type HandlerFunc func(context.Context, k8s.Object) error
```

Example pre-defined HandlerFuncs:

```go
// CreateHandler returns a HandlerFunc that will create objects
func CreateHandler(*resources.Resources, opts ...CreateOption) HandlerFunc
// UpdateHandler returns a HandlerFunc that will update objects
func UpdateHandler(*resources.Resources, opts ...UpdateOption) HandlerFunc
// DeleteHandler returns a HandlerFunc that will delete objects
func DeleteHandler(*resources.Resources, opts ...DeleteOption) HandlerFunc

// IgnoreErrorHandler returns a HandlerFunc that will ignore an error
func IgnoreErrorHandler(HandlerFunc, error) HandlerFunc

// CreateIfNotExistsHandler returns a HandlerFunc that will create objects if they do not already exist
func CreateIfNotExistsHandler(*resources.Resources, opts ...CreateOption) HandlerFunc
```

### Decoding a single-document YAML/JSON input

The following are proposed function signatures for decoding input that contain a single `k8s.Object` type. Example:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-service-account
  namespace: myappns
```

1. Decoding an object to a known type

```go
func Decode(manifest io.Reader, obj k8s.Object, options ...DecodeOption) error
```

Usage:

```go
sa := v1.ServiceAccount{}
err := Decode(strings.NewReader("..."), &sa)
```

With a MutateFunc:

```go
// Decode to sa and apply the label "test" : "feature-X"
sa := v1.ServiceAccount{}
err := Decode(strings.NewReader("..."), &sa, MutateLabels(map[string]string{"test" : "feature-X"}))
```

2. Decoding an object without knowing the type

`defaults` is an optional parameter, if specified, it is a hint to the decoder to help determine the underlying Go type to use for object creation.

```go
func DecodeAny(manifest io.Reader, options ...DecodeOption) (k8s.Object, error)
```

Usage:

```go
obj, err := DecodeAny(strings.NewReader("..."))
if err != nil {
    ...
}
if sa, ok := obj.(*v1.ServiceAccount); ok {
    ...
}
```

With defaults:

```go
obj, err := DecodeAny(strings.NewReader("..."), schema.GroupVersionKind{Version: "v1", Kind: "ServiceAccount"})
if err != nil {
    ...
}
if sa, ok := obj.(*v1.ServiceAccount); ok {
    ...
}
```

### Decoding a multi-document YAML/JSON input

The following are proposed function signatures for decoding input that may contain multiple distinct `k8s.Object` types. Example:

```yaml
## testdata/test-setup.yaml
apiVersion: v1
kind: Namespace
name: myappns
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-config
  namespace: myappns
data:
  appconfig.json: |
    key: value
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: my-service-account
  namespace: myappns
```

#### **Decoding without knowing the object type**

The following options would use the types registered with `scheme.Scheme` to help deserialize objects into a `k8s.Object` with the expected underlying API type.

`defaults` is an optional parameter, if specified, it is a hint to the decoder to help determine the underlying Go type to use for object creation.

1. Decode each document and call handlerFn for each processed object. If `handlerFn` returns an error, decoding is halted.

```go
func DecodeEach(ctx context.Context, manifest io.Reader, handlerFn HandlerFunc, options ...DecodeOption) error
```

Usage:

```go
list := &unstructured.UnstructuredList{}
err := DecodeEach(context.TODO(), strings.NewReader("..."), func(ctx context.Context, obj ks8.Object) error {
    if cfg, ok := obj.(*v1.ConfigMap); ok {
        // obj is a ConfigMap
    } else if svc, ok := obj.(*v1.ServiceAccount); ok {
        // obj is a ServiceAccount
    }
    return klient.Create(obj)
})
```

Usage with pre-defined HandlerFunc:

```go
err := DecodeEach(context.TODO(), strings.NewReader("..."), CreateHandler(klient.Resources(namespace)))
```

2. Decode all documents.

```go
func DecodeAll(ctx context.Context, manifest io.Reader, options ...DecodeOption) ([]k8s.Object, error)
```

Usage:

```go
objects, err := DecodeAll(context.TODO(), strings.NewReader("..."))
for _, obj := range objects {
    err := klient.Create(obj)
    ...
}
```

## Decode Proposal

The following is a final proposal on the function signatures, after considering the above options:

```go
// Decode a single-document YAML or JSON input into a known type.
// Patches are optional and applied after decoding.
func Decode(manifest io.Reader, obj k8s.Object, options ...DecodeOption) error

// Decode any single-document YAML or JSON input using either the innate typing of the scheme or the default kind, group, and version provided.
// Patches are optional and applied after decoding.
func DecodeAny(manifest io.Reader, options ...DecodeOption) (k8s.Object, error)

// Decode a stream of documents of any Kind using either the innate typing of the scheme or the default kind, group, and version provided.
// If handlerFn returns an error, decoding is halted.
// Patches are optional and applied after decoding and before handlerFn is executed.
func DecodeEach(ctx context.Context, manifest io.Reader, handlerFn HandlerFunc, options ...DecodeOption) error

// Decode a stream of  documents of any Kind using either the innate typing of the scheme.
// Falls back to the unstructured.Unstructured type if a matching type cannot be found for the Kind.
// Options may be provided to configure the behavior of the decoder.
func DecodeAll(ctx context.Context, manifest io.Reader, options ...DecodeOption) ([]k8s.Object, error)
```

Using a typed object when decoding multiple documents does not provide for an easy-to-use interface, so they are not being proposed at this time.

### Pre-defined Decoders

Building on the proposal, the following functions would be included that build on the base decoders:

```go
// Decode the file at the given manifest path into the provided object. Patches are optional and applied after decoding.
func DecodeFile(fsys fs.FS, manifestPath string, obj k8s.Object, options ...DecodeOption) error

// Decode the manifest string into the provided object. Patches are optional and applied after decoding.
func DecodeString(rawManifest string, obj k8s.Object, options ...DecodeOption) error

// Decode the manifest bytes into the provided object. Patches are optional and applied after decoding.
func DecodeBytes(manifestBytes []byte, obj k8s.Object, options ...DecodeOption) error

// DecodeEachFile resolves files at the filesystem matching the pattern, decoding JSON or YAML files. Supports multi-document files.
//
// If handlerFn returns an error, decoding is halted.
// Options may be provided to configure the behavior of the decoder.
func DecodeEachFile(ctx context.Context, fsys fs.FS, pattern string, handlerFn HandlerFunc, options ...DecodeOption) error

// DecodeAllFiles  resolves files at the filesystem matching the pattern, decoding JSON or YAML files. Supports multi-document files.
// Falls back to the unstructured.Unstructured type if a matching type cannot be found for the Kind.
// Options may be provided to configure the behavior of the decoder.
func DecodeAllFiles(ctx context.Context, fsys fs.FS, pattern string, options ...DecodeOption) ([]k8s.Object, error)
```

### Pre-defined Helpers

```go
// CreateHandler returns a HandlerFunc that will create objects
func CreateHandler(r *resources.Resources, opts ...CreateOption) HandlerFunc
// UpdateHandler returns a HandlerFunc that will update objects
func UpdateHandler(r *resources.Resources, opts ...UpdateOption) HandlerFunc
// DeleteHandler returns a HandlerFunc that will delete objects
func DeleteHandler(r *resources.Resources, opts ...DeleteOption) HandlerFunc

// IgnoreErrorHandler returns a HandlerFunc that will ignore the provided error
func IgnoreErrorHandler(HandlerFunc, error) HandlerFunc

// CreateIfNotExistsHandler returns a HandlerFunc that will create objects if they do not already exist
func CreateIfNotExistsHandler(r *resources.Resources, opts ...CreateOption) HandlerFunc

// apply an override set of labels to a decoded object
func MutateLabels(overrides map[string]string) DecodeOption
// apply an override set of annotations to a decoded object
func MutateAnnotations(overrides map[string]string) DecodeOption
// apply an owner annotation to a decoded object
func MutateOwnerAnnotations(owner k8s.Object) DecodeOption
```
