# Benchmark tests and skipping features

This directory contains the example of:
- how to run **without test features**, and use your preferred `testing` package (standard library and testify included in example)
  * When skipping features, you will not receive any benefit from the `features` related setup or teardown `envfuncs`, but you will still be able to access a kubernetes config (`envconfig`) to create a client that has access to the expected test cluster.
- how to run Go **Test Benchmarks**, e.g. via `go test -bench=.`

# Skipping features

```go
func BenchmarkListPods(b *testing.B) {
	client, err := testenv.EnvConf().NewClient()
	// ...your client is ready you use -- or use the `*rest.Config` to create your preferred client
}

func TestExample(t *testing.T) {
	client, err := testenv.EnvConf().NewClient()
	// ...
}
```


# Run Tests with flags

These test cases can be executed using the normal `go test -bench=-` command by passing the right arguments

```bash
go test -bench=. -v .
```

With the output generated as following.

```bash
goos: <YOUR_OS>
goarch: <YOUR_ARCH>
pkg: sigs.k8s.io/e2e-framework/examples/benchmark_tests
cpu: <YOUR_CPU_TYPE>
BenchmarkListPods
BenchmarkListPods-12    	     100	 180148936 ns/op
PASS
ok  	sigs.k8s.io/e2e-framework/examples/benchmark_tests	47.880s
```
