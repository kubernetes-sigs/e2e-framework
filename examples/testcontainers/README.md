# Using testcontainers with E2E Framework

[Testcontainers](https://golang.testcontainers.org/) is an open source library that provides lightweight, throwaway instances of databases, message brokers, web browsers, or any other software that can run in a container. It allows developers to create reliable and repeatable test and development environments by spinning up real, isolated services from within their codebase using containerized environments.

This specific example demonstrates how to use [`Testcontainers`](https://github.com/testcontainers/testcontainers-go) with the [`e2e-framework`](https://github.com/kubernetes-sigs/e2e-framework/) to create ephemeral test cluster(s) for testing. It sets up a [K3s](https://k3s.io/) cluster in a container using the [`k3s module`](https://golang.testcontainers.org/modules/k3s/) and then runs tests against that cluster.

Steps performed in the example:

1. Creates a K3s Kubernetes cluster in a single container using testcontainers.
2. Pulls and loads specified container images into the K3s cluster (currently nginx and busybox).
3. Configures the `e2e-framework` to use the cluster's kubeconfig.
4. Sets up temporary namespace for testing.
5. Creates a deployment called `test-deployment` with nginx and busybox images that were loaded previously and waits for the deployment to be ready.
6. Deletes the test deployment and namespace.
7. Cleans up the K3s container after tests are complete.

## Running the tests
To run the tests, ensure you have Docker or Podman installed and running on your machine. Then, execute the following command in the terminal:

```bash
    cd examples/testcontainers
    go mod tidy
    go test -v ./... -count=1
```