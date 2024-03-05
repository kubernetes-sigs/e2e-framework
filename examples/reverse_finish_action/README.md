# Reverse Ordering Test Finish Actions

The `e2e-framework` supports performing of the test finish actions in the reverse order in order to get a behavior similar to deferrring a function.

This enables the developers to perform setup/finish steps in lockstep as part of the test development workflow and the finish actions can be triggered in the reverse order of how they are registered. This help in performing cleanup workflow in a more graceful way and makes it more readable. This also enables mimicking the nature of how `t.Cleanup` works.

## Running Tests

The tests can be executed using the normal `go test` tools steps. For instance, to pass the flags to your tests, do the followings:

```shell
go test -v . -args -feature-gates ReverseTestFinishExecutionOrder=true
```

## Test Output
```bash
=== RUN   TestReverseFinishAction
    reverse_finish_action_test.go:69: Running ReverseFinishAction
--- PASS: TestReverseFinishAction (0.00s)
PASS
I0305 12:26:33.313466   66406 reverse_finish_action_test.go:63] "Action Trigger Ordering" setupAction=[1,2] finishAction=[2,1]
ok  	sigs.k8s.io/e2e-framework/examples/reverse_finish_action	0.416s
```

```bash
❯ go test -v .
=== RUN   TestReverseFinishAction
    reverse_finish_action_test.go:69: Running ReverseFinishAction
--- PASS: TestReverseFinishAction (0.00s)
PASS
I0305 12:26:46.470189   66475 reverse_finish_action_test.go:63] "Action Trigger Ordering" setupAction=[1,2] finishAction=[1,2]
ok  	sigs.k8s.io/e2e-framework/examples/reverse_finish_action	0.426s
```