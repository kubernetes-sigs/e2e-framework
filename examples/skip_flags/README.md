
# CLI flags

The test harness framework supports several CLI flags that can be used to influence how tests are executed. This example
shows how to run or skip specific tests (features) using `--skip-features`, `--labels`, and `--skip-labels`. See
[examples/flags](../flags/README.md) for a full list of supported flags.

## Configure tests with CLI flags

To drive your tests with CLI flags, you must initialize a test environment using the passed in CLI flags. This is done
by calling `envconf.NewFromFlags` function to create configuration for the environment as shown below:

```go
var test env.Environment

func TestMain(m *testing.M) {
    // parse and load flags to configure environment
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		
    test = env.NewWithConfig(cfg)
    ...
}
```

### Use `Labels` in your tests

To have more fine-grained control over which feature is run/skipped, you can use `WithLabel()` to add labels to your
tests and use those with the `--labels` and `--skip-labels` CLI flags.

### Running tests with flags

#### Using `--skip-features`

The tests can be executed using the normal `go test` tools steps. For instance, to skip the `pod list` feature, pass the
`--skip-features` flag to your tests:

```shell
go test -v . -args --skip-features "pod list"
```

You can also build a test binary, then pass the CLI flags to the binary. First, compile the test binary:

```shell
go test -c -o skipflags.test .
```

Then execute the test binary passing the CLI arguments:

```shell
./skipflags.test --skip-features "pod list"
```

#### Using `--labels` and `--skip-labels`

Adding one or more labels to your features using `WithLabels()` gives you more control over which features are executed.

To skip features labeled with `"env"` is `"dev"`:

```shell
go test -v . -args --skip-labels=env=dev
```

To only run features labeled with `"type"` `"k8score"`:

```shell
go test -v . -args --labels=type=k8score
```

To be even more explicit, `--labels` and `--skip-labels` allows for multiple labels to be passed. For example, to run
all tests labeled with `"type"` `"k8score"` **and** labeled with `"env"` `"prod"`:

```shell
go test -v . -args --labels=type=k8score,env=prod
```

You can combine `--labels` and `--skip-labels` for even more control. For example, to run all features labeled with
`"type"` `"k8score"` but exclude those labeled with `"env"` is `"dev"`:

```shell
go test -v . -args --labels=type=k8score --skip-labels=env=dev
```

> **Note**
> Features without a label or where the specified `--labels` do not exactly match are also excluded from the tests.

### Skip tests using built in -skip flag in go test 

Go 1.20 introduces the `-skip` flag for `go test` command to skip tests. 

Tests can also be skipped based on test function name, feature name and assesment name with `-skip` flag

```shell
go test -v . -skip <test_function_name>/<feature_name>/<assesment_name>
```

To skip a test by test function name `TestSkipFlags`, do the following

```shell
go test -v . -skip TestSkipFlags
```

To skip a feature with name `pod list` within test function `TestSkipFlags`, do the following

```shell
go test -v . -skip TestSkipFlags/pod list
```

To skip a assesment with name `pods from kube-system` within feature `pod list` within test function `TestSkipFlags`,
do the following

```shell
go test -v . -skip TestSkipFlags/pod list/pods from kube-system
``` 

It is not possible to skip features by label name with this option

### Skip tests using both -skip flag and --skip-xxx flags

We can also use the combination of `-skip` flag built in `go test` and `-skip-xxx` flags provided by the e2e-framework
to skip multiple tests

To skip a feature `pod list` within test function `TestSkipFlags` and feature `appsv1/deployment` within test function
`TestSkipFlags`, do the following

```shell
go test -v . -skip TestSkipFlags/appsv1/deployment -args --skip-features "pod list"
```

To skip a particular labeled feature with label `env=prod` and assesment `deployment creation` within feature
`appsv1/deployment` within test function `TestSkipFlags`, do the following

```shell
go test -v . -skip TestSkipFlags/appsv1/deployment/deployment_creation -args --skip-labels "env=prod"
```
