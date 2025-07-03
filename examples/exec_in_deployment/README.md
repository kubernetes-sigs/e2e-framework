# Execute Commands in a Deployment

This package demonstrates how to use `ExecInDeployment` to execute commands inside a specific deployment.

`ExecInDeployment` provides a convenient way to run a command using the deployment name — similar to `ExecInPod`, but for deployments. The deployment with the specified name must exist and be available.

## Options

* `WithPodIndex(int)`: pod index to select, 0 by default (i.e. the first pod).

## Examples

Fetch metrics from the `http://localhost:8080/metrics` endpoint in the deployment `controller` running in the namespace `default`:
```go
var stdout, stderr *bytes.Buffer
cmd := []string{"wget", "-qO-", "http://localhost:8080/metrics"}
if err := c.Client().Resources().ExecInDeployment(ctx, "default", "controller", cmd, &stdout, &stderr); err != nil {
  t.Log(stderr.String())
  t.Fatal(err)
}

metrics := stdout.String()
```

Show the hostname of the second pod in the deployment `controller` running in the namespace `default`:
```go
var stdout, stderr *bytes.Buffer
cmd := []string{"hostname"}
if err := c.Client().Resources().ExecInDeployment(ctx, "default", "controller", cmd, &stdout, &stderr, resources.WithPodIndex(1)); err != nil {
  t.Log(stderr.String())
  t.Fatal(err)
}
```
