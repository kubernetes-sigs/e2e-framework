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
	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

func TestNew(t *testing.T) {
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
		eval  func(*testing.T, *testEnv)
	}{
		{
			name: "empty actions",
			setup: func(t *testing.T) *testEnv{
				return newTestEnv(conf.New())
			},
			eval: func(t *testing.T, e *testEnv){
				if len(e.GetSetupActions()) != 0 {
					t.Error("setup actions should be 0")
				}
				if len(e.GetBeforeActions()) != 0 {
					t.Error("before actions should be 0")
				}
				if len(e.GetAfterActions()) != 0 {
					t.Error("after actions should be 0")
				}
				if len(e.GetFinishActions()) != 0 {
					t.Error("finish actions should be 0")
				}
			},
		},
		{
			name: "setup actions",
			setup: func(t *testing.T) *testEnv{
				env := newTestEnv(conf.New())
				env.Setup(context.Background(), func(ctx context.Context, conf types.Config) error {
					return nil
				}).Setup(context.Background(), func (ctx context.Context, conf types.Config) error {
					return nil
				})
				return env
			},
			eval: func(t *testing.T, e *testEnv){
				if len(e.GetSetupActions()) != 2 {
					t.Error("setup actions should be 2")
				}
			},
		},
		{
			name: "before actions",
			setup: func(t *testing.T) *testEnv{
				env := newTestEnv(conf.New())
				env.BeforeTest(context.Background(), func(ctx context.Context, conf types.Config) error {
					return nil
				})
				return env
			},
			eval: func(t *testing.T, e *testEnv){
				if len(e.GetBeforeActions()) != 1 {
					t.Error("before actions should be 1")
				}
			},
		},
		{
			name: "after actions",
			setup: func(t *testing.T) *testEnv{
				env := newTestEnv(conf.New())
				env.AfterTest(context.Background(), func(ctx context.Context, conf types.Config) error {
					return nil
				})
				return env
			},
			eval: func(t *testing.T, e *testEnv){
				if len(e.GetAfterActions()) != 1 {
					t.Error("after actions should be 1")
				}
			},
		},
		{
			name: "finish actions",
			setup: func(t *testing.T) *testEnv{
				env := newTestEnv(conf.New())
				env.Finish(context.Background(), func(ctx context.Context, conf types.Config) error {
					return nil
				})
				return env
			},
			eval: func(t *testing.T, e *testEnv){
				if len(e.GetFinishActions()) != 1 {
					t.Error("finish actions should be 1")
				}
			},
		},
		{
			name: "all actions",
			setup: func(t *testing.T) *testEnv{
				env := newTestEnv(conf.New())
				env.Setup(context.Background(), func(ctx context.Context, conf types.Config) error {
					return nil
				}).BeforeTest(context.Background(), func(ctx context.Context, config types.Config) error {
					return nil
				}).AfterTest(context.Background(), func(ctx context.Context, config types.Config) error {
					return nil
				}).Finish(context.Background(), func(ctx context.Context, config types.Config) error {
					return nil
				})
				return env
			},
			eval: func(t *testing.T, e *testEnv){
				if len(e.GetSetupActions()) != 1 {
					t.Error("setup actions should be 1")
				}
				if len(e.GetBeforeActions()) != 1 {
					t.Error("before actions should be 1")
				}
				if len(e.GetAfterActions()) != 1 {
					t.Error("after actions should be 1")
				}
				if len(e.GetFinishActions()) != 1 {
					t.Error("finish actions should be 1")
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T){
			test.eval(t, test.setup(t))
		})
	}
}
