# Multi Cluster Test Runs

This directory contains the example of how to run tests on multiple `kind` clusters at the same time using the same `env.Environment`

## What does this test do ?

1. Create two clusters
   1. One with prefix cluster-one
   2. One with prefix cluster-two
2. Install A sample helm chart on both clusters
3. Run an assessment to check if the chart has successfully been deployed by checking the pod status
4. Teardown the Test Environments

# Run Tests

These test cases can be executed using the normal `go test` command by passing the right arguments

```bash
go test -v .
```

```bash
go test -c -o multi-cluster.test .

./multi-cluster.test --v 4
```