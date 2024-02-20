# Fail Fast Mode

When writing test features using `e2e-framework`, it is possible that you can write test assessments grouped under the same feature that are dependent on each other.
This is possible as the framework ensures that the assessments are processed in the sequence of how they are registered. This key behavior brings up a need to be able to optionally terminate the Feature(s) under test when a specific assessment fails.

## Why not use `test.failfast`?

`go test` provides a handy way to terminate test execution at the first sign of failure via the `-failfast` argument.
However, this terminates the entire test suite in question.

Such termination of the suite is not desirable for the framework as the rest of the Tests can still be processed in case if an assessment in one test fails. This brings in the need to introduce a framework-specific `fail-fast` mode that can perform the following.

1. It should Terminate the feature(s) under test and mark the test as failure.
2. Skip the Teardown workflow of the feature(s) under test to enable easy debugging.

## Framework specific `--fail-fast` Mode

`e2e-framework` introduces a new CLI argument flag that can be invoked while triggering the test to achieve the fail-fast behavior built into the framework.

There are certain caveats to how this feature works.

1. The `fail-fast` mode doesn't work in conjunction with the `parallel` test mode.
2. Test developers have to explicitly invoke the `t.Fail()` or `t.FailNow()` handlers in the assessment to inform the framework that the fail-fast mode needs to be triggered.

## Example Assessment

Below section shows a simple example of how the feature can be leveraged in the assessment. This should be combined with `--fail-fast` argument while invoking the test to leverage the full feature.

```go
func TestFeatureOne(t *testing.T) {
    featureOne := features.New("feature-one").
        Assess("this fails", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
            if 1 != 2 {
                t.Log("1 != 2")
                t.FailNow() // mark test case as failed here, don't continue execution
            } else {
                t.Log("1 == 2")
            }
            return ctx
        }).
        Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
            t.Log("This teardown should not be invoked")
            return ctx
        }).
        Feature()
    testenv.Test(t, failFeature, nextFeature)
}
```
