/*
Copyright 2024 The Kubernetes Authors.

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

package stepfuncs

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/utils"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
	"sigs.k8s.io/e2e-framework/support"
)

// PerformNodeOperation returns a step function that performs a node operation on a cluster.
// This can be integrated as a setup function for a feature in question before the feature
// is tested.
func PerformNodeOperation(action support.NodeOperation, node *support.Node, args ...string) types.StepFunc {
	return func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
		t.Helper()

		err := utils.PerformNodeLifecycleOperation(ctx, action, node, args...)
		if err != nil {
			t.Fatalf("failed to perform node operation: %v", err)
		}
		return ctx
	}
}
