# Waiting for Resource Changes

The test harness supports several methods for querying Kubernetes object types and waiting for conditions to be met. This example shows how to create various wait conditions to drive your tests.

## Waiting for a single object

The wait package has built-in with utilities for waiting on Pods, Jobs, and Deployments:

```go
func TestPodRunning(t *testing.T) {
	var err error
	pod := v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "my-pod"}}
	err = wait.For(conditions.New(client.Resources()).PodRunning(pod), WithImmediate())
	if err != nil {
		t.Error(err)
	}
}
```

Additionally, it is easy to wait for changes to any resource type with the `ResourceMatch` method:

```go
func TestResourceMatch(t *testing.T) {
	...
	deployment := appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deploy-name"}}
	err = wait.For(conditions.New(client.Resources()).ResourceMatch(deployment, func(object k8s.Object) bool {
		d := object.(*appsv1.Deployment)
		return d.Status.AvailableReplicas == 2 && d.Status.ReadyReplicas == 2
	}))
	if err != nil {
		t.Error(err)
	}
	...
}
```

## Waiting for a lists of objects

It is common to need to check for the existence of a set of objects by name:

```go
func TestResourcesFound(t *testing.T) {
	...
	pods := &v1.PodList{
		Items: []v1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "p9", Namespace: namespace}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p10", Namespace: namespace}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p11", Namespace: namespace}},
		},
	}
	// wait for the set of pods to exist
	err = wait.For(conditions.New(client.Resources()).ResourcesFound(pods))
	if err != nil {
		t.Error(err)
	}
	...
}
```

Or to check for their absence:

```go
func TestResourcesDeleted(t *testing.T) {
	...
	pods := &v1.PodList{}
	// wait for 1 pod with the label `"app": "d5"`
	err = wait.For(conditions.New(client.Resources()).ResourceListN(
		pods,
		1,
		resources.WithLabelSelector(labels.FormatLabels(map[string]string{"app": "d5"}))),
	)
	if err != nil {
		t.Error(err)
	}
	err = client.Resources().Delete(context.Background(), deployment)
	if err != nil {
		t.Error(err)
	}
	// wait for the set of pods to finish deleting
	err = wait.For(conditions.New(client.Resources()).ResourcesDeleted(pods))
	if err != nil {
		t.Error(err)
	}
	...
}
```
