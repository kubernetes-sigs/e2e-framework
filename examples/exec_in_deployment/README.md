# Execute Commands in a Deployment

This package demonstrates how to use `ExecInDeployment` to execute commands inside a specific deployment.

`ExecInDeployment` provides a convenient way to run a command using the deployment name — similar to `ExecInPod`, but for deployments. The deployment with the specified name must exist and be available.

## Options

Pod and container selection within a deployment can be customized using these options:
* `WithDeploymentPodIndex(idx int)`: selects a pod by index.
* `WithDeploymentContainerIndex(idx int)`: selects a container by index.
* `WithDeploymentContainerName(name string)`: selects a container by name.
* `WithDeploymentPod(func(pod v1.Pod) bool)`: selects a pod using a custom rule for fine-grained control.
* `WithDeploymentContainer(func(container v1.Container) bool)`: selects a container using a custom rule for fine-grained control.

The default options are `WithDeploymentPodIndex(0), WithDeploymentContainerIndex(0)`, i.e. the first container of the first pod.

If multiple options are set for the same resource, only the last one is applied. For example:
```go
ExecInDeployment(...,
  WithDeploymentPodIndex(1),
  WithDeploymentPodIndex(2),
  WithDeploymentContainerIndex(1),
  WithDeploymentContainerName("foo"),
  WithDeploymentContainerName("bar"))
```
results in:
```go
ExecInDeployment(...,
  WithDeploymentPodIndex(2),
  WithDeploymentContainerName("bar"))
```

## Examples

Fetch metrics from the `http://localhost:8080/metrics` endpoint from the deployment `controller` in the `default` namespace:
```go
var stdout, stderr *bytes.Buffer
cmd := []string{"wget", "-qO-", "http://localhost:8080/metrics"}
if err := c.Client().Resources().ExecInDeployment(ctx, "default", "controller", cmd, &stdout, &stderr); err != nil {
  t.Log(stderr.String())
  t.Fatal(err)
}

metrics := stdout.String()
```

### Pod and Container Selection

Second pod, first container:
```go
ExecInDeployment(...,
  WithDeploymentPodIndex(1))
```

Second pod, second container:
```go
ExecInDeployment(...,
  WithDeploymentPodIndex(1),
  WithDeploymentContainerIndex(1))
```

First pod, container named `foo`:
```go
ExecInDeployment(...,
  WithDeploymentContainerName("foo"))
```

Pod name starts with `foo-`, container name does not start with `bar-`:
```go
ExecInDeployment(...,
  WithDeploymentPod(func(pod v1.Pod) bool {
    return strings.HasPrefix(pod.Name, "foo-")
  }),
  WithDeploymentContainer(func(container v1.Container) bool {
    return !strings.HasPrefix(container.Name, "bar-")
  }),
```
