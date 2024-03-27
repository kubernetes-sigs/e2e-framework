# `klient` - A Client Package for Kubernetes Clusters

This document proposes the design for klient, a Go package that is intended to abstract the complexities of project `client-go` and provide helper functionalities to make it dead simple to interact with clusters and their API-servers. The goal of `klient` is to provide an easy programmable interface that covers the creation and control of API resources and the underlying compute infrastructure where Kubernetes clusters are running.

> See original Google Doc design document [here](https://docs.google.com/document/d/1ZtN58kU8SKmDDkxeBKxR9Un76eqhszCVcZuqhs-fLIU/edit?usp=sharing).

## Motivation

While the `client-go` is a formidable project that comes with batteries included, it was designed for building Kubernetes resource controllers by exposing low-level constructs that require intimate knowledge of the internal workings of Kubernetes reconciliation mechanism. However, for developers who are looking for a simple API surface to interact with the API-server and its resources, `client-go` can feel clunky and overwhelming.

Another reason for this proposal is because it is part of a larger effort to create an upstream Go framework for end-to-end testing of components running in Kubernetes (e2e-framework). A large part of exercising resources during testing is to have the ability to interact with the cluster where those resources are deployed. Having a simple client that provides simple and consistent abstractions will help test writers quickly create tests that include Kubernetes resource CRUD (create, update, delete) operations and cluster control operations.

## Goals

- Simplify the access to Kubernetes cluster configuration
- Provide a uniform API to work with typed and untyped Kubernetes API objects
- Surface types that make it easy to create, update, and delete API objects
- Provide types to easily control Kubernetes clusters

## Non-Goals

A replacement for `client-go`
A control-loop framework to build Kubernetes resource controllers
Abstraction of the controller-runtime project

## Design

The overall design of klient will be to provide functionalities in following categories:

- ***Configuration***
  - Access cluster configuration (KubeConfig)
- ***Resource*** - this represents any resource (custom or otherwise) registered with the API server.
  - Resource creation and management
  - Resource search and list
  - Resource deletion
- ***Control***
  - Cluster control
  - Application deployment
  - Storage
  - Network
  - Nodes
- ***Infrastructure***
  - Local process execution
  - Remote process execution
  - Pod-hosted processes

### Resource representation

Users of this framework will be spared the burden of figuring out which Kubernetes object representation to use, mainly, typed or unstructured. To do this, klient uses the following types to wrap API objects and object lists. This is based on a model found in the runtime-controller project.

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


// Object is a union type that can represent either typed objects
// of type metav1.Object or dynamic objects of type runtime.object.
type Object interface {
    metav1.Object
    runtime.Object
}

type ObjectList interface {
    metav1.ListInterface
    runtime.Object
}
```

### Specifying optional parameters

Most methods used in this design will use an `Option` type (defined as a function) that will allow framework users to specify optional method arguments (as a variable list of arguments). For instance, the following shows how method `List` could receive optional arguments using type `ListOption`:

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListOptions struct{ ... }
type ListOptFunc func(*ListOptions)

func (_ Resource) List(ctx context.Context, namespace string, l ObjectList, opts ...ListOption) error
```

Optional parameters can be omitted from the call (as shown below), in which case the framework would use sensible default values where applicable.

```go
func main() {
    var deps v1.DeploymentList
    if res.List(
        context.TODO(),
        "default", &deps,
        func(opts *metav1.ListOptions){opts.Labels="tier=web"}); err != nil {
        log.Fatal("unable to list deployments ", err)
    }
}
```

`klient` should define ready-made convenience functions, for commonly used optional arguments, to facilitate usage of this pattern. For example, below you can see the same example using a predefined function that sets the label values.

```go
func main() {
    var deps v1.DeploymentList
    if res.List(
        context.TODO(),
        "default", &deps,
        WithLabelSelector("tier=web")); err != nil {
        log.Fatal("unable to list deployments ", err)
    }
}
```

### Automatic retries

Often, when issuing operations against resources in distributed infrastructure running Kubernetes, it is desirable to be able to automatically retry operations, for an amount of time, before giving up permanently on the call. Klient should include an automatic retry mechanism, where possible, to continuously attempt to make calls to Kubernetes or infrastructure resources for specified amount or time or until its timeout.

For instance, if we assume that type `ListOptions` includes a `RetryTimeout` field, the following example shows how klient could specify an optional argument to cause the call to the API server to be retried for a given amount of time.

```go
func main() {
    var deps v1.DeploymentList
    if res.List(
        context.TODO(),
        "default", &deps,
        func(opts *ListOptions){opts.Labels="tier=web"}
        func(opts *ListOptions){opts.RetryTimeout=time.Seconds*30}); err != nil {
        log.Fatal("unable to list deployments ", err)
    }
}
```

## Kubernetes Cluster Configuration

> Package: `./klient/conf`

This package helps initialize Kubernetes configuration of type `*rest.Client` to be used when connecting with the API-server.

### Configuration creator functions

The package should provide creator functions to create a new instance of `*rest.Config` as shown below:

```go
import (
    "k8s.io/client-go/rest"
)

// New returns *rest.Config based on the kubeconfig file name.
func New(fileName string) (*rest.Config, error){...}

// NewWithContextName returns *rest.Config using the kubeconfig and the
// kubeconfig context provided
func NewWithContextName(fileName, context string) (*rest.Config, error)

// NewInContext assumes the code is running within a cluster and returns
// the associated *rest.Config
func NewInCluster() (*rest.Config, error)

```

Additionally, package `conf` should provide helper functions that help create the config.

```go
// ResolveKubeConfigFile returns the kubeconfig file from
// either flag --kubeconfig or env KUBECONFIG.
// It will only call flag.Parse() if flag.Parsed() is false.
// If --kubeconfig, or KUBECONFIG, or  $HOME/.kube/config not provided then
// assume in cluster.
func ResolveKubeConfigFile() string{...}

// ResolveContext returns cluster context name based on --context flag.
func ResolveClusterContext() string{...}
```

The intent is for these convenient helper functions to be used when creating a cluster configuration instance whether the client code is running inside or outside of the cluster itself.

- Function `ResolveKubeConfigFile` provides logic to determine the path of a default Kubernetes configuration file to use.
- Function `ResolveClusterContext` attempts to read a provided cluster context name from a command-line argument.

#### Example

```go
func main() {
    cfg, err := New(ResolveKubeConfigFile())
}
```

## The k8s package
>
> Package: `./klient/k8s`

This package (and its subpackages) represents the root package that surfaces several types which provide functionalities to interact with the Kubernetes API server and its stored resources.

### Package `resources`
>
> Package: `./klient/k8s/resources`

This package contains types and operations for creating and interacting with Kubernetes resources.

Type `resources.Resources` is the root type that provides access to resource-related operations that are implemented in the package.

```go
type Resources struct {}
```

#### Constructor function

```go
func New(cfg *rest.Config) *Resources {...}
```

#### Example - creating a resource

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)
func main() {
    cfg, err := conf.New(conf.ResolveKubeConfigFile())
    if err != nil {
        log.Fatal(err)
    }
    res, err := resources.New(cfg)
    if err != nil {...}
    ...
}
```

### Method `Resources.Search`

Method Search allows for the search of arbitrary API resources that match provided arguments. The following snippets outline the types used to construct an object search.

```go
type SearchOptions struct {
    Groups     []string
    Categories []string
    Kinds      []string
    Namespaces []string
    Versions   []string
    Names      []string
    Labels     []string
    Containers []string
}

// SearchOption type to update SearchOptions
func SearchOption func(*SearchOptions)

type SearchResult struct {
    Options SearchOptions
    List ObjectList
}

// Search searches for API objects using a combination of search criteria
// passed in SearchOption
func (c *Resources) Search(ctx context.Context, opts... SearchOption)(SearchResult,error){...}
```

#### Example

The following shows an example that retrieves all pods whose names starts with `"Dns"` from namespace `"default"` or `"net-svc"`.

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, err := resources.New(cfg)
    if err != nil {...}

    result, err := res.Search(context.TODO(), func(opts SearchOptions){
        opts.Namespaces = []string{"default", "net-svc"},
        opts.Kinds = []string{"pods"}
        opts.Names = []string{"Dns"}
    })

    if len(result.List.Items) == 0 {
       fmt.Println("no objects found")
    }
}
```

### Method `Resources.Get`

Method `Resources.Get` retrieves a specific object instance based on name.

```go
// Get retrieves a specific API object based on specified `name` and type of parameter `obj`
func (c *Resources) Get(ctx context.Context, name, namespace string, obj k8s.Object) error
```

#### Example

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, err := resources.New(cfg)
    if err != nil {...}

    var pod v1.Pod
    if err := res.Get(context.TODO(), "mypod", "default", &pod); err != nil {
        log.Fatal("unable to retrieve pod: ", err)
    }
}
```

### Method `Resources.List`

Method `Resources.List` retrieves a list of API objects of a given type.

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListOption used to configure List method call
type ListOption func(*metav1.ListOptions)

// List retrieves a list of API objects with the same type as specified by `objs`
func (c *Resources) List(ctx context.Context, namespace string, objs k8s.ObjectList, opts ...k8s.ListOption) error
```

#### Example

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, err := resources.New(cfg)
    if err != nil {...}

    var deps v1.DeploymentList
    if err := res.List(
        context.TODO(),
        "default", &deps,
        resources.WithLabelSelector("tier=web"); err != nil {
        log.Fatal("unable to list deployments ", err)
    }
}
```

#### Possible convenience functions

Note in the above example, the `List` method uses function `resources.WithLabelSelector` to cleanly specify a label selector for the call. This could be done with pre-defined convenience functions as listed below:

```go
package "resources"
func WithLabelSelector(sel string) ListOption{}
func WithFieldSelector(sel string) ListOption{}
func WithTimeout(to time.Duration) ListOption{}
...
```

### Method `Resource.Create`

Method `Resource.Create` creates and stores a new object on the API server.

```go
package "resources"
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOption used to specify optional parameters for Create call
type CreateOption func(*metav1.CreateOptions)

// Create sends API object of type specified by `obj`
func (c *Resources) Create(ctx context.Context, obj k8s.Object, opts...k8s.CreateOption)
```

#### Example

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, err := resources.New(cfg)
    if err != nil {...}

    pod := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: podName,
        },
        Spec: v1.PodSpec{
            Containers: []v1.Container{{
                Name:  "nginx",
                Image: imageutils.GetPauseImageName(),
            }},
        },
    }

    if err := res.Create(context.TODO(), &pod); err != nil {
        log.Fatal("unable to create pod: ", err)
    }
}
```

#### Object constructor functions

The`resources` package could include helper functions to help construct common object resources such as pods, deployment, services etc. For instance, the previous could be rewritten as follows:

```go
func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, _ := resources.New(cfg)

    pod := resources.Pod(name, resources.PodSpec(resources.Container("nginx", "nginx:latest")))
    if err := res.Create(context.TODO(), &pod); err != nil {
        log.Fatal("unable to create pod: ", err)
    }
}
```

### Method `Resources.Update`

The `Resources.Update` method updates an existing cluster object.

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UpdateOption is used to specify optional configuration values for the Update call
type UpdateOption func(*metav1.UpdateOptions)

// Update updates an API object with the type specified by `obj`
func (c *Resources) Update(ctx context.Context, obj Object, opts...UpdateOption)
```

#### Example

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, err := resources.New(cfg)
    if err != nil {...}

    pod := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: podName,
        },
        Spec: v1.PodSpec{
            Containers: []v1.Container{{
                Name:  "nginx",
                Image: imageutils.GetPauseImageName(),
            }},
        },
    }
    if err := res.Update(context.TODO(), &pod); err != nil {
        log.Fatal("unable to update pod: ", err)
    }
}
```

### Method `Resources.Delete`

Method `Resources.Delete` deletes an existing API object.

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeleteOption is used to specify optional arguments for the Delete operation
type DeleteOption func(*metav1.DeleteOptions)

// Delete deletes an api object that matches the type of `obj` and its metadata
func (c *Resources) Delete(ctx context.Context, obj Object, opts...UpdateOption)
```

#### Example

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, err := resources.New(cfg)
    if err != nil {...}

    pod := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: podName,
        }
    }
    if err := res.Update(
        context.TODO(),
        &pod,
        func(opt *metav1.DeleteOptions){opt.GracePeriodSeconds=30}); err != nil {
        log.Fatal("unable to update pod: ", err)
    }
}
```

#### Possible convenience functions

Package `resources` could include helper functions to help specify delete options as shown below:

```go
package "resources"

func WithGracePeriod(sel int64) DeleteOption
func WithDeletePropagation(prop v1.DeletePropagation) ListOption
```

### Method `Resources.Patch`

This method facilitates patching portions of an existing object with new data from another object of the same type.

```go
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)

// PatchOption is used to provide additional arguments to the Patch call.
type PatchOption func(*metav1.PatchOptions)

// Patch patches portion of object `orig` with data from object `patch`
func (c *Resources) Patch(ctx context.Context, orig Object, patch Object, opts...PatchOption)
```

#### Example

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, err := resources.New(cfg)
    if err != nil {...}

    pod1 := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: podName,
        }
    }
    mergePatch, err := json.Marshal(map[string]interface{}{
        "metadata": map[string]interface{}{
                "annotations": map[string]interface{}{
                    "foo": "bar",
                 },
            },
        })

    patch := k8s.Patch{PatchType: types.StrategicMergePatchType, Data: mergePatch}
    if err := res.Patch(context.TODO(), &pod1, patch); err != nil {
        log.Fatal("unable to update pod: ", err)
    }
}
```

### Other Resources method

The following are other resource methods that should be considered for this design.

#### Method `Resources.Annotate`

This method could be used to attach annotations to an existing resource object.

#### `Resources.Label` method

Method `Label` could be used to apply labels to an existing resources.

#### Method `Resources.BuildFromJSON`

This method encodes a `JSON` string as a typed (or unstructured) object representation.

```go
func (r *Resources) BuildFromJSON(ctx context.Context, obj kclient.Object, json string) error
```

#### Example

```go
import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    res, err := resources.New(cfg)
    if err != nil {...}

    var deps v1.Pod
    json := `{"kind":"pod", objectMeta": {"name":"podName"}}`
    if err := res.BuildFromJSON(context.TODO(), &pod, json); err != nil {
        log.Fatal("unable to encode pod ", err)
    }
}
```
