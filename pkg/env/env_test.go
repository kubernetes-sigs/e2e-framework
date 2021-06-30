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
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/conf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestEnv_New(t *testing.T) {
	e := newTestEnv(conf.New())
	if e.Config() == nil {
		t.Error("missing config")
	}

	if len(e.actions) != 0 {
		t.Error("unexpected actions found")
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
				return newTestEnv(conf.New())
			},
			roles: map[actionRole]int{roleSetup: 0, roleBefore: 0, roleAfter: 0, roleFinish: 0},
		},
		{
			name: "setup actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv(conf.New())
				env.Setup(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				}).Setup(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 2, roleBefore: 0, roleAfter: 0, roleFinish: 0},
		},
		{
			name: "before actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv(conf.New())
				env.BeforeTest(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBefore: 1, roleAfter: 0, roleFinish: 0},
		},
		{
			name: "after actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv(conf.New())
				env.AfterTest(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBefore: 0, roleAfter: 1, roleFinish: 0},
		},
		{
			name: "finish actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv(conf.New())
				env.Finish(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 0, roleBefore: 0, roleAfter: 0, roleFinish: 1},
		},
		{
			name: "all actions",
			setup: func(t *testing.T) *testEnv {
				env := newTestEnv(conf.New())
				env.Setup(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				}).BeforeTest(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				}).AfterTest(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				}).Finish(func(ctx context.Context) (context.Context, error) {
					return ctx, nil
				})
				return env
			},
			roles: map[actionRole]int{roleSetup: 1, roleBefore: 1, roleAfter: 1, roleFinish: 1},
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
		setup    func(*testing.T, context.Context) int
		expected int
	}{
		{
			name:     "feature only",
			ctx:      context.TODO(),
			expected: 42,
			setup: func(t *testing.T, ctx context.Context) (val int) {
				env := newTestEnv(conf.New())
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T) context.Context {
					val = 42
					return ctx
				})
				env.Test(ctx, t, f.Feature())
				return
			},
		},
		{
			name:     "filtered feature",
			ctx:      context.TODO(),
			expected: 42,
			setup: func(t *testing.T, ctx context.Context) (val int) {
				env := newTestEnv(conf.New().WithFeatureRegex("test-feat"))
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T) context.Context {
					val = 42
					return ctx
				})
				env.Test(ctx, t, f.Feature())

				env2 := newTestEnv(conf.New().WithFeatureRegex("skip-me"))
				f2 := features.New("test-feat-2").Assess("assess", func(ctx context.Context, t *testing.T) context.Context {
					val = 42 + 1
					return ctx
				})
				env2.Test(ctx, t, f2.Feature())

				return
			},
		},
		{
			name:     "with before-test",
			ctx:      context.TODO(),
			expected: 86,
			setup: func(t *testing.T, ctx context.Context) (val int) {
				env := newTestEnv(conf.New())
				env.BeforeTest(func(ctx context.Context) (context.Context, error) {
					val = 44
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T) context.Context {
					val += 42
					return ctx
				})
				env.Test(ctx, t, f.Feature())
				return
			},
		},
		{
			name:     "with after-test",
			ctx:      context.TODO(),
			expected: 66,
			setup: func(t *testing.T, ctx context.Context) (val int) {
				env := newTestEnv(conf.New())
				env.AfterTest(func(ctx context.Context) (context.Context, error) {
					val -= 20
					return ctx, nil
				}).BeforeTest(func(ctx context.Context) (context.Context, error) {
					val = 44
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T) context.Context {
					val += 42
					return ctx
				})
				env.Test(ctx, t, f.Feature())
				return
			},
		},
		{
			name:     "with before-after-test",
			ctx:      context.TODO(),
			expected: 44,
			setup: func(t *testing.T, ctx context.Context) (val int) {
				env := newTestEnv(conf.New())
				env.AfterTest(func(ctx context.Context) (context.Context, error) {
					val = 44
					return ctx, nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T) context.Context {
					val = 42 + val
					return ctx
				})
				env.Test(ctx, t, f.Feature())
				return
			},
		},
		{
			name:     "filter assessment",
			ctx:      context.TODO(),
			expected: 45,
			setup: func(t *testing.T, ctx context.Context) (val int) {
				val = 42
				env := newTestEnv(conf.New().WithAssessmentRegex("add-*"))
				f := features.New("test-feat").
					Assess("add-one", func(ctx context.Context, t *testing.T) context.Context {
						val++
						return ctx
					}).
					Assess("add-two", func(ctx context.Context, t *testing.T) context.Context {
						val += 2
						return ctx
					}).
					Assess("take-one", func(ctx context.Context, t *testing.T) context.Context {
						val--
						return ctx
					})
				env.Test(ctx, t, f.Feature())
				return
			},
		},
		{
			name:     "with before-test, after-test using context",
			ctx:      context.TODO(),
			expected: 46,
			setup: func(t *testing.T, ctx context.Context) int {
				env := newTestEnv(conf.New())
				// TODO: add test case using env.Setup once context propagation is supported from Setup
				env.BeforeTest(func(ctx context.Context) (context.Context, error) {
					return context.WithValue(ctx, &ctxKey{}, 44), nil
				})
				env.AfterTest(func(ctx context.Context) (context.Context, error) {
					val, ok := ctx.Value(&ctxKey{}).(int)
					if !ok {
						t.Fatal("context value was not int")
					}
					val++

					return context.WithValue(ctx, &ctxKey{}, val), nil
				})
				f := features.New("test-feat").Assess("assess", func(ctx context.Context, t *testing.T) context.Context {
					val, ok := ctx.Value(&ctxKey{}).(int)
					if !ok {
						t.Fatal("context value was not int")
					}
					val++

					return context.WithValue(ctx, &ctxKey{}, val)
				})

				ctx = env.Test(ctx, t, f.Feature())
				return ctx.Value(&ctxKey{}).(int)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.setup(t, test.ctx)
			if result != test.expected {
				t.Error("unexpected result: ", result)
			}
		})
	}
}
