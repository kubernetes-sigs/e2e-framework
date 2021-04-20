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
	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

type Labels = types.Labels
type Feature = types.Feature
type Step = types.Step
type Func = types.StepFunc
type Level = types.Level



type featureTest struct {
	name string
	labels types.Labels
	steps []types.Step
}

func newTestFeature(name string) *featureTest {
	return &featureTest{name: name, labels: make(types.Labels)}
}

func (f *featureTest) Name() string {
	return f.name
}

func (f *featureTest) Labels() types.Labels {
	return f.labels
}

func (f *featureTest) Steps() []types.Step {
	return f.steps
}

func (f *featureTest) GetSetups() []types.Step {
	return f.getStepsByLevel(types.LevelSetup)
}

func (f *featureTest) GetAssessments() []types.Step {
	return f.getStepsByLevel(types.LevelAssess)
}

func (f *featureTest) GetTeardowns() []types.Step {
	return f.getStepsByLevel(types.LevelTeardown)
}

func (f *featureTest) getStepsByLevel(l types.Level) []Step {
	if f.steps == nil {
		return nil
	}
	var result []Step
	for _, s := range f.Steps() {
		if s.Level() == l {
			result = append(result, s)
		}
	}
	return result
}

type testStep struct {
	name string
	level Level
	fn Func
}

func newStep(name string, level Level, fn Func) *testStep {
	return &testStep{
		name: name,
		level: level,
		fn: fn,
	}
}

func (s *testStep) Name() string {
	return s.name
}

func (s *testStep) Level() Level {
	return s.level
}

func (s *testStep) Func() Func {
	return s.fn
}