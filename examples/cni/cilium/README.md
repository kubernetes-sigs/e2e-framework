# Using Cilium

This directory demonstrates how to use Cilium as a CNI in coordination with the test framework on Kind.

### How it works?

#### CNI Installation (main_test.go)

1. Create the cluster with `disableDefaultCNI` parameter. To do so `CreateClusterWithConfig` is invoked with custom
   configuration provided in `kind-config.yaml`.
2. Create a namespace for workloads.
3. Install Cilium as a Helm chart. First add necessary chart repository and later install the chart in `kube-system`
   namespace.
4. The cluster without CNI is non-functional as nodes status is set to `NotReady`, so that the setup is waiting for Cilium
   deamonset to properly configure network interface and mark nodes as `Ready`, so the tests may proceed.
5. At the end all components are being deleted.

#### Tests (np_test.go)

1. Upload basic Cilium configuration from `templates` folder to:  
   a. set `CiliumClusterwideNetworkPolicy`s to allow connections
   within a cluster to `kube-dns` and from `kube-dns` to `api-server` and externally.  It means that ingress and egress is denied and to enable any other traffic it is required to explicitly declare it (whitelist). (`templates/allow-dns.yaml`)  
   b. allow egress traffic to `api.github.com` on `443` port in `cilium-test` namespace for any nginx
   pod. (`templates/allow-github.yaml`)
2. Create nginx deployment in the specified namespace.
3. Ensure that nginx pod can connect to `api.github.com`.
4. Ensure that nginx pod can't connect to `www.wikipedia.org`.

### How to run?

```bash
go test -c -o cilium.test . && ./cilium.test --v 4
```
