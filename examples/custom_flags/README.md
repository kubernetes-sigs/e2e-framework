# Using custom flags

You can pass additional custom flags to the CLI
using the `flag` package and defining the custom flags before calling `envconf.NewFromFlags()`.

For example:

```console
❯ go test -v ./... -args --my-custom-flag hello
=== RUN   TestWithCustomFlag
=== RUN   TestWithCustomFlag/feature
=== RUN   TestWithCustomFlag/feature/custom_flag
    custom_flags_test.go:52: Custom flag my-custom-flag: hello
--- PASS: TestWithCustomFlag (0.00s)
    --- PASS: TestWithCustomFlag/feature (0.00s)
        --- PASS: TestWithCustomFlag/feature/custom_flag (0.00s)
PASS
ok   sigs.k8s.io/e2e-framework/examples/custom_flags 0.491s
```

or by compiling the test code

```console
❯ go test -c -o custom_flags.test .
❯ ./custom_flags.test --help 2>&1 | grep -A2 my-custom
  -my-custom-flag string
     my custom flag for my tests
  -namespace string
```
