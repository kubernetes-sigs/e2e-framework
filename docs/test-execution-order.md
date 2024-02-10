# E2E-Framework Test Execution Order

When an e2e-framework test executes, it follows a specific order of execution. You can register environment callback functions (defined [here](https://github.com/kubernetes-sigs/e2e-framework/blob/3883d11ff321f48c145ea6696ab27be0636330e5/pkg/internal/types/types.go#L49)) to add your own functionalities around the test code (the framework comes with some [pre-defined environment functions](https://github.com/kubernetes-sigs/e2e-framework/tree/main/pkg/envfuncs) that are used for different purposes).

The following source code example registers an environment callback function for each supported step of the test execution.
The example is desigined to highlight the order of execution when the framework runs its tests.

```go
package log_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()

	testenv.Setup(func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		fmt.Println("* Setting up e2e test...")
		return ctx, nil
	})

	testenv.BeforeEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		fmt.Printf("      --> Executing BeforeTest: test %s \n", t.Name())
		return ctx, nil
	})

	testenv.BeforeEachFeature(func(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature) (context.Context, error) {
		fmt.Printf("          > Executing BeforeFeature: %s \n", f.Name())
		return ctx, nil
	})

	testenv.AfterEachFeature(func(ctx context.Context, cfg *envconf.Config, t *testing.T, f features.Feature) (context.Context, error) {
		fmt.Printf("          > Executing AfterFeature: %s \n", f.Name())
		return ctx, nil
	})

	testenv.AfterEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		fmt.Printf("      --> Executing AfterTest: %s \n", t.Name())
		return ctx, nil
	})

	testenv.Finish(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
		fmt.Println("* Finishing e2e test ")
		return ctx, nil
	})

	os.Exit(testenv.Run(m))
}

func TestSomething(t *testing.T) {

	// executes testenv.BeforeEachFeature here
	f1 := features.New("Feature 1").
		Assess("Assessment 1", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			return ctx
		})
	// executes testenv.AfterEachFeature here

	// executes testenv.BeforeEachFeature here
	f2 := features.New("Feature 2").
		Assess("Assessment 2", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			return ctx
		})
	// executes testenv.AfterEachFeature here

	// executes testevn.BeforeEachTest
	testenv.Test(t, f1.Feature(), f2.Feature())
	// executes testenv.AfterEachTest
}
```

When you run the test above, you should see the registered function callbacks getting called at each step of the test execution:

```
go test -v .
* Setting up e2e test...
=== RUN   TestSomething
      --> Executing BeforeTest: test TestSomething 
          > Executing BeforeFeature: Feature 1 
=== RUN   TestSomething/Feature_1
=== RUN   TestSomething/Feature_1/Assessment_1
          > Executing AfterFeature: Feature 1 
          > Executing BeforeFeature: Feature 2 
=== RUN   TestSomething/Feature_2
=== RUN   TestSomething/Feature_2/Assessment_2
          > Executing AfterFeature: Feature 2 
      --> Executing AfterTest: TestSomething 
--- PASS: TestSomething (0.00s)
    --- PASS: TestSomething/Feature_1 (0.00s)
        --- PASS: TestSomething/Feature_1/Assessment_1 (0.00s)
    --- PASS: TestSomething/Feature_2 (0.00s)
        --- PASS: TestSomething/Feature_2/Assessment_2 (0.00s)
PASS
* Finishing e2e test 
ok      e2e-framework/workbench 0.662s
```
