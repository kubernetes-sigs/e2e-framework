# Test Suite Example

The examples in this package shows how to setup tests with suite (or package) 
test configuration. The suite is configured in `main_test.go` and it creates a 
global package variable `testenv` to store the test environment being configured.

```go
var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
	var err error
	testenv, err = env.NewWithContext(context.WithValue(context.Background(), 1, "bazz"), envconf.New())
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(testenv.Run(m))
}
```
The test functions access the test environment `testenv` to define the feature 
tests and launch the test itself as shown below:

```go
func TestHello(t *testing.T) {
	feat := features.New("Hello Feature").
		WithLabel("type", "simple").
		Assess("test message", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			name := ctx.Value(1).(string)
			result := Hello(name)
			if result != "Hello bazz" {
				t.Error("unexpected message")
			}
			return ctx
		}).Feature()

	testenv.Test(t, feat)
}
```
