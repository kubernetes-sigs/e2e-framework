# `KUBECONFIG` Usage

The test harness framework supports detection of kubeconfig similar to [`kubectl`](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_config/). 
This is useful when there are mulitple potential sources of kubeconfig's on the user's system, each having a different priority.  

This example shows how to fully leverage this configuration.

## Setting the Env

To use a single kubeconfig path identical to the recommended default

```bash
KUBECONFIG=~/.kube/config
```

To use a relative kubeconfig path for development and a fallback to the recommended default

```bash
KUBECONFIG=.kube/config:~/.kube/config
```

> The path separator is dependent on Operating System, so expect to use `;` instead of `:` on Windows systems

## Running tests

This replaces the need for the `--kubeconfig` CLI flag in cases where the `KUBECONFIG` is already set.

```bash
KUBECONFIG=.kube/config:~/kube.config go test -v .
```
