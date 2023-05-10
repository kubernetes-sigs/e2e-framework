# Custom Namespaces For Every Test

This example shows you how to use the env hooks in order to set up
a custom namespace for every test. This could easily be done for every
feature as well if that is your preference.

First, you'll have to set up the env. In this example we assume an in-cluster configuration.

```go
var testenv env.Environment

func TestMain(m *testing.M) {
 testenv = env.New()
 ...
}
```

Second, set the BeforeEachTest hook to create the namespace. We store it in the
context so that it can be looked up on a per-test basis for deletion.

```go
testenv.BeforeEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
    return createNSForTest(ctx,cfg,t,runID)
})
...
// The creation uses the typical c.Resources() object.
cfg.Client().Resources().Create(ctx,&nsObj)
```

Third, set the AfterEachTest hook to lookup and delete the namespace.

```go
testenv.AfterEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
    return deleteNSForTest(ctx,cfg,t,runID)
})
...
// The deletion uses the typical c.Resources() object.
cfg.Client().Resources().Delete(ctx,&nsObj)
```

Forth, in your test you can lookup the namespace and casting to a string

```go
namespace := ctx.Value(GetNamespaceKey(t)).(string)
```

So, tying it all together, the `TestMain` looks like this:

```go
func TestMain(m *testing.M) {
 testenv = env.New()

 // Specifying a run ID so that multiple runs wouldn't collide.
 runID := envconf.RandomName("", 4)

 /* Skipping cluster creation for brevity */
 
 testenv.BeforeEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
  return createNSForTest(ctx, cfg, t, runID)
 })
 testenv.AfterEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
  return deleteNSForTest(ctx, cfg, t, runID)
 })

 os.Exit(testenv.Run(m))
}
```
