# Kubernetes Custom Resource Integration Test

While developing a kubernetes operator for your custom use cases, it is very common practice that you end up creating your own custom resources.

This example shows how to leverage the helper functions provided by the Framework itself to setup
the CRD resources using the `decoder` package against your test cluster before starting the actual test workflow. 

## How does this work ?

1. You can leverage the framework's `env.Func` type helper for setting up the CRDs and tearing them down after the tests
2. Register the CRD scheme with the `resources.Resources` to leverage the helpers for interacting with Custom resource objects

## What does this test do ?

1. Create a Kind cluster with a random name generated with `crdtest-` as the cluster name prefix
2. Create a custom namespace with `my-ns` as the prefix
3. Register the CRDs listed under `./testdata/crds` using the resource decode helpers
4. Create a new Custom Resource for the CRD created in step #3
5. Fetch the CR created in Test setup and print the value

## How to run the tests

```bash
go test -v .
```
