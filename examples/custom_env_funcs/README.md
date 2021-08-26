# Using Custom Environment Funcs

This package demonstrates how to use all part of the framework to setup and run an
end-to-end test that retrieves API objects from Kubernetes for assessments.  Specifically,
it shows how to use custom environment functions when setting up testing environment.

## TestMain (test suite)
File `main_tests.go` sets up the test suite for the package.  In this test, the environment functions
are used to

### Declare the test environment
The first step is to declare the global environment that will be used to run the test.

```go
var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
   testenv = env.New()
}
```

### Define an environment `Setup` function
The next step is to define an environment function in the `Setup` method as shown below. The example
shows a function that is used to create a kind cluster (using the `support/kind` package), then
stores the cluster (of type `*kind.Cluster`) in the context for future operations.
 
```go
func TestMain(m *testing.M) {
	testenv = env.New()
	testenv.Setup(
		// Step: creates kind cluster, propagate kind cluster object
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			name := envconf.RandomName("my-cluster", 16)
			cluster := kind.NewCluster(name)
			kubeconfig, err := cluster.Create()
			if err != nil {
				return ctx, err
			}
			// stall a bit to allow most pods to come up
			time.Sleep(time.Second * 10)

			// update environment with kubecofig file
			if _, err := cfg.WithKubeconfigFile(kubeconfig); err != nil {
				return ctx, err
			}

			// propagate cluster value
			return context.WithValue(ctx, 1, cluster), nil
		},
	)
	...
}
```

### Define a teardown operation in  `Finish`
The next custom operation is defined as parameter for the `Finish` method which is executed at the end of all
package tests. The operation retrieves the cluster value, from the context, and calls its destroy method to delete
the kind cluster.

```go
func TestMain(m *testing.M) {
	testenv = env.New()
	...
	testenv.Finish(
		// Teardown func: delete kind cluster
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			cluster := ctx.Value(1).(*kind.Cluster) // nil should be tested
			if err := cluster.Destroy(); err != nil{
				return ctx, err
			}
			return ctx, nil
		},
	)

	os.Exit(testenv.Run(m))
}
```

### Start the test suite
The last step in defining the test suite is to launch it:
```go
func TestMain(m *testing.M) {
	...
	os.Exit(testenv.Run(m))
}
```
## Test functions
The framework uses a regular Go test function to define the end-to-end test.  Once a suite is launched, after its setup
is successful, Go test will run the Go test functions in the package.

### Testing Kubernetes
The test functions in this example, found in `k8s_test.go`, use the framework to define a test feature.  The following
example defines one test feature with a single assessment.  The assessment retrieves pods in the `kube-system`
namespace and inspect the quantity returned.

```go
func TestKubernetes(t *testing.T) {
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