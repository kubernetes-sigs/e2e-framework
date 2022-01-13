# Decoding Kubernetes Objects

This package demonstrates how to use the `klient/decoder` package in coordination with the test framework to use YAML or JSON representations of `k8s.Object` from any `io.Reader` or `fs.FS` compatible source to create resources to use in tests.

The decoder package supports decoding YAML or JSON encoded Kubernetes objects from files, strings, and byte slices (any `io.Readder`).

`MutateFunc` and `HandlerFunc` variations allow for various workflows:
- Load a set of YAML or JSON files to create objects in a feature Setup
- Delete a set of Kubernetes objects in a feature Teardown
- Load and patch a set of resources, injecting a dynamic namespace field
- Apply resource changes based on easily edited files
- Create/Delete resources in a `testenv.Setup` or `testenv.Teardown` function

## Decoding in a Setup Function

Kubernetes objects may be represented as a string for convenience:

```go
var initYAML string = `
apiVersion: v1
kind: Namespace
metadata:
  name: mytest-namespace
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mysa
  namespace: mytest-namespace
`
```

Multi-document decoding support allows multiple Kubernetes objects to be decoded from one io.Reader, as represented above.

To create a set of objects from a YAML document stream, before tests are run, call the decoder
in a testenv.Setup `env.Func`, passing it a `decoder.CreateHandler` to handle creation:
```go
    testenv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				return ctx, err
			}
			// decode and create a stream of YAML or JSON documents from an io.Reader
			decoder.DecodeEach(ctx, strings.NewReader(initYAML), decoder.CreateHandler(r))
			return ctx, nil
		},
	)
```

## Decoding Objects

To decode an object from an `io.Reader`, the `decoder` package offers many options.

The simplest options include:

Decoding from a single file:

```go
    ...
    .Assess("Single File", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		f, err := os.Open("testdata/configmap.yaml")
		if err != nil {
			t.Fatal(err)
		}
		obj, err := decoder.DecodeAny(f)
		if err != nil {
			t.Fatal(err)
		}
		configMap, ok := obj.(*v1.ConfigMap)
		if !ok {
			t.Fatal("object decoded to unexpected type")
		}
		t.Log(configMap)
		return ctx
	})
    ...
```

Or, decoding a set of files:

```go
    testdata := os.DirFS("testdata")
	pattern := "*"
    ...
    .Assess("Multiple Files", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		objects, err := decoder.DecodeAllFiles(ctx, testdata, pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, obj := range objects {
			t.Log(obj.GetObjectKind(), obj.GetNamespace(), obj.GetName())
		}
		return ctx
	})
    ...
```

## Decoding with HandlerFuncs

Some decoder functions accept a `HandlerFunc`, which specify a function to execute with the decoded object after decoding and any defined MutateFunc options are applied.

The decoder package includes a number of built-in HandlerFuncs that allow for basic CRUD (Create, Update, Read, and Delete) operations:

```go
// CreateHandler returns a HandlerFunc that will create objects
func CreateHandler(r *resources.Resources, opts ...resources.CreateOption) HandlerFunc

// ReadHandler returns a HandlerFunc that will use the provided object's Kind / Namespace / Name to retrieve
// the current state of the object using the provided Resource client.
// This helper makes it easy to use a stale reference to an object to retrieve its current version.
func ReadHandler(r *resources.Resources, handler HandlerFunc) HandlerFunc

// UpdateHandler returns a HandlerFunc that will update objects
func UpdateHandler(r *resources.Resources, opts ...resources.UpdateOption) HandlerFunc

// DeleteHandler returns a HandlerFunc that will delete objects
func DeleteHandler(r *resources.Resources, opts ...resources.DeleteOption) HandlerFunc
```

A common pattern in tests is to create a set of resources in a `Setup` function, and consequently delete them in a Teardown:

```go
// use files matching the Glob testdata/* in the following tests
testdata := os.DirFS("testdata")
pattern := "*"
features.New("setup and teardown").
    Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
        r, err := resources.New(cfg.Client().RESTConfig())
        if err != nil {
            t.Fatal(err)
        }

        if err := decoder.DecodeEachFile(ctx, testdata, pattern,
            decoder.CreateHandler(r),           // try to CREATE objects after decoding
            decoder.MutateNamespace(namespace), // inject a namespace into decoded objects, before calling CreateHandler
        ); err != nil {
            t.Fatal(err)
        }
        return ctx
    }).
    ... // An assessment function would be able to use or test the created resources
    Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
        // remove test resources before exiting
        r, err := resources.New(cfg.Client().RESTConfig())
        if err != nil {
            t.Fatal(err)
        }
        if err := decoder.DecodeEachFile(ctx, testdata, pattern,
            decoder.DeleteHandler(r),           // try to DELETE objects after decoding
            decoder.MutateNamespace(namespace), // inject a namespace into decoded objects, before calling DeleteHandler
        ); err != nil {
            t.Fatal(err)
        }
        return ctx
    }).Feature()
```

The `decoder.MutateNamespace(namespace)`  DecodeOption injects the dynamically generated namespace into the decoded objects before it tries to create or delete them from the test cluster.

The decoder package includes a number of built-in MutateFunc DecodeOptions to perform common operations:

```go
// MutateLabels is an optional parameter to decoding functions that will patch an objects metadata.labels
func MutateLabels(overrides map[string]string) DecodeOption

// MutateAnnotations is an optional parameter to decoding functions that will patch an objects metadata.annotations
func MutateAnnotations(overrides map[string]string) DecodeOption

// MutateOwnerAnnotations is an optional parameter to decoding functions that will patch objects using the given owner object
func MutateOwnerAnnotations(owner k8s.Object) DecodeOption

// MutateNamespace is an optional parameter to decoding functions that will patch objects with the given namespace name
func MutateNamespace(namespace string) DecodeOption
```
