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
	"fmt"

	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

type State = types.State
type Feature = types.Feature
type Step = types.Step
type Func = types.StepFunc
type Level = types.Level


// FeatureBuilder represents is a type to define a
// teatable feature
type FeatureBuilder struct {
	feat *featureTest
}

func New(name string) *FeatureBuilder {
	return &FeatureBuilder{feat: newTestFeature(name)}
}

func (b *FeatureBuilder) State(s State) *FeatureBuilder {
	b.feat.state = s
	return b
}

// Setup defines a single step applied prior to feature test.
// Subsequent calls to Setup will overwrite previous one.
func (b *FeatureBuilder) Setup(fn Func) *FeatureBuilder {
	b.feat.setup =  newStep(fmt.Sprintf("%s-setup", b.feat.name), types.LevelSetup, fn)
	return b
}

// Teardown defines a single step applied prior test completion.
// Subsequent calls to Teardown will overwrite previous one.
func (b *FeatureBuilder) Teardown(fn Func) *FeatureBuilder {
	b.feat.teardown =  newStep(fmt.Sprintf("%s-teardown", b.feat.name), types.LevelTeardown, fn)
	return b
}


func (b *FeatureBuilder) Assess(desc string, fn Func) *FeatureBuilder {
	i := len(b.feat.assessments)
	stepName := fmt.Sprintf(desc, b.feat.name, i)
	b.feat.assessments = append(b.feat.assessments, newStep(stepName, types.LevelRequired, fn))
	return b
}

func (b *FeatureBuilder) Feature() types.Feature {
	return b.feat
}

type featureTest struct {
	name string
	state State
	setup types.Step
	teardown types.Step
	assessments []types.Step
}

func newTestFeature(name string) *featureTest {
	return &featureTest{name: name}
}

func (f *featureTest) Name() string {
	return f.name
}

func (f *featureTest) State() State {
	return f.state
}

func (f *featureTest) Steps() []types.Step {
	return append(append([]types.Step{f.setup}, f.assessments...), f.teardown)
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