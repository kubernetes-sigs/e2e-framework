# Watching for Resource Changes

The test harness supports several methods for querying Kubernetes object types and watching for resource states. This example shows how to watch particular resource and how to register the functions to act upon based on the events received.


# Watch for the deployment and triger action based on the event

Watch has to run as goroutine to get the different events based on the k8s resource state changes.
```go
func TestWatchForResources(t *testing.T) {
...
dep := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "watch-dep", Namespace: cfg.Namespace()},
			}

			// watch for the deployment and triger action based on the event received.
			cl.Resources().Watch(&appsv1.DeploymentList{}, &client.ListOptions{
				FieldSelector: fields.OneTermEqualSelector("metadata.name", dep.Name),
				Namespace:     dep.Namespace}, cl.RESTConfig()).WithAddFunc(onAdd).WithDeleteFunc(onDelete).Start(ctx)
...
}
```

# Function/Action definition and registering these actions

```go
// onAdd is the function executed when the kubernetes watch notifies the
// presence of a new kubernetes deployment in the cluster
func onAdd(obj interface{}) {
	dep := obj.(*appsv1.Deployment)
	depName := dep.GetName()
	fmt.Printf("Dep name received is %s", depName)
	if depName == "watch-dep" {
		fmt.Println("Dep name matches with actual name!")
	}
}

// onDelete is the function executed when the kubernetes watch notifies
// delete event on deployment
func onDelete(obj interface{}) {
	dep := obj.(*appsv1.Deployment)
	depName := dep.GetName()
	if depName == "watch-dep" {
		fmt.Println("Deployment deleted successfully!")
	}
}
```

The above functions can be registered using Register functions(WithAddFunc(), WithDeleteFunc(), WithUpdateFunc()) defined under klient/k8s/watcher/watch.go as shown in the example.

# How to stop the watcher
Create a global EventHandlerFuncs variable to store the watcher object and call Stop() as shown in example TestWatchForResourcesWithStop() test.

Note: User should explicitly invoke the Stop() after the watch once the feature is done to ensure no unwanted go routine thread leackage.