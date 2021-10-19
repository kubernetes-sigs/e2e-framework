# Sonobuoy example (without test suites)

This package shows how the test framework can be used with an in-cluster config. This is a common
configuration since we may want to launch these tests from within the cluster itself, such as within
a [Sonobuoy][sonobuoy] [plugin][sonobuoy-plugins].

The `testenv` is a global that we rely on and it is set up in main_test.go
```go
func TestMain(m *testing.M) {
    // Empty string results in in-cluster config.
    testenv = env.NewInclusterConfig()
    os.Exit(testenv.Run(m))
}
```

Our test features will then have access to that in-cluster configuration (and klient). We do not need to instantiate new
envs or configs for any of our tests.
```go
// The following shows an example of a simple
// test function that reaches out to the API server.
func TestAPICall(t *testing.T) {
    feat := features.New("API Feature")...
        ...
        ...

        if err := c.Client().Resources("kube-system").List(ctx, &pods); err != nil {
            t.Error(err)
        }

        ...
        ...
    }).Feature()
    testenv.Test(t, feat)
}
```

[sonobuoy]: https://www.github.com/vmware-tanzu/sonobuoy
[sonobuoy-plugins]: https://www.github.com/vmware-tanzu/sonobuoy-plugins