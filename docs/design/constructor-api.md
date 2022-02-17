# `klient` - Object Constructor API for Kubernetes Workload Objects

Constructing a graph for Kubernetes objects require tedious steps of specifying each entry in the graph. Often, this forces and requires the code writer to have to look up how to assemble the pieces so that the graph can be constructed properly.

For instance, the following code snippet shows how to programmatically create an object graph for a simple `Deployment`. As you can see, it takes many entries to create a simple object structure to satisfy the object graph for `Deployment`.

> Example pulled from [here](https://sourcegraph.com/github.com/kubernetes/kubernetes/-/blob/test/e2e/framework/deployment/fixtures.go?L41).

```go
func NewDeployment() *appV1.Deployment {
    return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "my-deployment",
			Labels: podLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: podLabels},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsV1.RecreateDeploymentStrategyType,
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: v1.PodSpec{
					TerminationGracePeriodSeconds: &zero,
					Containers: []v1.Container{
						{
							Name:            containerName,
							Image:           image,
							SecurityContext: &v1.SecurityContext{},
						},
					},
				},
			},
		},
	}
}
```

## Motivation

This design proposes a functional approach where a developer would use function and method calls to automate the creation of Kubernetes object graphs. Each object would be built programmatically, using sensible defaults provided where necessary. Objects, in the object graph, would have an associated constructor function and type to create and manage its value.

For instance, given a set of constructor packages, such as `constructor/deployment`, `constructor/pod`, `constroctor/meta`, etc, the previous object graph could be created as follows:

```go

import (
    "sigs.k8s.io/e2e-framework/klient/constructors/deployment"
    "sigs.k8s.io/e2e-framework/klient/constructors/meta"
)

func NewDeployment() *appsV1.Deployment {
    myLabels := map[string]string{"type":"web"}
    return deployment.Deployment(
        meta.Object("my-deployment").Labels(myLabels),
        meta.MatchLabels(myLabels).MatchExpressions(myExpressions)
        deployment.Strategy(appsV1.RecreateDeploymentStrategyType),
        deployment.Replicas(2),
        pod.Template(
            meta.Object(meta.EmptyName).Labels(myLabels),
            pod.Spec(
               container.Name(containerName).Image(image).Args(container.Arg("key", "val")),
               container.Name(container.EmptyName).Image(image2).Command("/bin/sh -C date"),
            ),
        ),     
    )
}
```

## Design

In this design, each top-level workload API object would have a matching constructor package:

* `coreV1.MetaData` - `./constructor/meta`
* `*coreV1.Node` - `./constructor/node`
* `*coreV1.Pod` - `./constructor/pod`
* `*appsV1.Deployment` - `./constructor/deployment`
* `*appsV1.Service` - `./constructor/service`
* Etc

### Constructor types

Each top-level workload API type will be mapped to a constructor type as shown below.

* `coreV1.ObjectMeta` - `./constructor/meta.ObjectMetaConstructor`
* `*coreV1.Node` - `./constructor/node.NodeConstructor`
* `*coreV1.Pod` - `./constructor/pod.PodConstructor`
* `*appsV1.Deployment` - `./constructor/deployment.DeploymentConstructor`
* `*appsV1.Service` - `./constructor/service.ServiceConstructor`
* Etc

### Fluent API design

This design calls for a fluent-api, with successive method calls, to gradually build instances of the API object values. The fluent API style allows the method calls to be chained, with each method call inject a new value into the object.

* Each constructor type should use an underlying type that store values with each method call.
* The constructor type should expose fluent-api methods to collect those values.
* Each method call will return the constructor value so that successive calls may be chained.

Example: the following shows a snippet of fluent API style used to create a set of container objects:

```go
pod.Spec(
    container.Name(containerName).Image(image).Args(container.Arg("key", "val")),
    container.Name(container.EmptyName).Image(image2).Command("/bin/sh -C date"),
)
```
As mentioned above, the fluent API approach uses a series of chained method calls to construct values using initializer functions and finalizer methods.

#### Initializer functions

Each package will surface initializer functions, that serve as the starting point of the fluent call chain.  These functions are responsible for creating the initial value of the constructor type. The following snippet shows the use of initializer function `container.Name()` to create a value of type `pod.ContainerConstructor`

```go
func main() {

    c := container.Name(containerName).Image(image).Args(container.Arg("key", "val"))
    ...
}
```

In the previous snippet, function `Name` serves as an initializer and returns a value of type `*container.ContainerConstructor`.

> A package may have one or more initializer functions (either for the same type or different types).

An initializer function may also take one or more arguments as a way to streamline the fluent API. This approach is useful for types that can be constructed with three or fewer components. As a convenience, an initializer can be provided that accepts all of its needed values, in one shot, thus reducing the surface of the fluent API calls. For instance, a `PodConstructor` requires two values, a `MetaObjectConstructor` and a `PodSpecConstructor`, the `pod.Pod` initializer function accepts both values as shown below.

```go
func main() {
    myPod = pod.Pod(
        meta.Object(meta.EmptyName).Labels(myLabels),
        pod.Spec(
            container.Name(containerName).Image(image).Args(container.Arg("key", "val"))
            container.Name(container.EmptyName).Image(image2).Command("/bin/sh -C date"),
        ),
    )
    ...
}
```

#### Setter methods

Constructor types shall also provide setter methods to allow the storing of values, with each successive fluent method call, during the construction of the fluent API chain, as shown below.

```go
func main() {
    c := container.Name(containerName).Image(image).Args(container.Arg("key", "val")),
}
```

In the snippet above, method `Image` and `Args` are used to inject values into the initialized value of type `container.ContainerConstructor`.

#### Finalizer methods

Constructor types shall provide, at most, one finalizer method that is used to return the value of the underlying native server API object type. For instance, the following snippet shows the finalizer method `pod.PodConstructor.Build` used to return a value of type `coreV1.Pod`.

```go
func main() {
    myPod := pod.Pod(
        meta.Object(meta.EmptyName).Labels(myLabels),
        pod.Spec(
            container.Name(containerName).Image(image).Args(container.Arg("key", "val"))
            container.Name(container.EmptyName).Image(image2).Command("/bin/sh -C date"),
        ),
    ).Build()
    ...
}
```

The finalizer method shall:

* Be named `Build` when possible
* Return a value with the top-level native Kubernetes API type (i.e. appsV1.Deployment, coreV1.Pod, etc)

### Constructor type examples

This section shows the design examples for several constructor types.

### `meta.ObjectMetaConstructor`

The following shows how the fluent API for type `ObjectMetaConstructor` that is used to implement a constructor for native Kubernetes API type `metav1.ObjectMeta`.

```go
package "meta"

// ObjectMetaConstructor constructor for type coreV1.ObjectMeta
type ObjectMetaConstructor struct {
	obj metaV1.ObjectMeta
}

// Object is the initializer function for ObjectMetaConstructor
func Object(name string) ObjectMetaConstructor {
	return ObjectMetaConstructor{obj: metaV1.ObjectMeta{Name: name}}
}

// Namespace setter for namespace value
func (c ObjectMetaConstructor) Namespace(ns string) ObjectMetaConstructor {
	c.obj.Namespace = ns
	return c
}

// Labels setter for labels
func (c ObjectMetaConstructor) Labels(labels map[string]string) ObjectMetaConstructor {
	c.obj.Labels = labels
	return c
}

// Annotations setter for annotations
func (c ObjectMetaConstructor) Annotations(labels map[string]string) ObjectMetaConstructor {
	c.obj.Annotations = labels
	return c
}

// ClusterName setter for cluster name value
func (c ObjectMetaConstructor) ClusterName(name string) ObjectMetaConstructor {
	c.obj.ClusterName = name
	return c
}

// Build is the finalizer that builds and returns metaV1.ObjectMeta
func (c ObjectMetaConstructor) Build() metaV1.ObjectMeta {
	return c.obj
}
```

Usage:

```go
deployment.Deployment(
    meta.Object("my-deployment").Labels(map[string]string{"server-type":"web"}),
    ...
)
```

### `meta.LabelSelectorConstructor`

The `metal.LabelSelectorConstructor` is used to create values of type `metaV1.LabelSelector`

```go
package "meta"

type LabelSelectorConstructor struct{
	sel metaV1.LabelSelector
}

// MatchLabels initializer function for type LabelSelectorConstructor
func MatchLabels(labels map[string]string) LabelSelectorConstructor {
	return LabelSelectorConstructor{sel: metaV1.LabelSelector{MatchLabels: labels}}
}

// MatchExpressions initializer function for type LabelSelectorConstructor
func MatchExpressions(expressions...metaV1.LabelSelectorRequirement) LabelSelectorConstructor {
	return LabelSelectorConstructor{sel: metaV1.LabelSelector{MatchExpressions: expressions}}
}

// MatchLabels setter for map[string]string labels
func (c LabelSelectorConstructor) MatchLabels(labels map[string]string) LabelSelectorConstructor {
	c.sel.MatchLabels = labels
	return c
}

// MatchExpressions setter for metaV1.LabelSelectorRequirement
func (c LabelSelectorConstructor) MatchExpressions(expressions...metaV1.LabelSelectorRequirement) LabelSelectorConstructor {
	c.sel.MatchExpressions = expressions
	return c
}

// Build is the finalizer method that returns the built *metaV1.LabelSelector value
func (c LabelSelectorConstructor) Build() metaV1.LabelSelector {
	return c.sel
}
```

Usage:

```go
deployment.Deployment(
    meta.MatchLabels(myLabels).MatchExpressions(myExpressions)
    ...
)
```

### `pod.SpecConstructor`

Type `pod.SpecConstructor` is used to create value of type `coreV1.PodSpec`.

```go
package "pod"

type SpecConstructor struct {
	spec coreV1.PodSpec
}

// Spec initializer method for type PodSpecConstructor
func Spec(containerConstructors... container.Constructor) SpecConstructor {
    spec := SpecConstructor{spec: coreV1.PodSpec{}}
    ...
    return spec
}

// Build is the finalizer method that returns a value of type coreV1.PodSpec
func (c SpecConstructor) Build() coreV1.PodSpec {
	return c.spec
}
```

Usage:

```go
func main() {
    podSpec := pod.Spec(
        container.Name(containerName).Image(image).Args(container.Arg("key", "val"))
        container.Name(container.EmptyName).Image(image2).Command("/bin/sh -C date"),
    ).Build()
}
```

### `pod.Constructor`

`pod.Constructor` is a constructor type used to create `coreV1.Pod` values.

```go
package "pod"

type Constructor struct {
	pod coreV1.Pod
}

// Pod initializer function for type pod.Constructor
func Pod(metaConstructor meta.ObjectMetaConstructor, specConstructor SpecConstructor) Constructor {
	return Constructor{pod: coreV1.Pod{ObjectMeta: metaConstructor.Build(), Spec: specConstructor.Build()}}
}

func (c Constructor) Build() coreV1.Pod {
	return c.pod
}
```

Usage:

```go
func main() {
    myPod := pod.Pod(
        meta.Object(meta.EmptyName).Labels(myLabels),
        pod.Spec(
            container.Name(containerName).Image(image).Args(container.Arg("key", "val"))
            container.Name(container.EmptyName).Image(image2).Command("/bin/sh -C date"),
        ),
    ).Build()
    ...
}
```

### `pod.TemplateSpecConstructor`

`pod.TemplateSpecConstructor` is used to create new instance of `coreV1.PodTemplateSpec`.

```go
package "pod"

type TemplateSpecConstructor struct {
	spec coreV1.PodTemplateSpec
}

// Template initializer method for type TemplateSpecConstructor
func Template(metaConstructor meta.ObjectMetaConstructor, specConstructor SpecConstructor) TemplateSpecConstructor {
    return TemplateSpecConstructor{
        spec: coreV1.PodTemplateSpec{
            ObjectMeta: metaConstructor.Build(),
            Spec:       specConstructor.Build(),
        },
    }
}

func (c TemplateSpecConstructor) Build() coreV1.PodTemplateSpec {
    return c.spec
}
```

Usage:
```go

func main(){
    podTemp := pod.Template(
        meta.Object(meta.EmptyName).Labels(myLabels),
        pod.Spec(
            container.Name(containerName).Image(image).Args(container.Arg("key", "val"))
            container.Name(container.EmptyName).Image(image2).Command("/bin/sh -C date"),
        ),
    ),
}
```

### `deployment.Constructor`

Type `deployment.Constructor` surfaces methods to help construct a native Kubernetes API object of type `appV1.Deployment`.

```go
package "deployment"

import(
	"sigs.k8s.io/e2e-framework/klient/constructor/meta"
	"sigs.k8s.io/e2e-framework/klient/constructor/pod"
)

type Constructor struct {
	deployment appsV1.Deployment
}

func Deployment(
	deploymentMeta meta.ObjectMetaConstructor,
	replicas *int32,
	selector meta.LabelSelectorConstructor,
	strategy StrategyConstructor,
	template pod.TemplateSpecConstructor,
) Constructor {...}

func (c Constructor) Build() appsV1.Deployment{
	return c.deployment
}
```

Usage:

```go
func NewDeployment() *appV1.Deployment {
    dep := Deployment(
        meta.Object("test-deployment").Namespace(meta.DefaultNamespace),
        deployment.Replicas(2),
        meta.MatchLabels(map[string]string{"server-type": "web"}),
        StrategyDefault,
        pod.Template(
            meta.ObjectMetaNone, 
            pod.Spec(container.Name("server").Image("nginx").Commands("/start")),
        ),
    ).Build()
    return &dep
}
```
