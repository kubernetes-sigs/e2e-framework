# Fail Fast Mode

There are times in your test infra that you want the rest of your feature assessments to fail in
case if one of the current executing assessments fail.
This can aid in getting a better turn around time especially in cases where your assessments are
inter-related and a failed assessment in step 1 can mean that running the rest of the assessment
is guaranteed to fail. This can be done with the help of a `--fail-fast` flag provided at the
framework level.

This works similar to how the `-failfast` mode of the `go test` works but provides the same
feature at the context of the `e2e-framework`.

# How to Use this feature ?

1. Invoke the tests using `--fail-fast` argument
2. Test developers should make sure they invoke either `t.Fail()` or `t.FailNow()` from the assessment to make sure the 
additional handlers kick in to stop the test execution of the feature in question where the assessment has failed


When the framework specific `--fail-fast` mode is used, this works as follows:

1. It stops the rest of the assessments from getting executed for the feature under test
2. This stops the next feature from getting executed in case if the feature under test fails as per step 1.
3. Marks the feature and test associated with it as Failure.
4. Skips the teardown sequence to make sure it is easier to debug the test failure

> Current limitation is that this can't be combined with the `--parallel` switch

Since this can lead to a test failure, we have just documented an example of this. Thanks to @divmadan for the example.

```go
// main_test.go
package example

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testenv = env.NewWithConfig(cfg)

	os.Exit(testenv.Run(m))
}
```

```go
// example_test.go
package example

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestExample(t *testing.T) {
	failFeature := features.New("fail-feature").
		Assess("1==2", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if 1 != 2 {
				t.Log("1 != 2")
				t.FailNow() // mark test case as failed here, don't continue execution
			} else {
				t.Log("1 == 2")
			}
			return ctx
		}).
		Assess("print", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("THIS LINE SHOULDN'T BE PRINTED")
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			t.Log("This teardown should not be invoked")
			return ctx
		}).
 		Feature()

	nextFeature := features.New("next-feature").
		Assess("print", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("THIS LINE ALSO SHOULDN'T BE PRINTED")
			return ctx
		}).
		Feature()

	testenv.Test(t, failFeature, nextFeature)
}

// even if the previous testcase fails, execute this testcase
func TestNext(t *testing.T) {
	nextFeature := features.New("next-test-feature").
		Assess("print", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("THIS LINE SHOULD BE PRINTED")
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			t.Log("This teardown should be invoked")
			return ctx
		}).
		Feature()

	testenv.Test(t, nextFeature)
}
```

When run this using `--fail-fast` you get the following behavior.

```bash
❯ go test . -test.v -args --fail-fast

=== RUN   TestExample
=== RUN   TestExample/fail-feature
=== RUN   TestExample/fail-feature/1==2
    example_test.go:15: 1 != 2
--- FAIL: TestExample (0.00s)
    --- FAIL: TestExample/fail-feature (0.00s)
        --- FAIL: TestExample/fail-feature/1==2 (0.00s)
=== RUN   TestNext
=== RUN   TestNext/next-test-feature
=== RUN   TestNext/next-test-feature/print
    example_test.go:42: THIS LINE SHOULD BE PRINTED
--- PASS: TestNext (0.00s)
    --- PASS: TestNext/next-test-feature (0.00s)
        --- PASS: TestNext/next-test-feature/print (0.00s)
FAIL
```
