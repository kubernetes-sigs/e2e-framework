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

package setupfail

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func TestMain(m *testing.M) {
	var testEnv env.Environment

	tests := []struct {
		name             string
		testEnvGenerator func(context.Context, *[]string) env.Environment
		expected         []string
	}{
		{
			name: "No test setup failures, do all setup and finish actions",
			testEnvGenerator: func(ctx context.Context, actions *[]string) env.Environment {
				testEnv = env.New()

				testEnv.Setup(
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed setup 1")
						return ctx, nil
					},
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed setup 2")
						return ctx, nil
					},
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed setup 3")
						return ctx, nil
					},
				)
				testEnv.Finish(
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed finish 1")
						return ctx, nil
					},
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed finish 2")
						return ctx, nil
					},
				)

				return testEnv
			},
			expected: []string{"completed setup 1", "completed setup 2", "completed setup 3", "completed finish 1", "completed finish 2"},
		},
		{
			name: "Skip remaining setup actions after one fails, do finish actions",
			testEnvGenerator: func(ctx context.Context, actions *[]string) env.Environment {
				testEnv = env.New()

				testEnv.Setup(
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed setup 1")
						return ctx, nil
					},
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed setup 2")
						return ctx, fmt.Errorf("failed setup 2")
					},
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed setup 3")
						return ctx, nil
					},
				)
				testEnv.Finish(
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed finish 1")
						return ctx, nil
					},
					func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
						*actions = append(*actions, "completed finish 2")
						return ctx, nil
					},
				)

				return testEnv
			},
			expected: []string{"completed setup 1", "completed setup 2", "completed finish 1", "completed finish 2"},
		},
	}

	for _, test := range tests {
		var actions []string
		testEnv = test.testEnvGenerator(context.TODO(), &actions)
		_ = testEnv.Run(m)

		readableActions := strings.Join(actions, ", ")
		if len(actions) != len(test.expected) {
			klog.Fatalf("Expected slice of %v, but got %s", test.expected, readableActions)
		}

		for i := range actions {
			if actions[i] != test.expected[i] {
				klog.Fatalf("Expected slice of %v, but got %s", test.expected, readableActions)
			}
		}

		klog.Infof("PASS: %s\n\tActions completed: %+v", test.name, readableActions)
	}
}
