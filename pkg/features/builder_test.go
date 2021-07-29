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

package features

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

func TestNew(t *testing.T) {
	b := New("test-feat")
	if b.feat == nil {
		t.Error("builder has nil feature")
	}
	if b.feat.Name() != "test-feat" {
		t.Error("unexpected feature name set:", b.feat.Name())
	}
}

func TestFeatureBuilder(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T) types.Feature
		eval  func(*testing.T, types.Feature)
	}{
		{
			name: "empty feature",
			setup: func(t *testing.T) types.Feature {
				return New("empty").Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				if len(ft.labels) != 0 {
					t.Error("unexpected labels len:", len(ft.labels))
				}
				if len(ft.Steps()) != 0 {
					t.Error("unexpected number of steps:", len(ft.Steps()))
				}
			},
		},
		{
			name: "with labels",
			setup: func(t *testing.T) types.Feature {
				return New("test").WithLabel("a", "b").WithLabel("c", "d").Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				if len(ft.labels) != 2 {
					t.Error("unexpected labels len:", len(ft.labels))
				}
			},
		},
		{
			name: "one setup",
			setup: func(t *testing.T) types.Feature {
				return New("test").Setup(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				setups := GetStepsByLevel(ft.Steps(), types.LevelSetup)
				if len(setups) != 1 {
					t.Errorf("unexpected number of setup functions: %d", len(setups))
				}
				if len(ft.Steps()) != 1 {
					t.Errorf("unexpected number of steps %d", len(ft.Steps()))
				}
			},
		},
		{ // nolint
			name: "multiple setups",
			setup: func(t *testing.T) types.Feature {
				return New("test").Setup(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Setup(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				setups := GetStepsByLevel(ft.Steps(), types.LevelSetup)
				if len(setups) != 2 {
					t.Errorf("unexpected number of setup functions: %d", len(setups))
				}
				if len(ft.Steps()) != 2 {
					t.Errorf("unexpected number of steps %d", len(ft.Steps()))
				}
			},
		},
		{
			name: "one teardown",
			setup: func(t *testing.T) types.Feature {
				return New("test").Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				teardowns := GetStepsByLevel(ft.Steps(), types.LevelTeardown)
				if len(teardowns) != 1 {
					t.Errorf("unexpected number of teardown functions: %d", len(teardowns))
				}
				if len(ft.Steps()) != 1 {
					t.Errorf("unexpected number of steps %d", len(ft.Steps()))
				}
			},
		},
		{ // nolint
			name: "multiple teardowns",
			setup: func(t *testing.T) types.Feature {
				return New("test").Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				teardowns := GetStepsByLevel(ft.Steps(), types.LevelTeardown)
				if len(teardowns) != 2 {
					t.Errorf("unexpected number of setup functions: %d", len(teardowns))
				}
				if len(ft.Steps()) != 2 {
					t.Errorf("unexpected number of steps %d", len(ft.Steps()))
				}
			},
		},
		{
			name: "single assessment",
			setup: func(t *testing.T) types.Feature {
				return New("test").Assess("Some test", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				assessments := GetStepsByLevel(ft.Steps(), types.LevelAssess)
				if len(assessments) != 1 {
					t.Errorf("unexpected number of assessment function: %d", len(assessments))
				}
				if len(ft.Steps()) != 1 {
					t.Errorf("unexpected number of steps %d", len(ft.Steps()))
				}
			},
		},
		{
			name: "multiple assessments",
			setup: func(t *testing.T) types.Feature {
				return New("test").Assess("some test", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Assess("some tets 2", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				assessments := GetStepsByLevel(ft.Steps(), types.LevelAssess)
				if len(assessments) != 2 {
					t.Errorf("unexpected number of setup functions: %d", len(assessments))
				}
				if len(ft.Steps()) != 2 {
					t.Errorf("unexpected number of steps %d", len(ft.Steps()))
				}
			},
		},
		{
			name: "all steps",
			setup: func(t *testing.T) types.Feature {
				return New("test").Setup(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Assess("some tets 2", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Assess("some tets 3", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Teardown(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					// test
					return ctx
				}).Feature()
			},
			eval: func(t *testing.T, f types.Feature) {
				ft := f.(*defaultFeature) // nolint
				if len(ft.Steps()) != 4 {
					t.Errorf("unexpected number of steps %d", len(ft.Steps()))
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.eval(t, test.setup(t))
		})
	}
}
