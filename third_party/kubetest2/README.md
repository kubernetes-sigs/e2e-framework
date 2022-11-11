# Kubetest2 E2E-Framework Tester

This third-party package implements a Kubernetes-SIGs/[kubetest2](https://github.com/kubernetes-sigs/kubetest2) tester capable of executing tests written using the [e2e-framework](https://github.com/kubernetes-sigs/e2e-framework).

## Kubetest2 installation

First, you must ensure that you have the `kubetest2` binary installed in the environment where you will run your tests :

```
go install sigs.k8s.io/kubetest2/...@latest
```

> See further detail [here](https://github.com/kubernetes-sigs/kubetest2#installation).

## E2E-Framework installation

To build the e2e-framework tester, you can use the following `go install` command:

```
go install sigs.k8s.io/e2e-framework/third_party/kubetest2/...@latest
```

Or, you can build from source:

```
go build ./cmd/kubetest2-tester-e2e-framework/
```

Either approach will produce a binary file named `kubetest2-tester-e2e-framework`, which follows a naming convention that the `kubetest2` binary will use to launch the e2e-framework tester. Next, add your built e2e-framework tester binary to your operating system's `PATH`.

## Running the e2e-framework tester

If the installation steps are successful, run the following to get information on how to run your tests with kubetest2 and the e2e-framewro:

```
$> kubetest2-tester-e2e-framework --help
Usage: kubetest2-tester-e2e-framework [e2e-framework-flags]
When used with kubetest2: kubetest2 [<deployer>] --test=e2e-framework -- [e2e-framework-flags]
e2e-framework flags:
      --assess string               Regular expression to select assessment(s) to run
      --disable-graceful-teardown   Ignore panic recovery while running tests. This will prevent test finish steps from getting executed on panic
      --dry-run                     Run Test suite in dry-run mode. This will list the tests to be executed without actually running them
      --fail-fast                   Fail immediately and stop running untested code
      --feature string              Regular expression to select feature(s) to test
  -h, --help
      --kubeconfig string           Path to a cluster kubeconfig file (optional)
      --labels string               Comma-separated key=value to filter features by labels
      --namespace string            A namespace value to use for testing (optional)
      --packages string             Space-separated packages to test
      --parallel                    Run test features in parallel
      --skip-assessments string     Regular expression to skip assessment(s) to run
      --skip-features string        Regular expression to skip feature(s) to run
      --skip-labels string          Regular expression to skip label(s) to run
      --test-flags string           Space-separated flags applied to 'go test' command
```

To run a test with kubetest2, you must follow this command format as outlined above:

```
kubetest2 [<deployer>] --test=e2e-framework -- [e2e-framework-flags]
```

Where:

* `deployer` is a binary used to stand up an infrastructure (not provided here).
* `--test=e2e-framework` specifies the tester to used, in this case the e2e-framework tester.
* `e2e-framework-flags` are the list of CLI flags that are passed to the e2e-framework binary.

### Running a simple tests

Let us use `kubetest2` to run a simple test with no deployer and no arguments passed to the test:

```
kubetest2 noop --test=e2e-framework -- --packages ./examples/simple
```

The previous command will launch kubetest2 without starting a cluster (noop deployer). Note the `--` which indicates the start of arguments which will be passed to the tester (e2e-framework).

If the command is successful, you will see the `kubetest2` logs similar to what is shown:

```
I1211 08:56:43.836568   14380 app.go:61] The files in RunDir shall not be part of Artifacts
I1211 08:56:43.837144   14380 app.go:62] pass rundir-in-artifacts flag True for RunDir to be part of Artifacts
I1211 08:56:43.837188   14380 app.go:64] RunDir for this run: "/home/ubuntu/go/e2e-framework/examples/_rundir/cf42e94d-1c2e-4c6f-9912-2ff948608008"
I1211 08:56:43.860870   14380 app.go:128] ID for this run: "cf42e94d-1c2e-4c6f-9912-2ff948608008"
Running:  go test  ./simple -args
ok  	sigs.k8s.io/e2e-framework/examples/simple	0.363s
```

You can easily change this command to stand up and teardown a Kubernetes cluster using KinD for instance (`kubetest2` also supports other deployers):

```
$> kubetest2 kind --up --down --test=e2e-framework -- --packages .
I1211 09:20:46.640997   19579 up.go:62] Up(): creating kind cluster...
Creating cluster "kind" ...
 • Ensuring node image (kindest/node:v1.25.3) 🖼  ...
 ✓ Ensuring node image (kindest/node:v1.25.3) 🖼
 • Preparing nodes 📦   ...
...
Running:  go test   -args
PASS
ok  	sigs.k8s.io/e2e-framework/kubetest2test	0.307s
I1211 09:23:14.406049   19579 down.go:32] Down(): deleting kind cluster...
Deleting cluster "" ...
```

> It should be noted that e2e-framework offers its own programmatic hooks for more flexible controls of pre and post test tasks, such as creating and tearing down named clusters (see the examples folder).

### Passing arguments to e2e-framework tests
The kubetest2 e2e-framework tester supports the ability to pass CLI arguments to control settings such as test execution filters and test behaviors. Recall, in order for your e2e-framework tests to receive CLI arguments, you must programmatically create the test environment to receive its configuration from the command-line:

```go

var tenv env.Environment

func TestMain(m *testing.M) {
	var err error
    tenv, err = env.NewFromFlags()
	if err != nil {
		log.Fatalf("failed to build env from flags: %s", err)
	}
	os.Exit(tenv.Run(m))
}

func TestClusterObjects(t *testing.T) {
	f := features.New("cluster-test")
	f.Assess("pod-count", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		var pods corev1.PodList
		client, err := cfg.NewClient()
		if err != nil {
			t.Fatal(err)
		}
		err = client.Resources("kube-system").List(context.TODO(), &pods)
		if err != nil {
			t.Fatal(err)
		}
		if len(pods.Items) == 0 {
			t.Fatal("no pods in namespace kube-system")
		}
		return ctx
	})
	f.Assess("dep-count", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		var deps appsv1.DeploymentList
		client, err := cfg.NewClient()
		if err != nil {
			t.Fatal(err)
		}
		err = client.Resources().List(context.TODO(), &deps)
		if err != nil {
			t.Fatal(err)
		}
		if len(deps.Items) == 0 {
			t.Fatal("deployments not found")
		}
		return ctx
	})
	tenv.Test(t, f.Feature())
}

```

This test can be launched with kubetest2 to include additional CLI flags that are passed to the e2e-framework tester for configurations, as show:

```
$> kubetest2 kind --up --down      \
     --test=e2e-framework --       \
     --packages ./cluster          \
     --kubeconfig=$HOME/.kube/config
```
The previous command will stand up a kind cluster and pass additional CLI flags `--packages` to indicate the test package and and `--kubeconfig` to point to the configuration file.

The test can be launched to only run the `dep-count` assessment by skipping the others, as shown in the following comomand:

```
$> kubetest2 kind --up --down        \
     --test=e2e-framework --         \
     --packages ./cluster            \
     --kubeconfig=$HOME/.kube/config \
     --skip-assessments=pod-count
```

In case you want to pass additional CLI flags to the Go test binary itself, you can use flag `--test-flags` to do so. For instance, the following runs the test with verbose output:

```
$> kubetest2 kind --up --down        \
     --test=e2e-framework --         \
     --packages ./cluster            \
     --kubeconfig=$HOME/.kube/config \
     --skip-assessments=pod-count
     --test-flags="-v"
```

The previous `kubetest2` command will run the following `go test` command as shown:

```
$> go test -v ./cluster -args --kubeconfig=/Users/vivienv/.kube/config --skip-assessment=pod-count
=== RUN   TestClusterObjects
=== RUN   TestClusterObjects/cluster-test
=== RUN   TestClusterObjects/cluster-test/pod-count
    env.go:441: Skipping assessment: "pod-count": name matched
=== RUN   TestClusterObjects/cluster-test/dep-count
--- PASS: TestClusterObjects (0.02s)
    --- PASS: TestClusterObjects/cluster-test (0.02s)
        --- SKIP: TestClusterObjects/cluster-test/pod-count (0.00s)
        --- PASS: TestClusterObjects/cluster-test/dep-count (0.02s)
PASS
ok  	sigs.k8s.io/e2e-framework/kubetest2test/cluster	0.374s
```