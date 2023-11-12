# Testing Kubernetes Controllers and Operators with the E2E-Framework
This repository shows how to use the Kubernetes-SIGs/[e2e-framework](https://github.com/kubernetes-sigs/e2e-framework) write Go tests to do end-to-end testing of Kubernetes controllers/operators.

It uses the [CronJob](https://book.kubebuilder.io/cronjob-tutorial/cronjob-tutorial) controller,  from [Kubebuilder's](https://github.com/kubernetes-sigs/kubebuilder) tutorial, to show how to write effective end-to-end tests using the e2e-framework.

## Getting Started
The e2e-framework Go tests can be found in the [./e2e-test](./e2e-test/) directory.

To run the test, use the `go test` command:

```bash
go test -v ./e2e-test 
```

## License

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

