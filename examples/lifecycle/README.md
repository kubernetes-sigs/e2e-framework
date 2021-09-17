# Test Feature Lifecycle

The example in this package highlights the lifecycle of a feature test as supported by the framework.
At runtime (or test time), a test goes through multiple stages as listed below: 
```
- env.Setup
- env.BeforeEachTest
    env.Test {
        - env.BeforeEachFeature
          - feature.Feature.Setup
          - feature.Assessment
          - feature.Teardown
        - env.AfterEachFeature
    }
- env.AfterEachTest
- env.Finish
```

The framework API allows test writers to override each stage by defining a function that is executed at the appropriate stage.
