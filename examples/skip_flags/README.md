
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