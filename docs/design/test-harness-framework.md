# Go E2E Test Framework for Kubernetes

This document captures high-level design ideas for the next generation of a Go framework for testing components runnking on Kubernetes. The framework, referred to as `e2e-framework` provides ways that makes it easy to define tests functions that can programmatically test components running on a cluster.  The two overall goals of the new framework are to allow developers to quickly and easily assemble end-to-end tests and to provide a collection of support packages to help with interacting with the API-server.


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

* Config
* Environment
* Environment operations
* Test features
* Feature step functions

![](./e2e-environment-design.png)

### Config
The Config type represents a configuration that must take place to establish a test environment. The configuration type shall expose a fluent API to allow developers to create and configure an environment prior to starting a test as shown below.

```go
type Config struct{
    namespace string
    kubeconfig *rest.Config
    env Environment
}

// Fluent configuration API methods
func (c *Config) WithKubeConfig(cfg *rest.Config) {}
func (c *Config) WithNamespace(ns string) {}

func (c *Config) Namespace() string {}
func (c *Config) KubeConfig() *rest.Config {}
func (c *Config) Client() kubernetes.Interface {}
func (c *Config) DynaClient() dynamic.Interface {}

// builds an environment from config
func (c *Config) Env() (Environment, error) {}
```

### Setup a configuration
The Config type could expose several methods that are designed to configure the environment using a fluent (builder) API approach. For instance the following shows what that could look like:

```go
 cfg := New().WithKubeConfig(aRestConfig)
```

### Environment
The environment component provides hooks to define a test environment that can be used to run test features (see Test Features below).  The environment uses callback functions to let developers specify customized behaviors before or after a test suite has been exercised.

The following shows a proposed type for the environment:
// EnvFunc is an operation applied during/after environment setup
type EnvFunc func(context.Context) (context.Context, error)

```go
type Environment interface {
    // Config returns the associated configuration
    Config() *Config
        
    // Setup registers environment operations that are executed once
    // prior to the environment being ready and prior to any test. 
    Setup(EnvFunc...)

    // BeforeTest registers funcs that are executed before each Env.Test(...)
    BeforeTest(EnvFunc...)
    
    //Test executes a test feature
    Test(context.Context, *testing.T, features.Feature)
    
    // AfterTest registers funcs that are executed after each Env.Test(...)
    AfterTest(EnvFunc...)
	
    // Finish registers funcs that are executed at the end.
    Finish(EnvFunc...)

    // Launches the test suite from within a TestMain
    Run(context.Contex, *testing.M) int
}
```

### Environment Operations
Test writers can specify an environment operation using type EnvFunc.

```go
type EnvFunc func(context.Context) (context.Context, error)
```

An operation can be applied at different stages during the life cycle of an environment including environment setup, before a test, after a test, and to tear down the environment.

### Test Features
A test feature represents a collection of testing steps that can be performed as a group for a given feature. A feature shall have the following attributes:

* *Name* - an arbitrary name that identifies the feature being tested.  The string value can be used as part of a filtering mechanism to execute features matching this value.
* *Labels* - string values used to label the feature (i.e. “alpha”, “conformance”, etc).
* *Steps* - actions that can be performed at different feature lifecycle phases.

A feature shall be represented by the Feature type as shown below:

```go
type StepFunc func (context.Context, *testing.T) context.Context

type Feature interface {
	// Name is a descriptive text for the feature
	Name() string
	// Labels are used to label the feature (beta, conformance, etc)
	Labels() Labels
	// Steps are feature operations 
	Steps() []Step
}
```

### Feature labels
A feature can receive an arbitrary label as a hint about the nature of the feature being tested.  For instance, the followings could be used as labels:

* `level`: alpha
* `conformance`: Beta
* `type`: network

A feature state shall be encoded using the following type and constant values:

```go
type Labels map[string]sring
```

### Execution Step
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

A step level shall be encoded as the following type, shown below.

```go
type Level uint8

const (
    Setup Level = iota
    Assessment
    Teardown
)
```

### The step function
The operation performed during an execution step can be defined using a Go function of type StepFn as shown below:

```go
type StepFunc func (context.Context, *testing.T, *Config)
```

## Using the Framework
When implemented, the framework will come with numerous packages along with helper types to facilitate the creation, running, and tearing down feature tests. This section shows an example of a possible implementation by outlining the steps necessary to create and run feature tests.

### Setting up a test suite
The starting point for an E2E test is a Go TestMain function which can launch all test functions in the package (or test suite).  Assuming there exists a config package offered by the framework, the following snippet shows how a new test could be set up by declaring a global environment configuration.

```go
var (
    var global env.Environment
)

func TestMain(m *testing.M) {
    cfg := config.New().WithKubeConfig(someConfig)
    global = env.NewWithConfig(cfg)
   ...
}
```

In the snippet above, the environment configuration instance is made global because the Go’s test framework has no easy way (via context or otherwise) of injecting values into a running test.

### Specifying environment operations
Once a configuration is created, it can be used to access an environment.  Using an instance of the environment, a test author shall be able to register operations (functions), of type EnvFunc, that are applied at different stages of the test lifecycle.  For instance, during environment setup, the Setup method can be used to register a function to set up environment resources.

```go
var (
    var global *config.Config
)

func TestMain(m *testing.M) {
    cfg := config.New().WithKubeConfig(someConfig)
    global = env.NewWithConfig(cfg)
    env.Setup(func(ctx context.Context){
        // setup environment
        return nil
    })
    ...
}
```

### Launching the suite
After the environment is configured, it is ready to be launched. This is done using the Run method which will execute all test functions defined in the current package where it is located.

```go
var (
    var global *config.Config
)

func TestMain(m *testing.M) {
    cfg, err := config.New(restCfg)
    global = cfg
    cfg.Env().Setup(func(ctx context.Context, cfg *Config){
        // setup environment
        return nil
    })

    os.Exit(env.Run(m))
}
```

### Creating a test
To test a feature, you would need to create a Go test function firstly:

```go
func TestSomething(t *testing.T) {
    ...
}
```

Then within that function a test writer would declare the feature and create the environment to test the feature as outlined below.

### Defining a new feature feature
Before a feature can be tested, it must be defined.  Assuming that the framework provides a package named `feature` with fluent API to declare and set up a feature, the following snippet shows how this could work:

```go
func TestSomething(t *testing.T) {
       feat := feature.Name(name).Label(feature.Alpha)
       Setup(func(ctx context.Context, t *testing.T, cfg *Config){
            t.Log("setting up something")
       }).
       Setup(IsKindServer()).
       Assess("this test",func(ctx context.Context, t *testing.T){
            t.Log("testing this")
       }).
       Teardown(feature.DefaultTeardown())
...
}

func IsKindServer() feature.StepFunc  {
    return func(ctx context.Context, t *testing.T) {
        // check cluster is running on kind
    }
}
```

The snippet above creates feature `feat` with a label and includes several `step` functions with varying levels including `setup`, `assessment`, and `teardown`.

### Testing the feature
Next the feature can be tested.  This can be done by accessing the environment stored in the global configuration variable global.

```go
func TestSomething(t *testing.T) {
    feat := feature.Name(name).Label(feature.Alpha)
       Setup(func(ctx context.Context, t *testing.T, cfg *Config){
            t.Log("setting up something")
       }).
       Setup(IsKindServer()).
       Assess("this test",func(ctx context.Context, t *testing.T, cfg *Config){
            t.Log("testing this")
       }).
       Teardown(feature.DefaultTeardown())
 
    global.Test(ctx, t, feat)
}
```

The environment component will run the feature test passing each feature an instance of testing parameter t.  Features are expected to use variable t as the only way to signal failure.

### Finishing and clean up
After all tests in the package (or suite) are executed, the test framework will automatically trigger any teardown operation specified as Env.Finish method in the TestMain function as shown below.

```go
func TestMain(m *testing.M) {
    ...
    env.Setup(func(ctx context.Context, cfg *Config){
        // setup environment
        return nil
    })

    env.Finish(func(ctx context.Context, cfg *Config){
        // setup environment
        return nil
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
    e := env.New(conf.New())
    feat := features.New("Hello Feature").
        WithLabel("type", "simple").
        Assess("test message", func(ctx context.Context, t *testing.T, config conf.Config) {
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
	ctx = context.Background()
)

func TestMain(m *testing.M) {
    ctx = context.WithValue(ctx, 1, "bazz")
    testenv = env.New()
    testenv.BeforeTest(func(ctx context.Context) (context.Context, error) {
        return ctx, nil
    })
    testenv.Run(ctx, m)
}
```

The next code snippet shows how the test function could be defined to be used in the suite

```go=
func Hello(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

func TestHello(t *testing.T) {
    name := ctx.Value(1).(string)
    feat := features.New("Hello Feature").
        WithLabel("type", "simple").
        Assess("test message", func(ctx context.Context, t *testing.T) context.Context{
            result := Hello(name)
            if result != "Hello bazz" {
                t.Error("unexpected message")
            }
        return ctx
    }).Feature()

    testenv.Test(ctx, t, feat)
}
```

## Running the tests
The test framework is designed to completely rely on the existing Go’s test functionalities and tooling. Running a test shall only require constructs supported by the Go’s compiler and test framework tools.

### Tagging test functions
If a feature tests are part of a larger code base with many different types of tests (i.e. unit), it could be useful to tag the MainTest function with an arbitrary tag as follows:

```go=
// +build app1-e2e

var (
    var envCfg environment.EnvConfig
)

func TestMain(m *testing.M) {
    cfg := config.New().WithKubeConfig(someConfig)
    env, err := cfg.Env()
    ...
    os.Exit(env.Run(ctx, m))
}
```

When running the test (or building a binary), only the e2e tests will be selected:

```go
go test -timeout=15m -tags=app1-e2e ./test/...
```

### Filtering feature tests
The test framework should provide mechanism to filter (in or out) features to be tested based on the following criteria:

* Feature name
* Feature state (i.e. alpha, beta, stable)
* Step assertion kind (i.e. must, mustNot, should, shouldNot, etc)

For instance, the following shows a test being run to exercise only features with name “special”
```go
go test -tags=e2e ./test/... --feature="special" --assert "assert step name"
```

The framework shall provide predefined flagsets that will be automatically applied during test execution. Possible filters that could be supported by the framework implementation:

* `--feature` - a regular expression that target features to run
* `--assert` - a regular expression that targets the name of an assertive steps
* `--labels` - a comma-separated list of key/value pairs  used to filter features by their assigned labels

The framework should automatically inject these filter values into the environment component when it is created.

## Test support
Another important aspect of the test framework is to make available a collection of support packages to help with the interaction with the Kubernetes API server and its hosted objects. The test framework shall come with a set of  predefined environment operations, for common use cases, to make it easier for developers to write tests.

For instance, if we assume there is a `cluster` package, the following could be a list of framework-provided support functions that can be used during environment setup:

* `cluster.ApplyYamlFile(YamlFilePath)`
* `cluster.DeleteWithYamlFile(yamlFilePath)`
* `cluster.ApplyYamlText(string)`
* `cluster.RunPod(imageName)`
* `cluster.CreateGenericSecret(...)`
* `cluster.CreateFilesSecret(...)`
* `cluster.DeleteSecret(...)`

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
