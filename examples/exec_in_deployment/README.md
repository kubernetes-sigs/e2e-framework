# Execute Commands in a Deployment

This package demonstrates how to use `ExecInDeployment` to execute commands inside a specific deployment.

`ExecInDeployment` mimics the behavior of `kubectl exec -it -n <namespace> deploy/<deployment> -- <command>`: it selects the first pod of the specified deployment and runs the command in the first container of that pod.

It provides a convenient way to run a command using only the deployment name — similar to `ExecInPod`, but for deployments, and defaulting to the first container of the first pod.

`ExecInDeployment` expects the deployment with the specified name to exist and be available.

## Example

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
