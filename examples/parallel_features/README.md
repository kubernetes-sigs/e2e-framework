# Parallel Run of Test Features

This directory contains the example of how to run the features in parallel as part of your test run using
the right flags to the test runtime

# Running tests with flags

These tests can be executed using the normal `go test` command by passing the right arguments

```bash
go test -v . -args --parallel
```

One can also build a test binary and run that with the right arguments.

```bash
go test -c -o parallel.test .
./parallel.test --parallel
```
