## Execute commands from the running container

This directory contains the example of how to run a command inside a running container by buildin function `ExecInPod` provided by e2e framework. 

### How to use `ExecInPod` function?

First of all there should be a pod with a proper container, which has `Running` status to execute commands from it. To meet status condition within a test either `wait.For()` or `resources.Watch()` may be used. 

To invoke a function it is required to pass the following parameters:

| Param          | Type          | Description |
|----------------|---------------|-----------------------------------------|
| namespaceName  | string        | Namespace name, where the pod is running |
| podName  | string        | Pod name                                |
| containerName  | string        | Container name                          |
| command  | [] string     | Command to be executed in container     |
| stdout  | *bytes.Buffer | Buffer pointer to read from in case of successful command execution     |
| stderr  | *bytes.Buffer | Buffer pointer  to read from in case a command failed     |

### What does this test do?

1. Create a Kind cluster with a random name with `exectest-` as the cluster name prefix.
2. Create a custom namespace with `my-ns` as the prefix.
3. Create nginx deployment with one replica.
4. Wait for the deployment to be `Available`.
5. Use curl to request the main page of Wikipedia.org by `ExecInPod` function.
6. Check is a status code equals 200.

### How to run the tests

```bash
go test -v .
```
