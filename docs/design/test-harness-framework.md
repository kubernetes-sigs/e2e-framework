# Go E2E Test Framework for Kubernetes

This document captures high-level design ideas for the next generation of a Go framework for testing components runnking on Kubernetes. The framework, referred to as `e2e-framework` provides ways that makes it easy to define tests functions that can programmatically test components running on a cluster.  The two overall goals of the new framework are to allow developers to quickly and easily assemble end-to-end tests and to provide a collection of support packages to help with interacting with the API-server.

> See the original Google Doc design document [here](https://docs.google.com/document/d/11JKqcnUOrw5Lk98f_ylJXBXyxWSW1z3CZu27OLX1CbM/edit).

## Motivations
Currently, the existing E2E Test Framework is intrinsically baked into Kubernetes source code repository with organically grown assumptions, over the years, that shape how tests are created and executed.  This, however, has presented some shortcomings including:

* Difficulty (or even the impossibility) of consuming existing test framework packages from outside of the Kubernetes project
* Misaligned framework design can cause dependencies on unnecessary resources
* The requirements to use Ginkgo as a way to write tests

## Goals

* Properly documented framework to help with adoption
* Easily express end-to-end test and suites using the built-in Go testing package
* Rely and use Go testing package constructs to create test suites and test functions
* Tests should be go test(able)
* Provide sensible programmatic API components to compose tests
* Make test packages easily consumable programmatically
* Provide helper functions that abstracts Client-Go functionalities (i.e. similar to those found in kubernetes/kubernetes/test/e2e/framework repo).
* Rely on built-in Go test features to easily select/filter tests to run during execution
* Define optional standardized env/argument variables to control/pass data to tests
* Avoid direct dependencies on upstream k8s.io/kubernetes packages
* Avoid dependencies on cluster components (i.e. cluster-providers, CSI, CNI, etc)
* Move away from reliance on Ginkgo or other non-standard Go based tests packages

## Non-Goals

* Does not specify how tests are built and executed (i.e. e2e.test binary)
* Maintain backward compatibility with existing e2e test framework
* Not responsible for bootstrapping or executing the tests themselves (i.e. Ginkgo)
* Initially, this framework will not be responsible for hosting clusters infra components (i.e. Kubetest/Kubetest2)
* Excludes support for fake/mock cluster components for testing

## Design
The initial design of the end-to-end test framework, defined here, will be heavily inspired and influenced by the Knative reconciler-test framework which is a lightweight framework to easily compose end-to-end Kubernetes behavior testing.  This new design introduces several components that can be used to compose end-to-end tests as depicted below:

* Environment
* Environment operations
* Test features
* Feature step functions

![](./e2e-environment-design.png)

### Environment
The environment component is the heart of the framework. It allows devlopers to define attributes and features to be tested (see Test Features below).  The environment uses callback functions to let developers implement customized behaviors before or after a test suite has been exercised.

The following shows a proposed type for the environment:

```go=
// EnvFunc is an operation applied during/after environment setup
type EnvFunc func(context.Context, klient.Config) (context.Context, error)

type Environment interface {
    // Config returns the associated klient.Config value
    Config() klient.Config
    
    // Context returns the environment context
    Context() context.Context
    
    // WithContext returns an environment with an updated context
    WithContext(context.Context) Environment
    
    // Setup registers environment operations that are executed once
    // prior to the environment being ready and prior to any test. 
    Setup(EnvFunc...)

    // BeforeTest registers funcs that are executed before each Env.Test(...)
    BeforeTest(EnvFunc...)
    
    //Test executes a test feature
    Test(*testing.T, features.Feature)
    
    // AfterTest registers funcs that are executed after each Env.Test(...)
    AfterTest(EnvFunc...)
	
    // Finish registers funcs that are executed at the end.
    Finish(EnvFunc...)

    // Launches the test suite from within a TestMain
    Run(*testing.M) int
}
```

#### Environment package
This design assumes that the `Environment` type and associated functions are hosted in a package called `env`.

#### Environment constructor functions
The `Environment` type could be created with the following constructor functions:
```go=
// env.New creates env.Environment with default context.Context 
// and a default klient.Config
env.New() 

// env.NewWithConfig creates env.Environment with a specified 
// klient.Config and a default context.Context
env.NewWithConfig(klient.Config)

// env.NewWithContext creates an environment with a specified
// klient.Config and context.Context
env.NewWithContext(context.Context, klient.Config)
```

### `env.Environment` and context.Context
Before an `Environment` can be used to run tests, it goes through several stages including configurations, feature definitions, and feature testing. During the entire lifetime of these cycles, a context can be used to inject control, signaling, or pass data into each phase.

> The context propagation strategy used in this design is simlar to how it is done in package [net/http](https://pkg.go.dev/net/http).


#### Propagating environment context
This framework is designed for a context to be injected early during the contstruction of the `Environment` value (see Constructon functions above). The context associated with the environment can then be propagated during the many execution stages of the test.

#### Updating the `Environment` context
In some instances, it will be necessary to update a previously injected context. To update an environment's context, after the environment has been already created, test writers would need to use `Environment.WithContext` method to update the context and get a new environment.

#### Accessing an `Environment`'s context
After an enviroment has been initialized, it's context can be accessed at any time using the `Environment.Context` method.

### Environment Operations
Test writers can specify an environment operation using type EnvFunc.

```go
type EnvFunc func(context.Context, klient.Config) (context.Context, error)
```

An operation can be applied at different stages during the life cycle of an environment including environment setup, before a test, after a test, and to tear down the environment.

### Test Features
A test feature represents a collection of testing steps that can be performed as a group for a given feature. A feature shall have the following attributes:

* *Name* - an arbitrary name that identifies the feature being tested.  The string value can be used as part of a filtering mechanism to execute features matching this value.
* *Labels* - string values used to label the feature (i.e. “alpha”, “conformance”, etc).
* *Steps* - actions that can be performed at different phases of a feature test.

A feature shall be represented by the Feature type as shown below:

```go
type Feature interface {
	// Name is a descriptive text for the feature
	Name() string
	// Labels are used to label the feature (beta, conformance, etc)
	Labels() Labels
    // Steps are feature operations (see Execution Step)
	Steps() []Step
}
```
#### Feature labels
A feature can receive an arbitrary label as a hint about the nature of the feature being tested.  For instance, the followings could be used as labels:

* `level`: alpha
* `conformance`: Beta
* `type`: network

A feature state shall be encoded using the following type and constant values:

```go
type Labels map[string]sring
```

### Execution steps
An execution step is a granular operation that can be declared and combined to perform user-defined actions for the feature.  An execution step has the following attributes:

* *Name* - a string describing the step
* *Step* function - implements user-defined operations to be executed by the step

The following interface shows a Step.:

```go
type Step interface {
    // Name is the step name
    Name() string
    // Func is the operation for the step
    Func() StepFunc
    // Level indicates a framework-defined level for the step
    Level() Level
}
```

#### The step function
The operation performed during an execution step is defined as the following Go function:

```go
type StepFunc func (context.Context, *testing.T, klient.Config) context.Context
```

When a step is executed, it will receive the last updated `context.Context`, `*testing.T` for test signaling, and the `klient.Config` that was used for the environment. Note a step function can update the context and return an updated context value that will be passed to subequent step.

#### Step levels
A step level identifies the type of a step.  It shall be encoded as the following type, shown below.

```go
type Level uint8

const (
    Setup Level = iota
    Assessment
    Teardown
)
```

## Using the Framework
When implemented, the framework will come with many packages along with helper types to facilitate the creation, running, and tearing down feature tests. This section shows an example of a possible implementation by outlining the steps necessary to create and run feature tests.

### Setting up a test suite
The starting point for an E2E test is a Go TestMain function which can launch all test functions in the package (or test suite).  The following snippet shows how a new `Environment` could be set up with using a provided `klient.Config` value.

```go=
import (
    "sigs.k8s.io/e2e-framework/pkg/env"
    conf "sigs.k8s.io/e2e-framework/klient/conf"
)

var (
    var global env.Environment
)

func TestMain(m *testing.M) {
    cfg, err := conf.New(conf.ResolveConfigFile())
    if err != nil {os.Exit(1)}
    global = env.NewWithConfig(cfg)
   ...
}
```

In the snippet above, the environment configuration instance is made global because the Go’s test framework has no easy way (via context or otherwise) of injecting values into a running test functions.

### Specifying environment operations
Using an `environment` instance, a test author shall be able to register operations (functions), of type `EnvFunc`, that are applied at different stages of the test lifecycle.  For instance, during environment setup, the `Setup` method can be used to register a function to set up environment resources.

```go=
import (
    "sigs.k8s.io/e2e-framework/pkg/env"
    conf "sigs.k8s.io/e2e-framework/klient/conf"
)

var (
    var global env.Environment
)

func TestMain(m *testing.M) {
    cfg, err := conf.New(conf.ResolveConfigFile())
    if err != nil {os.Exit(1)}
    global = env.NewWithConfig(cfg)
    
    global.Setup(func(context.Context, cfg klient.Config) (context.Context, error){
        // setup environment
        return nil, nil
    })
    ...
}
```

### Launching the suite
After the environment is configured, it is ready to be launched using the `Environment.Run` method which will execute all test functions defined in the current package where it is located.

```go
import (
    "sigs.k8s.io/e2e-framework/pkg/env"
    conf "sigs.k8s.io/e2e-framework/klient/conf"
)

var (
    var global env.Environment
)

func TestMain(m *testing.M) {
    cfg, err := conf.New(conf.ResolveConfigFile())
    if err != nil {os.Exit(1)}
    global = env.NewWithConfig(cfg)
    
    global.Setup(func(context.Context, cfg klient.Config) (context.Context, error){
        // setup environment
        return nil, nil
    })
    ...
    os.Exit(global.Run(m))
}
```

### Creating a test
To test a feature, the framework uses a plain Go test function.  Test authors  would simply create a test function as shown:

```go
func TestSomething(t *testing.T) {
    ...
}
```

Then within that function a test authors would declare the feature and use the previously created environment to test the features.

### Defining a new feature
Before a feature can be tested, it must be defined.  Assuming that the framework provides a package named `feature`, with a fluent API, to declare and set up a `feature`, the following snippet shows how this would work:

```go
func TestSomething(t *testing.T) {       
	feat := features.New("Hello Feature
        Setup(StartKind()).
        Assess("Simple test", func(ctx context.Context, t *testing.T, cfg klientcfg.Config) context.Context {
            result := "foo"
            if result != "Hello foo" {
                t.Error("unexpected message")
            }
            return ctx
        }).
        Teardown(StopKind())
        .Feature()   
...
}

func StartKind() feature.StepFunc  {
    return func(ctx context.Context, t *testing.T, cfg klientcfg.Config) {
        // check cluster is running on kind
    }
}

func StopKind() feature.StepFunc {
    return func(ctx context.Context, t *testing.T, cfg klientcfg.Config) {
        // check cluster is running on kind
    }
}

```

The snippet above creates feature `feat` with a label and includes several `step` functions with varying levels including `setup`, `assessment`, and `teardown`.

### Testing the feature
Next the feature can be tested.  This is done by invoking the `Test` method on the `global` environment.

```go
func TestSomething(t *testing.T) {
	feat := features.New("Hello Feature
        Setup(StartKind()).
        Assess("Simple test", func(ctx context.Context, t *testing.T, cfg klientcfg.Config) context.Context {
            result := "foo"
            if result != "Hello foo" {
                t.Error("unexpected message")
            }
            return ctx
        }).
        Teardown(StopKind())
        .Feature()
        
    global.Test(t, feat)
}
```

The environment component will run the feature test passing each execution step function a context, a `*testing.T` and`klient.conf.Config`.

### Finishing and clean up
After all tests in the package (or suite) are executed, the test framework will automatically trigger any teardown operation specified as `Environment.Finish` method, as shown below.

```go
func TestMain(m *testing.M) {
    ...
    env.Setup(func(ctx context.Context, cfg klientconf.Config) (context.Context, error){
        // setup environment
        return ctx, nil
    })

    env.Finish(func(ctx context.Context, cfg klientconf.Config) (context.Context, error){
        // setup environment
        return ctx, nil
    })
    ...
}
```

## Examples
The following shows actual examples of how the early implementation of this framework could work.

### Simple feature test

The following shows how a feature test can be defined outside of a test suite (TestMain).

```go=
func Hello(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

func TestHello(t *testing.T) {
    e := env.New()
    
    feat := features.New("Hello Feature").
        WithLabel("type", "simple").
        Assess("test message", func(ctx context.Context, t *testing.T, conf klientconf.Config) {
            result := Hello("foo")
            if result != "Hello foo" {
                t.Error("unexpected message")
            }
        }).Feature()
        
    e.Test(context.TODO(), t, feat)
}
```

### Test in a Suite
This example shows how a test can be setup in a suite (package test with TestMain).  First, the following shows the definition for TestMain. Note that for now, in order to make the same context visible to the rest of the test functions, it has to be declared as a global package variable, similar to the environment variable.

```go=
var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
    cfg, err := klientcfg.New(klientcfg.ResolveConfigFile)
    if err != nil { ... }
    
    ctx = context.WithValue(context.TODO(), 1, "bazz")

    testenv = env.NewWithContext(ctx, cfg)
    testenv.BeforeTest(func(ctx context.Context, cfg klientcfg.Config) (context.Context, error) {
        return ctx, nil
    })
    
    // run suite
    os.Exit(testenv.Run(m))
}
```

The next code snippet shows how the test function could be defined to be used in the suite:

```go=
func Hello(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

func TestHello(t *testing.T) {
    feat := features.New("Hello Feature").
        WithLabel("type", "simple").
        Assess("test message", func(ctx context.Context, t *testing.T) context.Context{
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

## Running the tests
The test framework is designed to completely rely on the Go test tool and its functionalities. Running a test shall only require constructs supported by the Go’s compiler and test framework tools.

### Tagging test functions
If a feature tests are part of a larger code base with many different types of tests (i.e. unit), it could be useful to tag the MainTest function with an arbitrary tag as follows:

```go=
// +build app1-e2e

var (
    var envCfg environment.EnvConfig
)

func TestMain(m *testing.M) {
    cfg, err := conf.New(conf.ResolveConfigFile())
    if err != nil {os.Exit(1)}
    global = env.NewWithConfig(cfg)

    ...
    os.Exit(global.Run(m))
}
```

When running the test (or building a binary), only the e2e tests will be selected:

```go
go test -timeout=15m -tags=app1-e2e ./test/...
```

### Filtering feature tests
The test framework should provide mechanism to filter (in or out) features to be tested based on the following criteria:

For instance, the following shows a test being filtered to exercise only features with name “special”
```go
go test -tags=e2e ./test/... --feature="special"
```

The framework shall provide predefined flagsets that will be automatically applied during test execution. Possible filters that could be supported by the framework implementation:

* `--feature` - a regular expression that target features to run
* `--assert` - a regular expression that targets the name of an assertive steps
* `--labels` - a comma-separated list of key/value pairs  used to filter features by their assigned labels

The framework should automatically inject these filter values into the environment component when it is created.

### Skipping features
The test framework should provide the ability to explicitly exclude features during a test run.  This could be done with the following flags:

* `--skip-feature` - a regular expression that skips features with matching names
* `--skip-assert` - a regular expression that skips assertions with matching name
* `--skip-lables` - a comma-separated list of key/value pairs used to skip features with matching lables

## Test support
Another important aspect of the test framework is to make available a collection of support packages to help with the interaction with the Kubernetes API server and its hosted objects. The test framework shall offer a helper package (see klient package) that offers types and functions to make it easier for developers to write tests.

For instance, with the `klient` package, the following could be a list of pre-defined functions that can be used during environment setup:

* `klient.ApplyYamlFile(YamlFilePath)`
* `klient.DeleteWithYamlFile(yamlFilePath)`
* `klient.ApplyYamlText(string)`
* `klient.RunPod(imageName)`
* `klient.CreateGenericSecret(...)`
* `klient.CreateFilesSecret(...)`
* `klient.DeleteSecret(...)`

Etc.

## Resources
* E2E Test Framework 2 Repo - https://github.com/kubernetes-sigs/e2e-framework
* Helper package for test - https://docs.google.com/document/d/1ZtN58kU8SKmDDkxeBKxR9Un76eqhszCVcZuqhs-fLIU/edit#
* Official E2E Test Framework  doc - https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/e2e-tests.md
* Kubernetes.io blog on using the current test framework  - https://kubernetes.io/blog/2019/03/22/kubernetes-end-to-end-testing-for-everyone/
* Knative Reconciler Test Framework - https://github.com/knative-sandbox/reconciler-test
* Contour Integration Test Harness - https://github.com/projectcontour/contour/issues/2222
* Integration Tester for Kubernetes - https://github.com/projectcontour/integration-tester
* https://github.com/kubernetes-sigs/ingress-controller-conformance
