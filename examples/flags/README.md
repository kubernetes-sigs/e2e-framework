
# CLI flags

The test harness framework supports several CLI flags that can be used to influence how tests are executed. This example shows how to create tests that are configured using the CLI flags.

## Configure tests with CLI flags

To drive your tests with CLI flags, you must initialize a test environment using the passed in CLI flags. This is done by calling `envconf.NewFromFlags` function to create configuration for the environment as shown below:

```go
var test env.Environment

func TestMain(m *testing.M) {
    cfg, err := envconf.NewFromFlags()
    if err != nil {
    	log.Fatalf("envconf failed: %s", err)
    }
    test = env.NewWithConfig(cfg)
    os.Exit(test.Run(m))
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
* `v`

> We also embed all supported flags from klog into supported flags. For details of these flags please
> refer to [k8s.io/klog/v2](https://github.com/kubernetes/klog/blob/main/klog.go#L424)

### Running tests with flags

The tests can be executed using the normal `go test` tools steps. For instance, to pass the flags to your tests, do the followings:

```shell
go test -v . -args --assess en
```

You can also build a test binary, then pass the CLI flags to the binary. First, compile the test binary:

```shell
go test -c -o flags.test .
```

Then execute the test binary passing the CLI arguments:

```shell
./flags.test --assess en
```

To skip a particular assessment , do the following

```shell
./flags.test --skip-assessment en
```

To get additional verbose logs

```shell
./flags.test --assess es --v 2
```

To run a test against a particular Kubeconfig context

```shell
./flags.test --kubeconfig ~/path/to/kubeconfig --context my-context
```
