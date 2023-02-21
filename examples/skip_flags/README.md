
# CLI flags

The test harness framework supports several CLI flags that can be used to influence how tests are executed. This example shows how to create tests that are configured using the CLI flags.

## Configure tests with CLI flags

To drive your tests with CLI flags, you must initialize a test environment using the passed in CLI flags. This is done by calling `envconf.NewFromFlags` function to create configuration for the environment as shown below:

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

### Supported flags

There are several supported flags (for more accurate list, see package `pkg/flag`):

* `assess`
* `features`
* `labels`
* `kubeconfig`
* `namespace`
* `skip-assessment`
* `skip-features`
* `skip-labels`

### Running tests with flags

The tests can be executed using the normal `go test` tools steps. For instance, to pass the flags to your tests, do the followings:

```shell
go test -v . -args --skip-features "pod list"
```

You can also build a test binary, then pass the CLI flags to the binary. First, compile the test binary:

```shell
go test -c -o skipflags.test .
```

Then execute the test binary passing the CLI arguments:

```shell
./skipflags.test --labels "env=dev"
```

To skip a particular labeled feature , do the following

```shell
./skipflags.test --skip-labels "env=prod"
```

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


To skip a assesment with name `pods from kube-system` within feature `pod list` within test function `TestSkipFlags`,  do the following

```shell
go test -v . -skip TestSkipFlags/pod list/pods from kube-system
``` 

It is not possible to skip features by label name with this option


### Skip tests using both -skip flag and --skip-xxx flags

We can also use the combination of `-skip` flag built in `go test` and `-skip-xxx` flags provided by the e2e-framework to skip multiple tests


To skip a feature `pod list` within test function `TestSkipFlags` and feature `appsv1/deployment` within test function `TestSkipFlags`, do the following

```shell
go test -v . -skip TestSkipFlags/appsv1/deployment -args --skip-features "pod list"
```

To skip a particular labeled feature with label `env=prod` and assesment `deployment creation` within feature `appsv1/deployment` within test function `TestSkipFlags`, do the following

```shell
go test -v . -skip TestSkipFlags/appsv1/deployment/deployment_creation -args --skip-labels "env=prod"
```
