# Simple examples (without test suites)

As the title implies, this package shows how the test framework can be used
directly in test functions without setting up a test suite in a `TestMain` function.

```go
func TestHello_WithSetup(t *testing.T) {
	e := env.NewWithConfig(envconf.New())
	var name string
	feat := features.New("Hello Feature").
		WithLabel("type", "simple").
		Setup(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			name = "foobar"
			return ctx
		}).
		Assess("test message", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello(name)
			if result != "Hello foobar" {
				t.Error("unexpected message")
			}
			return ctx
		}).Feature()

	e.Test(t, feat)
}
```