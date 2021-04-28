# E2E Framework
An experimental Go framework for E2E testing of Kubernetes cluster components.

The primary goal of this project is to provide a `go test`(able) 
framework to uses Go's existing testing API to define end-to-end tests suites 
that can be used to test Kubernetes components. Some additional goals 
include:

* Provide a sensible programmatic API to compose tests 
* Leverage Go's testing API to compose test suites
* Expose packages that are easy to programmatically consume
* Collection of helper functions that abstracts Client-Go functionalities
* Rely on built-in Go test features to easily select/filter tests to run during execution
* And more

For more detail, see the [design document](https://docs.google.com/document/d/11JKqcnUOrw5Lk98f_ylJXBXyxWSW1z3CZu27OLX1CbM/edit?usp=sharing).

## Community, discussion, contribution, and support

Learn how to engage with the Kubernetes community on the [community page](http://kubernetes.io/community/).

You can reach the maintainers of this project at:

- [Slack](https://kubernetes.slack.com/messages/sig-testing)
- [Mailing List](https://kubernetes.slack.com/messages/sig-testing)

### Code of conduct

Participation in the Kubernetes community is governed by the [Kubernetes Code of Conduct](code-of-conduct.md).
