# Testing Kubernetes Resources

This package demonstrates how to use all part of the framework to setup and run an
end-to-end test that retrieves API objects from Kubernetes for assessments.  The test
is split into two parts, the test suite definition and the test functions.

## Test suite definition
File `main_tests.go` sets up the test suite for the package.

### Declare the global 
The first step is to declare the global environment that will used to run the test.

```go
var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
   testenv = env.New()
}
```

### Define environment `Setup` functions
The next step is to define some environment functions. Here we have two `Setup` EnvFuncs 
to do the following setups:
  * The first `Setup` func, uses functions from `kindsupport.go`, to create a kind cluster and gets the kubeconfig file as a result
  * The second setup func, creates a `klient.Client` type used to interact with the API server
```go
func TestMain(m *testing.M) {
	testenv = env.New()
	testenv.Setup(
		// env func: creates kind cluster, propagate kubeconfig file name
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			cluster := envconf.RandomName("my-cluster", 16)
			kubecfg, err := createKindCluster(cluster)
			if err != nil {
				return ctx, err
			}
			// stall a bit to allow most pods to come up
			time.Sleep(time.Second*10)

			// propagate cluster name and kubeconfig file name
			return context.WithValue(context.WithValue(ctx, 1, kubecfg), 2, cluster), nil
		},
		// env func: creates a klient.Client for the envconfig.Config
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			kubecfg := ctx.Value(1).(string)
			// create a klient.Client and set it for the env config
			client, err := klient.NewWithKubeConfigFile(kubecfg)
			if err != nil {
				return ctx, fmt.Errorf("create klient.Client: %w", err)
			}
			cfg.WithClient(client) // set client in envconfig
			return ctx, nil
		},
	)
...
}
```

### Define environment `Finish` for cleanup
The last step in creating the suite is to define cleanup code for the environment.

```go
func TestMain(m *testing.M) {
	testenv = env.New()
	...
	testenv.Finish(
        // Teardown func: delete kind cluster and delete kubecfg file
        func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
            kubecfg := ctx.Value(1).(string)
            cluster := ctx.Value(2).(string)
            if err := deleteKindCluster(cluster, kubecfg); err != nil {
                return ctx, err
            }
            return ctx, nil
        }, 
    )
}
```
### Start the suite
The last step in defining the test suite is to launch it:
```go
func TestMain(m *testing.M) {
	testenv = env.New()
	...
	os.Exit(testenv.Run(m))
}
```

* Uses support functions, found in `kindsupport.go` to stand up a kind cluster

The test suite  uses context propagation to propagate the `kubeconfig` file name and the name of the cluster, that created,
so that they can be cleaned in the `Finish` stage.

## Test functions
The framework uses a regular Go test function to define the end-to-end test.  Once a suite is launched, after its setup
is successful, Go test will run the Go test functions in the package.

### Testing Kubernetes
The test functions in this example, found in `k8s_test.go`, use the E2E test harness framework to define a test feature. The framework
allows test authors to inject custom function definitions as hooks that get executed during the test.  The following
example defines one test feature with a single assessment.  The assessment retrieves pods in the kube-system
namespace and inspect the quantity returned.

```go
func TestListPods(t *testing.T) {
	f := features.New("example with klient package").
		Assess("get pods from kube-system", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var pods corev1.PodList
			err := cfg.Client().Resources("kube-system").List(context.TODO(), &pods)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("found %d pods", len(pods.Items))
			if len(pods.Items) == 0 {
				t.Fatal("no pods in namespace kube-system")
			}
			return ctx
		})

	testenv.Test(t, f.Feature())
}
```

## Run the test
Use the Go test tool to run the test.

```bash
$ go test . -v
```
