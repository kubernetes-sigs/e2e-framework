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

type testEnv struct {
	cfg types.Config
}

func New(cfg types.Config) types.Environment {
	return &testEnv{cfg:cfg}
}

func (e *testEnv) Config() types.Config {
	return e.cfg
}

func (e *testEnv) Setup(ctx context.Context, f ...Func) error {
	panic("implement me")
}

func (e *testEnv) BeforeTest(ctx context.Context, t2 *testing.T, f ...Func) {
	panic("implement me")
}

func (e *testEnv) Test(ctx context.Context, t2 *testing.T, feature features.Feature) {
	panic("implement me")
}

func (e *testEnv) AfterTest(ctx context.Context, t2 *testing.T, f ...Func) {
	panic("implement me")
}

func (e *testEnv) Finish(ctx context.Context, f ...Func) error {
	panic("implement me")
}

func (e *testEnv) Run(m *testing.M) int {
	panic("implement me")
}
