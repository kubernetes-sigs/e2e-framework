# Flux Integration

This section of the document gives you an example of how to integrate the flux workflow
into the `e2e-framework` while writing your tests.

## Pre-Requisites

1. `Flux` Installed on your system where the tests are being run for details visit flux official [website](https://fluxcd.io/).

## Flux supported commands

For the time being the framework supports following functionality: 
- Flux installation and uninstallation.
- Creation and removal of [Kustomization](https://fluxcd.io/flux/components/kustomize/kustomization/) objects.
- Creation and removal of [GitRepository](https://fluxcd.io/flux/components/source/gitrepositories/) objects.
- Creation and removal of [HelmRepository](https://fluxcd.io/flux/components/source/helmrepositories/) objects.
- Creation and removal of [HelmRelease](https://fluxcd.io/flux/components/helm/helmreleases/) objects.

## Directory structure
```
flux  
│   
└───nginx    - example of using GitRepository and Kustomization flux objects based on simple nginx deployment.  
│   
└───kyverno  - example of using HelmRepository and HelmRelease flux objects based on Kyverno policy engine.  
│  
└───template - a deployment, service and namespace for nginx example. They are referenced for reconciliation by gitops agent. 
               Content and path should not be changed as it is referenced directly from nginx example.
```
## Examples

### 1. NGINX (GitRepository + Kustomization)

Installation of nginx deployment by using GitRepository and Kustomization flux objects. Steps:

1. Create a kind cluster with `flux` prefix.
2. Create a namespace with `flux` prefix.
3. Install flux and its resources.
4. Create a reference to the git repository (e2e-framework).
5. Starts reconciliation by a flux kustomization manifest to path `examples/third_party_integration/flux/template` of previously defined git repository.
6. Assess if the deployment is working correctly in the cluster.
7. After the test passes all objects are removed in the reversed order.

### 2. KYVERNO (HelmRepository + HelmRelease)

Installation of Kyverno helm chart by combination of HelmRepository and HelmRelease flux objects. There are at least two scenarios, where e2e tests may be written. First check configuration of polices or/and check how different workflows may react in the cluster, where different security policies were preinstalled by cluster operators. In this simple example all essential policies, which comes with Kyverno are enforced. For the sake of this example kyverno is configured to allow to run pod with privileged containers only if they have specific labels (`app:admin`). The chart configuration is specified in `kyverno-values.yaml`. The test perform the following steps: 

1. Create a kind cluster with `flux` prefix.
2. Create a namespace with `flux` prefix.
3. Install flux and its resources.
4. Create a reference to the helm repository (kyverno).
5. Create a helm release controller, which installs and starts reconciliation of kyverno chart.
6. Create another helm release controller, but this time for kyverno policies chart. The values file is provided.
7. Deploy a simple nginx deployment with a privileged container. Since the policies are enforced the pod should not run.
8. Deploy the same deployment, but with specific labels. The pod should run correctly.
9. After the test passes all objects are removed in the reversed order.


## How to run tests

```shell
go test -c -o flux.test . && ./flux.test --v 4
```