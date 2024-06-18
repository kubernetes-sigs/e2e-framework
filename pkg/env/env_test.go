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

package env

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/types"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestEnv_New(t *testing.T) {
	e := newTestEnv()
	if e.ctx == nil {
		t.Error("missing default context")
	}

	if len(e.actions) != 0 {
		t.Error("unexpected actions found")
	}

	if e.cfg.Namespace() != "" {
		t.Error("unexpected envconfig.Namespace value")
	}
}

func TestEnv_APIMethods(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T) *testEnv
		roles map[actionRole]int
	}{
		{
			name: "empty actions",
			setup: func(t *testing.T) *testEnv {
				return newTestEnv()
			},
			roles: map[actionRole]int{roleSetup: 0, roleBeforeTest: 0, roleAfterTest: 0, roleFinish: 0},
		},
		{
			name: "setup actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.Setup(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				}).Setup(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 2, roleBeforeTest: 0, roleAfterTest: 0, roleFinish: 0},
		},
		{
			name: "before actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBeforeTest: 1, roleAfterTest: 0, roleFinish: 0},
		},
		{
			name: "after actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBeforeTest: 0, roleAfterTest: 1, roleFinish: 0},
		},
		{
			name: "finish actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.Finish(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBeforeTest: 0, roleAfterTest: 0, roleFinish: 1},
		},
		{
			name: "all actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv()
				env.Setup(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				}).BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					return ctx, nil
				}).AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					return ctx, nil
				}).Finish(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 1, roleBeforeTest: 1, roleAfterTest: 1, roleFinish: 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			env := test.setup(t)
			for role, count := range test.roles {
				actual := len(env.getActionsByRole(role))
				if actual != count {
					t.Errorf("unexpected number of actions %d for role %d", actual, role)
				}
			}
		})
	}
}

func TestEnv_Test(t *testing.T) {
	tests := []struct {
		name     string
		ctx      context.Context
		setup    func(context.Context, *testing.T) []string
		expected []string
	}{
		{
			name: "feature only",
			ctx:  context.TODO(),
			expected: []string{
				"test-feat",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := newTestEnv()
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				_ = env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "filtered feature",
			ctx:  context.TODO(),
			expected: []string{
				"test-feat",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := NewWithConfig(envconf.New().WithFeatureRegex("test-feat"))
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				_ = env.Test(t, f.Feature())

				env2 := NewWithConfig(envconf.New().WithFeatureRegex("skip-me"))
				f2 := features.New("test-feat-2").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				_ = env2.Test(t, f2.Feature())

				return
			},
		},
		{
			name: "with before-test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test",
				"test-feat",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := newTestEnv()
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "before-each-test")
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				_ = env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "with after-test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test",
				"test-feat",
				"after-each-test",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := newTestEnv()
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "after-each-test")
					return ctx, nil
				}).BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "before-each-test")
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				_ = env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "with before-after-test",
			ctx:  context.TODO(),
			expected: []string{
				"test-feat",
				"after-each-test",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := newTestEnv()
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "after-each-test")
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				_ = env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "filter assessment",
			ctx:  context.TODO(),
			expected: []string{
				"add-1",
				"add-2",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				val = []string{}
				env := NewWithConfig(envconf.New().WithAssessmentRegex("add-*"))
				f := features.New("test-feat").
					Assess("add-one", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "add-1")
						return ctx
					}).
					Assess("add-two", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "add-2")
						return ctx
					}).
					Assess("take-one", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "take-1")
						return ctx
					})
				_ = env.Test(t, f.Feature())
				return
			},
		},
		{
			name: "context value propagation with before, during, and after test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test",
				"test-feat",
				"after-each-test",
			},
			setup: func(ctx context.Context, t *testing.T) []string {
				env, err := NewWithContext(context.WithValue(ctx, &ctxTestKeyString{}, []string{}), envconf.New())
				if err != nil {
					t.Fatal(err)
				}
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update before test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "before-each-test")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update after the test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "after-each-test")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "test-feat")

					return context.WithValue(ctx, &ctxTestKeyString{}, val)
				})

				out := env.Test(t, f.Feature())
				return out.Value(&ctxTestKeyString{}).([]string)
			},
		},
		{
			name: "context value propagation with with multiple features, before, during, and after test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test-1",
				"before-each-test-2",
				"test-feat-1",
				"test-feat-2",
				"after-each-test-1",
				"after-each-test-2",
			},
			setup: func(ctx context.Context, t *testing.T) []string {
				env, err := NewWithContext(context.WithValue(ctx, &ctxTestKeyString{}, []string{}), envconf.New())
				if err != nil {
					t.Fatal(err)
				}
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update before test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "before-each-test-1")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update before test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "before-each-test-2")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update after the test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "after-each-test-1")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update after the test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "after-each-test-2")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				f1 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "test-feat-1")

					return context.WithValue(ctx, &ctxTestKeyString{}, val)
				}).Feature()
				f2 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "test-feat-2")

					return context.WithValue(ctx, &ctxTestKeyString{}, val)
				}).Feature()

				out := env.Test(t, f1, f2)
				return out.Value(&ctxTestKeyString{}).([]string)
			},
		},
		{
			name: "context value propagation with with multiple features in parallel, before, during, and after test",
			ctx:  context.WithValue(context.TODO(), &ctxTestKeyString{}, []string{}),
			expected: []string{
				"before-each-test",
				"after-each-test",
			},
			setup: func(ctx context.Context, t *testing.T) []string {
				env := NewParallel().WithContext(context.WithValue(ctx, &ctxTestKeyString{}, []string{}))
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update before test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "before-each-test")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					// update after the test
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "after-each-test")
					return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
				})
				f1 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "test-feat-1")

					return context.WithValue(ctx, &ctxTestKeyString{}, val)
				}).Feature()
				f2 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
					if !ok {
						t.Fatal("context value was not []string")
					}
					val = append(val, "test-feat-2")

					return context.WithValue(ctx, &ctxTestKeyString{}, val)
				}).Feature()

				out := env.TestInParallel(t, f1, f2)
				return out.Value(&ctxTestKeyString{}).([]string)
			},
		},
		{
			name:     "no features specified",
			ctx:      context.TODO(),
			expected: []string{},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := newTestEnv()
				_ = env.Test(t)
				return
			},
		},
		{
			name: "multiple features",
			ctx:  context.TODO(),
			expected: []string{
				"test-feature-1",
				"test-feature-2",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := newTestEnv()
				f1 := features.New("test-feat-1").
					Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "test-feature-1")
						return ctx
					})

				f2 := features.New("test-feat-2").
					Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
						val = append(val, "test-feature-2")
						return ctx
					})

				_ = env.Test(t, f1.Feature(), f2.Feature())
				return
			},
		},
		{
			name: "multiple features with before-after-test",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-test",
				"test-feat-1",
				"test-feat-2",
				"after-each-test",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := newTestEnv()
				val = []string{}
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "before-each-test")
					return ctx, nil
				})
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "after-each-test")
					return ctx, nil
				})
				f1 := features.New("test-feat-1").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-1")
					return ctx
				})
				f2 := features.New("test-feat-2").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				_ = env.Test(t, f1.Feature(), f2.Feature())
				return
			},
		},
		{
			name: "with before-and-after features",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-feature",
				"test-feat-1",
				"after-each-feature",
				"before-each-feature",
				"test-feat-2",
				"after-each-feature",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := newTestEnv()
				env.BeforeEachFeature(func(ctx context.Context, _ *envconf.Config, _ *testing.T, info features.Feature) (context.Context, error) {
					val = append(val, "before-each-feature")
					return ctx, nil
				}).AfterEachFeature(func(ctx context.Context, _ *envconf.Config, _ *testing.T, info features.Feature) (context.Context, error) {
					val = append(val, "after-each-feature")
					return ctx, nil
				})
				f1 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-1")
					return ctx
				})
				f2 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				_ = env.Test(t, f1.Feature(), f2.Feature())
				return
			},
		},
		{
			name: "before-and-after features unable to mutate feature",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-feature",
				"test-feat-1",
				"after-each-feature",
				"before-each-feature",
				"test-feat-2",
				"after-each-feature",
			},
			setup: func(ctx context.Context, t *testing.T) []string {
				env := newTestEnv()
				val := []string{}
				env.BeforeEachFeature(func(ctx context.Context, _ *envconf.Config, _ *testing.T, info features.Feature) (context.Context, error) {
					val = append(val, "before-each-feature")
					t.Logf("%#v, len(steps)=%v step[0].Name: %v\n", info, len(info.Steps()), info.Steps()[0].Name())

					if len(info.Steps()) == 0 {
						t.Fatal("Expected more than 0 steps at start but found 0")
					}
					if info.Steps()[0].Func() != nil {
						t.Fatal("Expected step functions to only be nil but found non-nil value")
					}

					// Prior to fixing this logic, this would cause the test to fail/panic.
					// Ensure nil'ing the value out doesn't mess up flow of function.
					info.Steps()[0] = nil

					// Ensure changes aren't persisted to the afterEachFeature hook
					labelMap := info.Labels()
					labelMap["foo"] = []string{"bar"}
					return ctx, nil
				}).AfterEachFeature(func(ctx context.Context, _ *envconf.Config, _ *testing.T, info features.Feature) (context.Context, error) {
					val = append(val, "after-each-feature")
					t.Logf("%#v, len(steps)=%v\n", info, len(info.Steps()))
					if info.Labels().Contains("foo", "bar") {
						t.Errorf("Expected label from previous feature hook to not include foo:bar")
					}
					if len(info.Steps()) == 0 {
						t.Fatal("Expected more than 0 steps at start but found 0")
					}
					if info.Steps()[0].Func() != nil {
						t.Fatal("Expected step functions to only be nil but found non-nil value")
					}
					return ctx, nil
				})
				f1 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-1")
					return ctx
				})
				f2 := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				_ = env.Test(t, f1.Feature(), f2.Feature())
				return val
			},
		},
		{
			name: "with before-and-after features",
			ctx:  context.TODO(),
			expected: []string{
				"before-each-feature",
				"test-feat-1",
				"after-each-feature",
			},
			setup: func(ctx context.Context, t *testing.T) (val []string) {
				env := NewWithConfig(envconf.New().WithSkipLabels(map[string][]string{"test": {"skip"}}))

				env.BeforeEachFeature(func(ctx context.Context, _ *envconf.Config, _ *testing.T, info features.Feature) (context.Context, error) {
					val = append(val, "before-each-feature")
					return ctx, nil
				}).AfterEachFeature(func(ctx context.Context, _ *envconf.Config, _ *testing.T, info features.Feature) (context.Context, error) {
					val = append(val, "after-each-feature")
					return ctx, nil
				})
				f1 := features.New("test-feat").
					WithLabel("test", "run").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-1")
					return ctx
				})
				f2 := features.New("test-feat").
					WithLabel("test", "skip").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				_ = env.Test(t, f1.Feature(), f2.Feature())
				return
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.setup(test.ctx, t)
			if len(test.expected) != len(result) {
				t.Fatalf("Expected:\n%v but got result:\n%v", test.expected, result)
			}
			for i := range test.expected {
				if result[i] != test.expected[i] {
					t.Errorf("Expected:\n%v but got result:\n%v", test.expected, result)
					break
				}
			}
		})
	}
}

// This test shows the full context propagation from
// environment setup functions (started in main_test.go) down to
// feature step functions.
func TestEnv_Context_Propagation(t *testing.T) {
	f := features.New("test-context-propagation").
		Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
			if !ok {
				t.Fatal("context value was not int")
			}
			val = append(val, "test-context-propagation")
			return context.WithValue(ctx, &ctxTestKeyString{}, val)
		})

	out := envForTesting.Test(t, f.Feature())

	finalVal, ok := out.Value(&ctxTestKeyString{}).([]string)
	if !ok {
		t.Fatal("wrong type")
	}

	expected := []string{"setup-1", "setup-2", "before-each-test", "test-context-propagation", "after-each-test"}
	if len(finalVal) != len(expected) {
		t.Fatalf("Expected:\n%v but got result:\n%v", expected, finalVal)
	}
	for i := range finalVal {
		if finalVal[i] != expected[i] {
			t.Errorf("Expected:\n%v but got result:\n%v", expected, finalVal)
			break
		}
	}
}

func TestTestEnv_TestInParallel(t *testing.T) {
	env := NewParallel()
	beforeEachCallCount := 0
	afterEachCallCount := 0
	var beforeFeatureCount,
		afterFeatureCount atomic.Int32
	env.BeforeEachTest(func(ctx context.Context, config *envconf.Config, t *testing.T) (context.Context, error) {
		beforeEachCallCount++
		return ctx, nil
	})

	env.AfterEachTest(func(ctx context.Context, config *envconf.Config, t *testing.T) (context.Context, error) {
		afterEachCallCount++
		return ctx, nil
	})

	env.BeforeEachFeature(func(ctx context.Context, config *envconf.Config, _ *testing.T, feature types.Feature) (context.Context, error) {
		t.Logf("Running before each feature for feature %s", feature.Name())
		beforeFeatureCount.Add(1)
		return ctx, nil
	})

	env.AfterEachFeature(func(ctx context.Context, config *envconf.Config, _ *testing.T, feature types.Feature) (context.Context, error) {
		t.Logf("Running after each feature for feature %s", feature.Name())
		afterFeatureCount.Add(1)
		return ctx, nil
	})

	f1 := features.New("test-parallel-feature1").
		Assess("check addition", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			t.Log("assertion in feature1 for assessment 1")
			time.Sleep(2 * time.Second)
			return ctx
		}).
		Assess("check addition again", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			t.Log("assertion in feature1 for assessment 2")
			time.Sleep(1 * time.Second)
			return ctx
		})

	f2 := features.New("test-parallel-feature2").
		Assess("check subtraction", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			t.Log("Assertion in feature2")
			return ctx
		})

	_ = env.TestInParallel(t, f1.Feature(), f2.Feature())
	if beforeEachCallCount != 1 {
		t.Fatal("BeforeEachTest handler should be invoked exactly once")
	}
	if afterEachCallCount != 1 {
		t.Fatal("AfterEachTest handler should be invoked exactly once")
	}
	if beforeFeatureCount.Load() != 2 {
		t.Fatal("BeforeEachFeature handler should be invoked exactly twice")
	}
	if afterFeatureCount.Load() != 2 {
		t.Fatal("AfterEachFeature handler should be invoked exactly twice")
	}
}

// TestTParallelMultipleFeaturesInParallel runs multple features in parallel with a dedicated Parallel environment,
// just to check there are no race conditions with this setting
func TestTParallelMultipleFeaturesInParallel(t *testing.T) {
	env := NewParallel()
	t.Parallel()
	f1 := features.New("feature 1").
		Assess("assess", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			time.Sleep(2 * time.Second)
			return ctx
		})
	f2 := features.New("feature 2").
		Assess("assess", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			time.Sleep(3 * time.Second)
			return ctx
		})
	_ = env.TestInParallel(t, f1.Feature(), f2.Feature())
}

// env with parallel disabled to be used in the two tests below, reusing testEnv could result on a race condition due to
// the Before/AfterEachTest accessing the same array from the context at the same time, which is not thread safe
var envTForParallelTesting = New()

// TestMultipleAssess runs multiple assessments sequentially, but can run in parallel with other parallel tests in the suite
// just to check there are no race conditions, the resulting context is not defined though as it t.Parallel() is used
// a dedicated Context for each test has to be manually created and injected into the environment before running Test
func TestTParallelMultipleAssess(t *testing.T) {
	t.Parallel()
	f := features.New("assess").
		Assess("assess one", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Logf("Started at: %s", time.Now().UTC())
			time.Sleep(3 * time.Second)
			t.Logf("Terminated at: %s", time.Now().UTC())
			return ctx
		}).
		Assess("assess two", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Logf("Started at: %s", time.Now().UTC())
			time.Sleep(2 * time.Second)
			t.Logf("Terminated at: %s", time.Now().UTC())
			return ctx
		})
	_ = envTForParallelTesting.Test(t, f.Feature())
}

// TestTParallelOne, TestTParallelTwo are used to test that there is no race condition when running in parallel by using
// t.Parallel() instead of TestInParallel on an environment with parallel disabled and running in parallel with other
// tests in the suite
func TestTParallelOne(t *testing.T) {
	t.Parallel()
	f := features.New("parallel one").
		Assess("log a message", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("Running in parallel")
			return ctx
		}).Feature()

	_ = envTForParallelTesting.Test(t, f)
}

// See comment of TestTParallelOne
func TestTParallelTwo(t *testing.T) {
	t.Parallel()
	f := features.New("parallel two").
		Assess("log a message", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("Running in parallel")
			return ctx
		}).Feature()

	_ = envTForParallelTesting.Test(t, f)
}

// env with parallel enabled to be used in the two test below, reusing testEnv defined by TestMain would result on a
// race condition due to the Before/AfterEachTest accessing the same array from the context at the same time, for
// multiple parallel tests which is not thread safe
var envForTParallelInParallelTesting = NewParallel()

type ctxRunsKeyString struct{}

// TestTParallelInParallelOne, TestTParallelInParallelTwo are used to test that there is no race condition when running in
// parallel both multiple tests using t.Parallel() and multiple features using TestInParallel per test
func TestTParallelInParallelOne(t *testing.T) {
	t.Parallel()
	out := envForTParallelInParallelTesting.TestInParallel(t, getFeaturesForTest()...)
	if i, ok := out.Value(ctxRunsKeyString{}).(int); ok && i != 0 {
		t.Fatalf("Runs should be 0, the context should not be shared between features tested in parallel with tests running in parallel, got %v", i)
	}
}

func TestTParallelInParallelTwo(t *testing.T) {
	t.Parallel()
	out := envForTParallelInParallelTesting.TestInParallel(t, getFeaturesForTest()...)
	if i, ok := out.Value(ctxRunsKeyString{}).(int); ok && i != 0 {
		t.Fatalf("Runs should be 0, the context should not be shared between features tested in parallel with tests running in parallel, got %v", i)
	}
}

func getFeaturesForTest() []features.Feature {
	f1 := features.New("parallel one").
		Assess("log a message", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("Running in parallel 1 1")
			if i := ctx.Value(ctxRunsKeyString{}); i != nil {
				return context.WithValue(ctx, ctxRunsKeyString{}, i.(int)+1)
			}
			return context.WithValue(ctx, ctxRunsKeyString{}, 1)
		}).Feature()
	f2 := features.New("parallel one").
		Assess("log a message", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("Running in parallel 1 2")
			if i := ctx.Value(ctxRunsKeyString{}); i != nil {
				return context.WithValue(ctx, ctxRunsKeyString{}, i.(int)+1)
			}
			return context.WithValue(ctx, ctxRunsKeyString{}, 1)
		}).Feature()
	return []features.Feature{f1, f2}
}
