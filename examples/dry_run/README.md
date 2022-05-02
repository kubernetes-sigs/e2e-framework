# Dry Run of Test Features

This directory contains the example of how to run the test features in `dry-run` mode using framework specific flags.

# Run Tests with flags

These test cases can be executed using the normal `go test` command by passing the right arguments

```bash
go test -v . -args --dry-run
```

With the output generated as following.

```bash
=== RUN   TestDryRunOne
=== RUN   TestDryRunOne/F1
=== RUN   TestDryRunOne/F1/Assessment_One
=== RUN   TestDryRunOne/F2
=== RUN   TestDryRunOne/F2/Assessment_One
=== RUN   TestDryRunOne/F2/Assessment_Two
--- PASS: TestDryRunOne (0.00s)
    --- PASS: TestDryRunOne/F1 (0.00s)
        --- PASS: TestDryRunOne/F1/Assessment_One (0.00s)
    --- PASS: TestDryRunOne/F2 (0.00s)
        --- PASS: TestDryRunOne/F2/Assessment_One (0.00s)
        --- PASS: TestDryRunOne/F2/Assessment_Two (0.00s)
=== RUN   TestDryRunTwo
=== RUN   TestDryRunTwo/F1
=== RUN   TestDryRunTwo/F1/Assessment_One
--- PASS: TestDryRunTwo (0.00s)
    --- PASS: TestDryRunTwo/F1 (0.00s)
        --- PASS: TestDryRunTwo/F1/Assessment_One (0.00s)
PASS
ok  	sigs.k8s.io/e2e-framework/examples/dry_run	0.618s
```

Without the `--dry-run` mode you will see the additional log `Do not run this when in dry-run mode` getting printed onto your terminal.

In order to integrate this into the `test.list`, please run the following

```bash
go test -v -list .
```

Which generates the output as following

```bash
TestDryRunOne
TestDryRunTwo
ok  	sigs.k8s.io/e2e-framework/examples/dry_run	0.375s
```

To understand the difference in Output, please refer to the [Design Document](../../docs/design/dry-run-mode.md)
