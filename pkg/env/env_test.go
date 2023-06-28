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
	"sync"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/internal/types"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestEnv_New(t *testing.T) {
	e := newTestEnv()

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
		setup    func(*testing.T, context.Context) []string
		expected []string
	}{
		{
			name: "feature only",
			ctx:  context.TODO(),
			expected: []string{
				"test-feat",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(ctx, t, f.Feature())
				return
			},
		},
		{
			name: "filtered feature",
			ctx:  context.TODO(),
			expected: []string{
				"test-feat",
			},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := NewWithConfig(envconf.New().WithFeatureRegex("test-feat"))
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(ctx, t, f.Feature())

				env2 := NewWithConfig(envconf.New().WithFeatureRegex("skip-me"))
				f2 := features.New("test-feat-2").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat-2")
					return ctx
				})
				env2.Test(ctx, t, f2.Feature())

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
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				env.BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "before-each-test")
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(ctx, t, f.Feature())
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
			setup: func(t *testing.T, ctx context.Context) (val []string) {
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
				env.Test(ctx, t, f.Feature())
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
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				env.AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
					val = append(val, "after-each-test")
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
					val = append(val, "test-feat")
					return ctx
				})
				env.Test(ctx, t, f.Feature())
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
			setup: func(t *testing.T, ctx context.Context) (val []string) {
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
				env.Test(ctx, t, f.Feature())
				return
			},
		},
		{
			name: "context value propagation with before, during, and after test",
			ctx:  context.TODO(),
			expected: []string{
				"non-mutated",
			},
			setup: func(t *testing.T, ctx context.Context) []string {
				ctx = context.WithValue(ctx, &ctxTestKeyString{}, []string{"non-mutated"})
				env := newTestEnv()

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

				env.Test(ctx, t, f.Feature())
				return ctx.Value(&ctxTestKeyString{}).([]string)
			},
		},
		{
			name:     "no features specified",
			ctx:      context.TODO(),
			expected: []string{},
			setup: func(t *testing.T, ctx context.Context) (val []string) {
				env := newTestEnv()
				env.Test(ctx, t)
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
			setup: func(t *testing.T, ctx context.Context) (val []string) {
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

				env.Test(ctx, t, f1.Feature(), f2.Feature())
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
			setup: func(t *testing.T, ctx context.Context) (val []string) {
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
				env.Test(ctx, t, f1.Feature(), f2.Feature())
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
			setup: func(t *testing.T, ctx context.Context) []string {
				env := newTestEnv()
				val := []string{}
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
				env.Test(ctx, t, f1.Feature(), f2.Feature())
				return val
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
			setup: func(t *testing.T, ctx context.Context) []string {
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
				env.Test(ctx, t, f1.Feature(), f2.Feature())
				return val
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.setup(t, test.ctx)
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

// This test shows there is no context propagation from
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

	envForTesting.Test(mainCtx, t, f.Feature())

	finalVal, ok := mainCtx.Value(&ctxTestKeyString{}).([]string)
	if !ok {
		t.Fatal("wrong type")
	}

	expected := []string{"initial-val"}
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
	ctx := context.TODO()
	env := NewParallel()
	mutex := sync.Mutex{}
	beforeEachCallCount := 0
	afterEachCallCount := 0
	beforeFeatureCount := 0
	afterFeatureCount := 0
	env.BeforeEachTest(func(ctx context.Context, config *envconf.Config, t *testing.T) (context.Context, error) {
		mutex.Lock()
		defer mutex.Unlock()
		beforeEachCallCount++
		return ctx, nil
	})

	env.AfterEachTest(func(ctx context.Context, config *envconf.Config, t *testing.T) (context.Context, error) {
		mutex.Lock()
		defer mutex.Unlock()
		afterEachCallCount++
		return ctx, nil
	})

	env.BeforeEachFeature(func(ctx context.Context, config *envconf.Config, _ *testing.T, feature types.Feature) (context.Context, error) {
		t.Logf("Running before each feature for feature %s", feature.Name())
		mutex.Lock()
		defer mutex.Unlock()
		beforeFeatureCount++
		return ctx, nil
	})

	env.AfterEachFeature(func(ctx context.Context, config *envconf.Config, _ *testing.T, feature types.Feature) (context.Context, error) {
		t.Logf("Running after each feature for feature %s", feature.Name())
		mutex.Lock()
		defer mutex.Unlock()
		afterFeatureCount++
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

	env.TestInParallel(ctx, t, f1.Feature(), f2.Feature())
	if beforeEachCallCount > 1 {
		t.Fatal("BeforeEachTest handler should be invoked only once")
	}
}

func TestParallelOne(t *testing.T) {
	ctx := context.TODO()
	t.Parallel()
	f := features.New("parallel one").
		Assess("log a message", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("Running in parallel")
			return ctx
		})
	envForTesting.Test(ctx, t, f.Feature())
}

func TestParallelTwo(t *testing.T) {
	ctx := context.TODO()
	t.Parallel()
	f := features.New("parallel two").
		Assess("log a message", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("Running in parallel")
			return ctx
		})

	envForTesting.Test(ctx, t, f.Feature())
}
