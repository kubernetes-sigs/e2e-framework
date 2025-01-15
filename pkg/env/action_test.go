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

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
)

func TestAction_Run(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		cfg        *envconf.Config
		setup      func(context.Context, *envconf.Config) (int, error)
		expected   int
		shouldFail bool
	}{
		{
			name: "single-step action",
			cfg:  &envconf.Config{},
			ctx:  context.WithValue(context.TODO(), &ctxTestKeyString{}, 1),
			setup: func(ctx context.Context, cfg *envconf.Config) (val int, err error) {
				funcs := []types.EnvFunc{
					func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
						val = 12
						return ctx, nil
					},
				}
				_, err = (&action{role: roleSetup, funcs: funcs}).run(ctx, cfg)
				return
			},
			expected: 12,
		},
		{
			name: "multi-step action",
			cfg:  &envconf.Config{},
			ctx:  context.WithValue(context.TODO(), &ctxTestKeyString{}, 1),
			setup: func(ctx context.Context, cfg *envconf.Config) (val int, err error) {
				funcs := []types.EnvFunc{
					func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
						val = 12
						return ctx, nil
					},
					func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
						val *= 2
						return ctx, nil
					},
				}
				_, err = (&action{role: roleSetup, funcs: funcs}).run(ctx, cfg)
				return
			},
			expected: 24,
		},
		{
			name: "read from context",
			cfg:  &envconf.Config{},
			ctx:  context.WithValue(context.TODO(), &ctxTestKeyString{}, 1),
			setup: func(ctx context.Context, cfg *envconf.Config) (val int, err error) {
				funcs := []types.EnvFunc{
					func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
						i := ctx.Value(&ctxTestKeyString{}).(int) + 2 // nolint: errcheck
						val = i
						return ctx, nil
					},
					func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
						val += 3
						return ctx, nil
					},
				}
				_, err = (&action{role: roleSetup, funcs: funcs}).run(ctx, cfg)
				return
			},
			expected: 6,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.setup(test.ctx, test.cfg)
			if !test.shouldFail && err != nil {
				t.Fatalf("unexpected failure: %v", err)
			}
			if result != test.expected {
				t.Error("unexpected value:", result)
			}
		})
	}
}

func TestActionRole_String(t *testing.T) {
	tests := []struct {
		name string
		r    actionRole
		want string
	}{
		{
			name: "RoleSetup",
			r:    roleSetup,
			want: "Setup",
		},
		{
			name: "RoleBeforeTest",
			r:    roleBeforeTest,
			want: "BeforeEachTest",
		},
		{
			name: "RoleBeforeFeature",
			r:    roleBeforeFeature,
			want: "BeforeEachFeature",
		},
		{
			name: "RoleAfterEachFeature",
			r:    roleAfterFeature,
			want: "AfterEachFeature",
		},
		{
			name: "RoleAfterTest",
			r:    roleAfterTest,
			want: "AfterEachTest",
		},
		{
			name: "RoleFinish",
			r:    roleFinish,
			want: "Finish",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.r.String(); got != test.want {
				t.Errorf("String() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestActionRole_String_Unknown(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Unknown ActionRole should panic")
		}
	}()

	_ = actionRole(100).String()
}
