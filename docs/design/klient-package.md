# `klient` - Convenient Client Package for Kubernetes Clusters

This document proposes the design for klient, a collection of Go packages and types that are intended to provide helper client functionalities to interact with clusters and their API-servers.  The goal of these helper packages is to provide an easy programmable interface to access to interact with resources running on a Kubernetes cluster and its underlying infrastructure.

> See original Google Doc design document [here](https://docs.google.com/document/d/1ZtN58kU8SKmDDkxeBKxR9Un76eqhszCVcZuqhs-fLIU/edit?usp=sharing).
## Motivation
While client-go is a formidable, batteries included, library that has been serving developers well, it was designed for building Kubernetes resource controllers by exposing low-level constructs that require intimate knowledge of the internal workings of Kubernetes reconciliation mechanism. However, for developers who are looking for a simple API surface to interact with the API-server and its resources, client-go can feel clunky and overwhelming.

Another reason for this proposal is because it is part of a larger effort to create an upstream Go framework for end-to-end testing of components running in Kubernetes (e2e-framework). A large part of exercising resources during testing is to have the ability to interact with the cluster where those resources are deployed. Having a collection of helper packages would provide a consistent abstraction to help test writers with Kubernetes resource CRUD (create, update, delete) operations, support for cluster control operations, and seamless operation retries.

## Goals

* Package and types for a high level API to interact with the API server
* Simplify the access to Kubernetes cluster configuration
* Provide a uniform API to work with typed and untyped Kubernetes API objects
* Surface types that make it easy to create, update, and delete API objects
* Types to easily create control operations of Kubernetes clusters

## Non-Goals
A replacement for client-go
A control-loop framework to build Kubernetes resource controllers
Abstraction of the controller-runtime project

## Design
The overall design of klient will be to provide functionalities in following categories:

* ***Configuration***
    * Access cluster configuration (KubeConfig)
* ***Resource*** - this represents any resource (custom or otherwise) registered with the API server.
    * Resource creation and management
    * Resource search and list
    * Resource deletion
* ***Control***
    * Cluster control
    * Application deployment
    * Storage
    * Network
    * Nodes
* ***Infrastructure***
    * Local process execution
    * Remote process execution
        * Pod-hosted processes

### Resource representation
Users of this framework will be spared the burden of figuring out which Kubernetes object representation to use, mainly, typed or unstructured. To do this, klient uses the following types to wrap API objects and object lists. This is based on a model found in the runtime-controller project.

```go=
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
Most methods used in this design will use an option type (defined as a function) that will allow framework users to specify optional method arguments (as a variable list of arguments). For instance, the following shows how method List could receive optional arguments using type ListOption:

```go=
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)
type ListOptions struct{ ... }
type ListOptFunc func(*ListOptions)

func (_ Res) List(ctx context.Context, namespace string, l ObjectList, opts ...ListOption) error
```

Optional parameters can be omitted from the call (as shown below), in which case the framework would use sensible default values where applicable.

```go=
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

```go=
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
Often, when issuing operations against resources in distributed infrastructure running Kubernetes,  it is desirable to be able to automatically retry for an amount of time before giving up permanently on the call. Klient should include an automatic retry mechanism, where possible, to continuously attempt to make calls to Kubernetes or infrastructure resources for specified amount or time or until its timeout.

For instance, if we assume that type ListOptions includes a RetryTimeout field, the  following example shows how klient could specify an optional argument to cause the call to the API server to be retried for a given amount of time.

```go=
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

### Kubernetes Cluster Configuration

> Package: `./klient/conf`

The following functions can be used to create a new instance of *rest.Config to represent a Kubernetes cluster configuration which can be used as arguments when using other types found in the remainder of this document.

```go=
import (
    "k8s.io/client-go/rest"
)

var (
    // DefaultClusterContext value for default context name
    DefaultClusterContext = ""
)
// NewKubeConfig returns Kubernetes configuration value of type *rest.Config. 
func NewRESTConfig(fileName, context string)(*rest.Config, error){...}

// ResolveKubeConfigFile returns the kubeconfig file from
// either flag --kubeconfig or env KUBECONFIG. 
// It will only call flag.Parse() if flag.Parsed() is false.
// If --kubeconfig, or KUBECONFIG, or  $HOME/.kube/config not provided then
// assume in cluster.
func ResolveKubeConfigFile() string{...}

// ResolveContext returns cluster context name based on --context flag.
func ResolveClusterContext() string{...}
```

The intent is that `NewRESTConfig` shall be all that is needed to obtain a cluster configuration instance whether the client code is running inside or outside of the cluster itself.

* Function `ResolveKubeConfigFile` provides logic to determine the path of a default Kubernetes configuration file to use.
* Function `ResolveClusterContext` attempts to read a provided cluster context name from a command-line argument.

#### Example

```go
func main() {
    cfg, err := klient.NewRESTConfig(ResolveKubeConfigFile(), DefaultClusterContext)
}
```

## The k8s package
> Package: `./klient/k8s`

This package represents the root package for an aggregated client that surfaces many types to interact with the Kubernetes API server and its stored resources.

### Package `res`
> Package: `./klient/k8s/res`
This package contains types and operations for creating and interacting with Kubernetes resources.

#### Type k8s.Resources
> Package: ./klient/k8s

Type `k8s.Resources` is a façade that exposes resource-related operations to create, retrieve, list, and update Kubernetes cluster resources.

```go
type Resources struct {}
```

#### Constructor function
```go
func Res(cfg *rest.Config) *Resources {...}
```

#### Example
```go
func main() {
    cfg, err := klient.NewRESTConfig("path/to/kubecfg", DefaultClusterContext)
    if err != nil {
        log.Fatal(err)
    }
    res := k8s.Res(cfg)
}
```
### Method `Resources.Search`
Method Search allows for the search of arbitrary API resources that match provided arguments.  The following snippets outline the types used to construct an object search.

```go=
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
The following shows an example that retrieves all pods whose names starts with `"Dns"` from
namespace `"default"` or `"net-svc"`.

```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", k8s.DefaultClusterContext)
    result, err := k8s.Res(cfg).Search(context.TODO(), func(opts SearchOptions){
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
Method Resources.Get retrieves a specific object instance based on name.

```go
func (c *Resources) Get(ctx context.Context, name, namespace string, obj k8s.Object) error
```

#### Example
```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", k8s.DefaultClusterContext)
    var pod v1.Pod
    if err := k8s.Res(cfg).Get(context.TODO(), "mypod", "default", &pod); err != nil {
        log.Fatal("unable to retrieve pod: ", err)   
    }
}
```

### Method `Resources.List`
Method Resources.List retrieves a list of API objects of a given type.

```go=
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ListOption func(*metav1.ListOptions)

func (c *Resources) List(ctx context.Context, namespace string, objs k8s.ObjectList, opts ...k8s.ListOption) error
```

#### Example

```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", DefaultClusterContext)
    var deps v1.DeploymentList
    if err := k8s.Res(cfg).List(
        context.TODO(), 
        "default", &deps, 
        k8s.WithLabelSelector("tier=web"); err != nil {
        log.Fatal("unable to list deployments ", err)   
    }
}
```

#### Possible convenience functions

```go=
func WithLabelSelector(sel string) ListOption{}
func WithFieldSelector(sel string) ListOption{}
func WithTimeout(to time.Duration) ListOption{}
...
```

### Method `Resource.Create`
Method Resource.Create creates and stores a new object on the API server.

```go=
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateOption func(*metav1.CreateOptions)

func (c *Resources) Create(ctx context.Context, obj k8s.Object, opts...k8s.CreateOption)
```

#### Example

```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", DefaultClusterContext)
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
    if err := k8s.Res(cfg).Create(context.TODO(), &pod); err != nil {
        log.Fatal("unable to create pod: ", err)   
    }
}
```

#### Possible helper functions
The k8s package could include helper functions to help with the creation of common objects such as pods, deployment, services etc. For instance, the previous could be rewritten as follows:

```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", DefaultClusterContext)
    pod := k8s.Pod(name, k8s.PodSpec(k8s.Container("nginx", "nginx:latest")))
    if err := k8s.Res(cfg).Create(context.TODO(), &pod); err != nil {
        log.Fatal("unable to create pod: ", err)   
    }
}
```

### Method `Resources.Update`
The Resources.Update method updates an existing cluster object.

```go=
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpdateOption func(*metav1.UpdateOptions)

func (c *Resources) Update(ctx context.Context, obj Object, opts...UpdateOption)
```

#### Example

```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", DefaultClusterContext)
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
    if err := k8s.Res(cfg).Update(context.TODO(), &pod); err != nil {
        log.Fatal("unable to update pod: ", err)   
    }
}
```

### Method `Resources.Delete`
Method Resources.Delete deletes an existing API object.

```go=
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeleteOption func(*metav1.DeleteOptions)

func (c *Resources) DeleteObject(ctx context.Context, obj Object, opts...UpdateOption)
```

#### Example

```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", DefaultClusterContext)
    pod := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: podName,
        }
    }
    if err := k8s.Res(cfg).Update(
        context.TODO(), 
        &pod, 
        func(opt *metav1.DeleteOptions){opt.GracePeriodSeconds=30}); err != nil {
        log.Fatal("unable to update pod: ", err)   
    }
}
```

#### Possible convenience functions

```go=
func WithGracePeriod(sel int64) DeleteOption
func WithDeletePropagation(prop v1.DeletePropagation) ListOption
```

### Method `Resources.Patch`
This method facilitates patching portions of an existing object with new data from another object of the same type.

```go=
import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)

type PatchOption func(*metav1.PatchOptions)

func (c *Resources) Patch(ctx context.Context, orig Object, patch Object, patchType types.PatchType, opts...PatchOption)
```

#### Example

```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", DefaultClusterContext)
    pod1 := &v1.Pod{
        ObjectMeta: metav1.ObjectMeta{
            Name: podName,
        }
    }
    pod2 := &v1.Pod{
        Spec: v1.PodSpec{
            Containers: []v1.Container{{
                Name:  "nginx",
                Image: imageutils.GetPauseImageName(),
            }},
        },
    }
    if err := K8s.Res(cfg).Patch(context.TODO(), &pod1, &pod2); err != nil {
        log.Fatal("unable to update pod: ", err)   
    }
}
```

### Method `Resources.FromJSON`
This method encodes a JSON string as a typed (or unstructured) object representation.

```go
func (r *Resources) EncodeJSON(ctx context.Context, obj kclient.Object, json string) error
```

#### Example

```go=
func main() {
    cfg, _ := klient.NewRESTConfig("path/to/kubecfg", DefaultClusterContext)
    var deps v1.Pod
    json :=
    if err := k8s.Res(cfg).FromJSON(context.TODO(), &pod,
     `{"objectMeta": {"name":"podName"}}`,
    ); err != nil {
        log.Fatal("unable to encode pod ", err)   
    }
}
```