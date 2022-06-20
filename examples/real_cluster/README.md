# Real cluster example (including cloud provider)

This package shows how the test framework can be used with a real self-managed or cloud vendor managed cluster (AKS/GKE/EKS), since one may want to launch some tests inside the cluster. This example creates a random namespace, deploys simple deployment and if done removes the namespace.

To properly connect to a cloud provider cluster it is required to import for side effects one of auth module. If testing on self-managed cluster there is no need to include these dependencies.
```go
	_ "k8s.io/client-go/plugin/pkg/client/auth/azure" // auth for AKS clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"   // auth for GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"  // auth for OIDC (EKS)
```

ResolveKubeConfigFile() function is called to get kubeconfig loaded, it uses either `--kubeconfig` flag, `KUBECONFIG` env or by default ` $HOME/.kube/config` path.

```go
func TestMain(m *testing.M) {
	testenv = env.New()
	path := conf.ResolveKubeConfigFile()
	cfg := envconf.NewWithKubeConfig(path)
	testenv = env.NewWithConfig(cfg)
```

Later testing is the same as on Kind cluster.

```go
	testenv.Setup(
		envfuncs.CreateNamespace(namespace),
	)
	testenv.Finish(
		envfuncs.DeleteNamespace(namespace),
	)
	os.Exit(testenv.Run(m))
```
