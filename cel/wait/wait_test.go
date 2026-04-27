/*
Copyright 2026 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package wait

import "testing"

// The Match/MatchAny functions rely on *resources.Resources to fetch
// from a live API server, which is covered by the integration test in
// examples/cel. These unit tests exercise only the isTerminal helper —
// the logic that decides whether a CEL error halts polling.

func TestIsTerminal_compileErrorTerminates(t *testing.T) {
	err := mockError("cel: compile \"x >\": syntax error")
	if !isTerminal(err) {
		t.Fatal("compile error should terminate polling")
	}
}

func TestIsTerminal_typeErrorTerminates(t *testing.T) {
	err := mockError("cel: expression \"1+1\" returned types.Int, want bool")
	if !isTerminal(err) {
		t.Fatal("type error should terminate polling")
	}
}

func TestIsTerminal_assertionFailureRetries(t *testing.T) {
	err := mockError("cel: assertion failed: object.status.ready == object.spec.replicas")
	if isTerminal(err) {
		t.Fatal("assertion failure should retry, not terminate")
	}
}

func TestIsTerminal_nil(t *testing.T) {
	if isTerminal(nil) {
		t.Fatal("nil error should not be terminal")
	}
}

type mockError string

func (m mockError) Error() string { return string(m) }
