# Using testcontainers with E2E Framework

This example demonstrates how to use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) with the `e2e-framework` to create ephemeral test cluster(s) for testing. It sets up a K3s cluster in a container using the `testcontainers` [k3s module](https://golang.testcontainers.org/modules/k3s/) and then runs tests against that cluster.

Steps:

1. Creates a K3s container using testcontainers
2. Pulls and loads specified images into the cluster (currently nginx and busybox)
3. Configures the `e2e-framework` to use the the cluster's kubeconfig
4. Sets up temporary namespace for testing
5. Creates a deployment called `test-deployment` with nginx and busybox images that were loaded previously and waits for the deployment to be ready
6. Deletes the test deployment and namespace
7. Cleans up the K3s container after tests are complete

