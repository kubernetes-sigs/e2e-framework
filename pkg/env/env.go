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

	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

type Environment = types.Environment
type Func = types.EnvFunc

type actionRole uint8
const(
	roleSetup = iota
	roleBefore
	roleAfter
	roleFinish
)

// action a group env functions
type action struct {
	role actionRole
	ctx context.Context
	funcs []types.EnvFunc
}

type testEnv struct {
	cfg types.Config
	actions []action
}

func New(cfg types.Config) types.Environment {
	return newTestEnv(cfg)
}

func newTestEnv(cfg types.Config) *testEnv {
	return &testEnv{cfg:cfg}
}

func (e *testEnv) Config() types.Config {
	return e.cfg
}

func (e *testEnv) Setup(ctx context.Context, funcs ...Func) types.Environment{
	if len(funcs) == 0{
		return e
	}
	e.actions = append(e.actions, action{ctx:ctx, role:roleSetup, funcs:funcs})
	return e
}

func (e *testEnv) BeforeTest(ctx context.Context, funcs ...Func) types.Environment{
	if len(funcs) == 0{
		return e
	}
	e.actions = append(e.actions, action{ctx:ctx, role:roleBefore, funcs:funcs})
	return e
}

func (e *testEnv) Test(ctx context.Context, t *testing.T, feature features.Feature) {

}

func (e *testEnv) AfterTest(ctx context.Context, funcs ...Func) types.Environment{
	if len(funcs) == 0{
		return e
	}
	e.actions = append(e.actions, action{ctx:ctx, role:roleAfter, funcs:funcs})
	return e
}

func (e *testEnv) Finish(ctx context.Context, funcs ...Func) types.Environment{
	if len(funcs) == 0{
		return e
	}
	e.actions = append(e.actions, action{ctx:ctx, role:roleFinish, funcs:funcs})
	return e
}

func (e *testEnv) Run(m *testing.M) int {
	return m.Run()
}

func (e *testEnv) getActionsByRole(r actionRole) []action {
	if e.actions == nil {
		return nil
	}
	var result []action
	for _, a := range e.actions {
		if a.role == r {
			result = append(result, a)
		}
	}
	return result
}

func (e *testEnv) GetSetupActions() []action {
	return e.getActionsByRole(roleSetup)
}

func (e *testEnv) GetBeforeActions() []action {
	return e.getActionsByRole(roleBefore)
}

func (e *testEnv) GetAfterActions() []action {
	return e.getActionsByRole(roleAfter)
}

func (e *testEnv) GetFinishActions() []action {
	return e.getActionsByRole(roleFinish)
}