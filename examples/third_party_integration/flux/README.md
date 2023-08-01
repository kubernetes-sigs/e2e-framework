# Flux Integration

This section of the document gives you an example of how to integrate the flux workflow
into the `e2e-framework` while writing your tests.

## Pre-Requisites

1. `Flux` Installed on your system where the tests are being run for details visit flux official [website](https://fluxcd.io/).

## Flux supported commands

For the time being the framework supports following functionality: 
- Flux installation and uninstallation.
- Handling [Kustomization](https://fluxcd.io/flux/components/kustomize/kustomization/) objects.
- Handling [GitRepository](https://fluxcd.io/flux/components/source/gitrepositories/) objects.

## How does the example work?

1. It creates a kind cluster with `flux` prefix.
2. Creates a namespace with `flux` prefix.
3. Installs all flux resources.
4. Creates a reference to the git repository, where a simple hello world application deployment is specified. You can find it [here](https://github.com/matrus2/go-hello-world).
5. Starts reconciliation by a flux kustomization manifest to path `template` of the git repository.
6. Assesses if the deployment of simple hello world app is up and running.
7. After the test passes it removes all resources.

## How to run tests

```shell
go test -c -o flux.test . && ./flux.test --v 4
```