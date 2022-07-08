# E2E Framework

[![godoc](https://pkg.go.dev/badge/github.com/sigs.k8s.io/e2e-framework)](https://pkg.go.dev/sigs.k8s.io/e2e-framework)

A Go framework for end-to-end testing of components running in Kubernetes clusters.

The primary goal of this project is to provide a `go test`(able)
framework that uses the native Go testing API to define end-to-end tests suites
that can be used to test Kubernetes components. Some additional goals
include:

* Provide a sensible programmatic API to compose tests
* Leverage Go's testing API to compose test suites
* Expose packages that are easy to programmatically consume
* Collection of helper functions that abstracts Client-Go functionalities
* Rely on built-in Go test features to easily select/filter tests to run during execution
* And more

For more detail, see the [design document](./docs/design/test-harness-framework.md).

## Getting started

The Go package is designed to be integrated directly in your test. Simply update your project to pull the desired Go modules:

```
go get sigs.k8s.io/e2e-framework/pkg/env
go get sigs.k8s.io/e2e-framework/klient
```

### Using the framework

The framework uses the built-in Go testing framework directly to define and run tests.

#### Setup `TestMain`

Use function `TestMain` to define package-wide testing steps and configure behavior. The following examples uses pre-defined steps to create a `KinD` cluster before running any test in the package:

```go
var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
    testenv = env.New()
    kindClusterName := envconf.RandomName("my-cluster", 16)
    namespace := envconf.RandomName("myns", 16)

    // Use pre-defined environment funcs to create a kind cluster prior to test run
    testenv.Setup(
        envfuncs.CreateKindCluster(kindClusterName),
    )

    // Use pre-defined environment funcs to teardown kind cluster after tests
    testenv.Finish(
        envfuncs.DeleteNamespace(namespace),
    )
    
    // launch package tests
    os.Exit(testenv.Run(m))
}
```

#### Define a test function

Use a Go test function to define features to be tested as shown below:

```go
func TestKubernetes(t *testing.T) {
    f1 := features.New("count pod").
        WithLabel("type", "pod-count").
        Assess("pods from kube-system", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
            var pods corev1.PodList
            err := cfg.Client().Resources("kube-system").List(context.TODO(), &pods)
            if err != nil {
                t.Fatal(err)
            }
            if len(pods.Items) == 0 {
                t.Fatal("no pods in namespace kube-system")
            }
            return ctx
        }).Feature()

    f2 := features.New("count namespaces").
        WithLabel("type", "ns-count").
        Assess("namespace exist", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
            var nspaces corev1.NamespaceList
            err := cfg.Client().Resources().List(context.TODO(), &nspaces)
            if err != nil {
                t.Fatal(err)
            }
            if len(nspaces.Items) == 1 {
                t.Fatal("no other namespace")
            }
            return ctx
        }).Feature()
        
    // test feature
    testenv.Test(t, f1, f2)
}
```
#### Running the test
Use the Go testing tooling to run the tests in the package as shown below. The following would run all tests except those with label `type=ns-count`:

```
go test ./package -args --skip-labels="type=ns-count"
```

## Examples
See the [./examples](./examples) directory for additional examples showing how to use the framework.

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/sig-testing)
- [Mailing List](https://kubernetes.slack.com/messages/sig-testing)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
