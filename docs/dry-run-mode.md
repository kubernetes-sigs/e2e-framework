# Dry Run Mode

As your test suits get bigger and complicated over the period of time, it is essential that the toolings used for creating tests provide an easy way to identify and list the tests being
processed as part of your framework when invoked with certain arguments. And this listing needs
to be quick and clean in order to enable quick turn around time of test development. This requirement bring in the need to introduce a `dry-run` behavior into the `e2e-framework`.

## Unit Of Test

Go treats each function starting with `Testxxx` as the Test unit. However, the same is not entirely true in case of the `e2e-framework`. This introduces dynamic tests that are generated during the runtime programmatically for each assessment of each feature.

From the perspective of the `e2e-framework`, the Unit of test is an `Assessment` that actually performs the assertion of an
expected behavior or state of the system. These assessments are run as a sub-test of the main test identified by the function
`Testxxx`. All framework specific behaviors built around this fundamental test unit of `Assessment`.

## Why not use `test.list` from `go test` ?

The `test.list` is a great way to run the dry-run equivalent behavior. However, it is not easily extendable into the core of `e2e-framework` as
there are framework specific behavior such as `setup` and `teardown` workflows. 

That, in conjunction with how the `test.list` works, it is not possible to extract information such as the `assessments` in the feature using the `test.list` mode brings the need to introduce a framework specific `dry-run` mode that can work well with `test.list` while providing all
the framework specific benefits of how the Tests to be processed can be listed

## `--dry-run` mode
`e2e-framework` adds a new CLI flag that can be used while invoking the test called `--dry-run`. This works in conjunction with `test.list` to provide the following behavior. 

1. When the `--dry-run` mode is invoked No Setup/Teardown workflows are processed
2. Will display the Assessments as individual tests like they would be processed if not invoked with `--dry-run` mode
3. Skip all pre-post actions around the Before/After Features or Before/After Tests

When tests are invoked with `-test.list` argument, the `--dry-run` mode is automatically switched to enabled to make sure setup/teardown as well as the pre-post actions can be skipped.

## Example Output with `--dry-run`
```bash
❯ go test . -test.v -args --dry-run
=== RUN   TestPodBringUp
=== RUN   TestPodBringUp/Feature_One
=== RUN   TestPodBringUp/Feature_One/Create_Nginx_Deployment_1
=== RUN   TestPodBringUp/Feature_One/Wait_for_Nginx_Deployment_1_to_be_scaled_up
=== RUN   TestPodBringUp/Feature_Two
=== RUN   TestPodBringUp/Feature_Two/Create_Nginx_Deployment_2
=== RUN   TestPodBringUp/Feature_Two/Wait_for_Nginx_Deployment_2_to_be_scaled_up
--- PASS: TestPodBringUp (0.00s)
    --- PASS: TestPodBringUp/Feature_One (0.00s)
        --- PASS: TestPodBringUp/Feature_One/Create_Nginx_Deployment_1 (0.00s)
        --- PASS: TestPodBringUp/Feature_One/Wait_for_Nginx_Deployment_1_to_be_scaled_up (0.00s)
    --- PASS: TestPodBringUp/Feature_Two (0.00s)
        --- PASS: TestPodBringUp/Feature_Two/Create_Nginx_Deployment_2 (0.00s)
        --- PASS: TestPodBringUp/Feature_Two/Wait_for_Nginx_Deployment_2_to_be_scaled_up (0.00s)
PASS
ok  	sigs.k8s.io/e2e-framework/examples/parallel_features	0.353s
```

```bash
❯ go test . -test.v -args --dry-run --assess "Deployment 1"
=== RUN   TestPodBringUp
=== RUN   TestPodBringUp/Feature_One
=== RUN   TestPodBringUp/Feature_One/Create_Nginx_Deployment_1
=== RUN   TestPodBringUp/Feature_One/Wait_for_Nginx_Deployment_1_to_be_scaled_up
=== RUN   TestPodBringUp/Feature_Two
=== RUN   TestPodBringUp/Feature_Two/Create_Nginx_Deployment_2
    env.go:425: Skipping assessment "Create Nginx Deployment 2": name not matched
=== RUN   TestPodBringUp/Feature_Two/Wait_for_Nginx_Deployment_2_to_be_scaled_up
    env.go:425: Skipping assessment "Wait for Nginx Deployment 2 to be scaled up": name not matched
--- PASS: TestPodBringUp (0.00s)
    --- PASS: TestPodBringUp/Feature_One (0.00s)
        --- PASS: TestPodBringUp/Feature_One/Create_Nginx_Deployment_1 (0.00s)
        --- PASS: TestPodBringUp/Feature_One/Wait_for_Nginx_Deployment_1_to_be_scaled_up (0.00s)
    --- PASS: TestPodBringUp/Feature_Two (0.00s)
        --- SKIP: TestPodBringUp/Feature_Two/Create_Nginx_Deployment_2 (0.00s)
        --- SKIP: TestPodBringUp/Feature_Two/Wait_for_Nginx_Deployment_2_to_be_scaled_up (0.00s)
PASS
ok  	sigs.k8s.io/e2e-framework/examples/parallel_features	0.945s
```

## Example with `-test.list`
```bash
❯ go test . -test.v -test.list ".*" -args
TestPodBringUp
ok  	sigs.k8s.io/e2e-framework/examples/parallel_features	0.645s
```

As you can see from the above two examples, the output of the two commands are not really the same. Using `--dry-run` gives you a more framework specific behavior of how the tests are going to be processed in comparison to `-test.list`

