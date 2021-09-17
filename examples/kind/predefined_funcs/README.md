# Using Pre-defined Environment Funcs

This package demonstrates how to use all part of the framework to setup and run an
end-to-end test that retrieves API objects from Kubernetes for assessments.  Specifically,
it shows how to use pre-defined environment functions when setting up testing environment.

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

### Configure environment `Setup` functions
The next step is to define environment functions that are executed during setup.
This example uses predefined environment functions, in `env.Step`, to specify the steps to execute prior to running any test:
  * `envfuncs.CreateKindCluster` - creates a new kind cluster as part of the setup (and stores it in the context)
  * `envfuncs.CreateNamespace` - creates a new namespace API object on the server and sets the environment configuration to use it.
```go
func TestMain(m *testing.M) {
    testenv = env.New()
    kindClusterName := envconf.RandomName("my-cluster", 16)
    namespace := envconf.RandomName("myns", 16)

    // Setup uses pre-defined funcs to create kind cluster
    // and create a namespace for the environment
    testenv.Setup(
        envfuncs.CreateKindCluster(kindClusterName),
        envfuncs.CreateNamespace(namespace),
    )
...
}
```

### Configure environment `Finish` for cleanup
Next, the `Environment.Finish` method is used to define steps that are executed to teardown the environment
after all tests are executed. The example uses predefined environment functions:

* `envfuncs.DeleteNamespace` - removes the namespace that was previously created
* `envfuncs.DestroyKindCluster` - removes all cluster resources and deletes the kind cluster that was created earlier 

```go
func TestMain(m *testing.M) {
	testenv = env.New()
	...

	// Finish uses pre-defined funcs to
	// remove namespace, then delete cluster
	testenv.Finish(
        envfuncs.DeleteNamespace(namespace),
        envfuncs.DestroyKindCluster(kindClusterName),
    )
...
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

### Testing Kubernetes resources
The test function in this example, found in `k8s_test.go`, use the framework to define two test features.

* `podFeature` - One test simply pulls a list of pods and report on the number of items 
* `depFeature` - The other test inject a deployment object, inspect, and then deletes it

#### Defining a feature to test pod items

```go
func TestKubernetes(t *testing.T) {
    podFeature := features.New("pod list").
    	Assess("pods from kube-system", func (ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
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
    	}).Feature()
...    
}
```

#### Defining a feature to test deployment objects

```go
func TestKubernetes(t *testing.T) {
...
	// feature uses pre-generated namespace (see TestMain)
	depFeature := features.New("appsv1/deployment").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// insert a deployment
			deployment := newDeployment(cfg.Namespace(), "test-deployment", 1)
			if err := cfg.Client().Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			time.Sleep(2 * time.Second)
			return ctx
		}).
		Assess("deployment creation", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var dep appsv1.Deployment
			if err := cfg.Client().Resources().Get(ctx, "test-deployment", cfg.Namespace(), &dep); err != nil {
				t.Fatal(err)
			}
			if &dep != nil {
				t.Logf("deployment found: %s", dep.Name)
			}
			return context.WithValue(ctx, "test-deployment", &dep)
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			dep := ctx.Value("test-deployment").(*appsv1.Deployment)
			if err := cfg.Client().Resources().Delete(ctx, dep); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

...
}
```

#### Execute the feature test
The last step in defining the test is to use the framework to launch the tests.

```go
func TestKubernetes(t *testing.T) {
	podFeature := features.New("pod list")...
	depFeature := features.New("appsv1/deployment")...

	testenv.Test(t, podFeature, depFeature)
}
```

## Run the test
Use the Go test tool to run the test.

```bash
$ go test . -v
```
