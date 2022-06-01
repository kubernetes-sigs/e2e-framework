# Watcher for K8S Objects

K8S object watchers are great functionality provided by k8s to get efficient change notifications on resources.

The events supported by these watchers are
1. ADD
2. MODIFY/UPDATE
3. DELETE

## Motivation
Users must implement watchers such a way that whenever any events recieved some action/job has to be triggered. Developer has to write lots of code to do this and its sometimes difficult to manage. 

This can be achieved by using the k8s client built in function but understanding which packages to import or which core type needs to be used might be a complex for the developer.

The below design would make developer life easier. They have to just register their actions for respective events. To stay informed about when these events get triggered just use Watch(), which resides inside klient/k8s/resources package.

## Proposal
Watch function accepts a `object ObjectList` as a argument. ObjectList type is used to inject the resource type in which Watch has to be applied.

`klient/k8s/resources/resources.go`
```go=
import (
    "sigs.k8s.io/controller-runtime/pkg/client"
    "k8s.io/apimachinery/pkg/watch"
)

func (r *Resources) Watch(object k8s.ObjectList, opts ...ListOption) *watcher.EventHandlerFuncs {
	listOptions := &metav1.ListOptions{}

	for _, fn := range opts {
		fn(listOptions)
	}

	o := &cr.ListOptions{Raw: listOptions}

	return &watcher.EventHandlerFuncs{
		ListOptions: o,
		K8sObject:   object,
		Cfg:         r.GetConfig(),
	}
}
```

Watch() in resources.go will return the `watcher` type which helps to call `Start()`. InvokeEventHandler accepts `EventHandlerFuncs` which carries the user registerd function sets.

file : klient/k8s/resources/watch.go

```go=
// Start triggers the registered methods based on the event recieved for particular k8s resources.
func (watcher watch.Interface)Start(ctx context.Context) {
    ...
    go func() {
		for {
			select {
			case <-ctx.Done():
				if ctx.Err() != nil {
					return
				}
			case event := <-e.watcher.ResultChan():
				// retrieve the event type
				eventType := event.Type

				switch eventType {
				case watch.Added:
					// calls AddFunc if it's not nil.
					if e.addFunc != nil {
						e.addFunc(event.Object)
					}
				case watch.Modified:
					// calls UpdateFunc if it's not nil.
					if e.updateFunc != nil {
						e.updateFunc(event.Object)
					}
				case watch.Deleted:
					// calls DeleteFunc if it's not nil.
					if e.deleteFunc != nil {
						e.deleteFunc(event.Object)
					}
				}
			}
		}
	}()
    ...
}

// EventHandlerFuncs is an adaptor to let you easily specify as many or
// as few of functions to invoke while getting notification from watcher
type EventHandlerFuncs struct {
	addFunc     func(obj interface{})
	updateFunc  func(newObj interface{})
	deleteFunc  func(obj interface{})
	watcher     watch.Interface
	ListOptions *cr.ListOptions
	K8sObject   k8s.ObjectList
	Cfg         *rest.Config
}

// EventHandler can handle notifications for events that happen to a resource.
// Start will be waiting for the events notification which is responsible
// for invoking the registered user defined functions.
// Stop used to stop the watcher.
type EventHandler interface {
	Start(ctx context.Context)
	Stop()
}

```

`Start()` is invoked in a goroutine so that whenever watched resource changes the states it will call the registered user defined functions.
`Stop()` should be explicitly invoked by the user after the watch once the feature is done to ensure no unwanted go routine thread leackage.

If any error while Start() one can retry it for number of times.

This example shows how to use klient/resources/resources.go Watch() func and how to register the user defined functions.
 
```go=

import (
    "sigs.k8s.io/e2e-framework/klient/conf"
    "sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func main() {
    ...
    cfg, _ := conf.New(conf.ResolveKubeConfigFile())
    cl, err := cfg.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	dep := appsv1.Deployment{
    	ObjectMeta: metav1.ObjectMeta{Name: "watch-dep", Namespace: cfg.Namespace()},
	}

	// watch for the deployment and triger action based on the event recieved.
	cl.Resources().Watch(&appsv1.DeploymentList{}, resources.WithFieldSelector(labels.FormatLabels(map[string]string{"metadata.name": dep.Name}))).
	WithAddFunc(onAdd).WithDeleteFunc(onDelete).Start(ctx)

    ...
}

// onAdd is the function executed when the kubernetes watch notifies the
// presence of a new kubernetes deployment in the cluster
func onAdd(obj interface{}) {
    dep := obj.(*appsv1.Deployment)
    _, ok := dep.GetLabels()[K8S_LABEL_AWS_REGION]
    if ok {
        fmt.Printf("It has the label!")
    }
}

// onDelete is the function executed when the kubernetes watch notifies 
// delete event on deployment
func onDelete(obj interface{}) {
    dep := obj.(*appsv1.Deployment)
    _, ok := dep.GetLabels()[K8S_LABEL_AWS_REGION]
    if ok {
        fmt.Printf("It has the label!")
    }
}

```

The e2e flow of how to use watch is demonsrated in the examples/ folder.