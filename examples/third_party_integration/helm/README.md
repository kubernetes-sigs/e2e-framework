# Helm Integration

This section of the document gives you an example of how to integrate the helm chart related workflow
into the `e2e-framework` while writing your tests.

## Pre-Requisites

1. `Helm3` Installed on your system where the tests are being run

## How does `TestLocalHelmChartWorkflow` test work ?

1. It creates a Kind Cluster with `third-party` prefix
2. Creates a new namespace with `third-party` prefix
3. Deploys the local helm chart under [example_chart](testdata/example_chart) with a name `example` to namespace created in Step #2
5. Run `helm test example` command to run a test on the Helm chart deployed in step #3

## How does `TestHelmChartRepoWorkflow` test work?

1. It creates a Kind Cluster with `third-party` prefix
2. Creates a new namespace with `third-party` prefix
3. Adds the `nginx-stable` helm repo and triggers a repo update workflow
4. Installs the helm chart using `nginx-stable/nginx-ingress` (without Wait mode) with name `nginx`
6. Waits for the Deployment to be up and running
7. Runs the `helm test nginx` command to run a basic helm test


## How to Run the Tests

```bash
go test -c -o helm.test .

./helm.test --v 4
```