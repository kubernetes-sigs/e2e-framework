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
	"log"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/conf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

type (
	Environment = types.Environment
	Func        = types.EnvFunc

	actionRole uint8
)

type testEnv struct {
	cfg     *conf.Config
	actions []action
}

// New creates a test environment with no config attached.
func New() types.Environment {
	return &testEnv{cfg: conf.New()}
}

// NewWithConfig convenience constructor function that takes
// a default *conf.Config
func NewWithConfig(cfg *conf.Config) types.Environment {
	return &testEnv{cfg: cfg}
}

func newTestEnv(cfg *conf.Config) *testEnv {
	return &testEnv{cfg: cfg}
}

func (e *testEnv) Config() *conf.Config {
	return e.cfg
}

func (e *testEnv) Setup(funcs ...Func) types.Environment {
	if len(funcs) == 0 {
		return e
	}
	e.actions = append(e.actions, action{role: roleSetup, funcs: funcs})
	return e
}

func (e *testEnv) BeforeTest(funcs ...Func) types.Environment {
	if len(funcs) == 0 {
		return e
	}
	e.actions = append(e.actions, action{role: roleBefore, funcs: funcs})
	return e
}

// Test executes a feature test from within a TestXXX function.
//
// Feature setups and teardowns are executed at the same *testing.T
// contextual level as the test that invoked this method. Assessments
// are executed as a subtests of the feature.  This approach allows
// features/assessments to be filtered using go test -run flag.
//
// Feature tests will have access to and able to update the context
// passed to it. The updated context is returned by this call.
//
// BeforeTest and AfterTest operations are executed before and after
// the feature is tested respectively.
func (e *testEnv) Test(ctx context.Context, t *testing.T, feature types.Feature) context.Context {
	befores := e.GetBeforeActions()
	var err error
	for _, action := range befores {
		if ctx, err = action.run(ctx); err != nil {
			t.Fatalf("BeforeTest failure: %s: %v", feature.Name(), err)
		}
	}

	ctx = e.execFeature(ctx, t, feature)

	afters := e.GetAfterActions()
	for _, action := range afters {
		if ctx, err = action.run(ctx); err != nil {
			t.Fatalf("AfterTest failure: %s: %v", feature.Name(), err)
		}
	}

	return ctx
}

func (e *testEnv) AfterTest(funcs ...Func) types.Environment {
	if len(funcs) == 0 {
		return e
	}
	e.actions = append(e.actions, action{role: roleAfter, funcs: funcs})
	return e
}

func (e *testEnv) Finish(funcs ...Func) types.Environment {
	if len(funcs) == 0 {
		return e
	}

	e.actions = append(e.actions, action{role: roleFinish, funcs: funcs})
	return e
}

// Run is to launch the test suite from a TestMain function.
// It will run m.Run() and exercise all test functions in the
// package.  This method will all Env.Setup operations prior to
// starting the tests and run all Env.Finish operations after
// before completing the suite.
//
func (e *testEnv) Run(ctx context.Context, m *testing.M) int {
	setups := e.GetSetupActions()
	// fail fast on setup, upon err exit
	var err error
	for _, setup := range setups {
		// context passed down to each setup
		if ctx, err = setup.run(ctx); err != nil {
			log.Fatal(err)
		}
	}

	exitCode := m.Run()

	finishes := e.GetFinishActions()
	// attempt to gracefully clean up.
	// Upon error, log and continue.
	for _, fin := range finishes {
		// context passed down to each finish step
		if ctx, err = fin.run(ctx); err != nil {
			log.Println(err)
		}
	}

	return exitCode
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

func (e *testEnv) execFeature(ctx context.Context, t *testing.T, f types.Feature) context.Context {
	featName := f.Name()

	// feature-level subtest
	t.Run(featName, func(t *testing.T) {
		if e.cfg.FeatureRegex() != nil && !e.cfg.FeatureRegex().MatchString(featName) {
			t.Skipf(`Skipping feature "%s": name not matched`, featName)
		}

		// setups run at feature-level
		setups := features.GetStepsByLevel(f.Steps(), types.LevelSetup)
		for _, setup := range setups {
			ctx = setup.Func()(ctx, t)
		}

		// assessments run as feature/assessment sub level
		assessments := features.GetStepsByLevel(f.Steps(), types.LevelAssess)

		for _, assess := range assessments {
			t.Run(assess.Name(), func(t *testing.T) {
				if e.cfg.AssessmentRegex() != nil && !e.cfg.AssessmentRegex().MatchString(assess.Name()) {
					t.Skipf(`Skipping assessment "%s": name not matched`, assess.Name())
				}
				ctx = assess.Func()(ctx, t)
			})
		}

		// teardowns run at feature-level
		teardowns := features.GetStepsByLevel(f.Steps(), types.LevelTeardown)
		for _, teardown := range teardowns {
			ctx = teardown.Func()(ctx, t)
		}
	})

	return ctx
}
