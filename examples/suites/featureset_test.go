/*
Copyright 2021 The Kubernetes Authors.

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

package suites

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// TestFeatureSet shows how Environment.Test can be used to
// test a collections of features (feature set). The example
// also shows the before/after feature actions which causes
// callbacks functions to be executed during the feature tests.
func TestFeatureSet(t *testing.T) {
	f1 := features.New("bazz test").
		Assess("Hello Bazz", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello("bazz")
			if result != "Hello bazz" {
				t.Error("unexpected message")
			}
			return ctx
		}).Feature()

	f2 := features.New("batt test").
		Assess("Hello Batt", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello("batt")
			if result != "Hello batt" {
				t.Error("unexpected message")
			}
			return ctx
		}).Feature()

	testenv.Test(t, f1, f2)
}
