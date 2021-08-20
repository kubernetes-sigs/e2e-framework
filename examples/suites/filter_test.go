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

package suites

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// The following test sets ups three assessments
// `add-bazz`, `repeat-msg`, and `add-bat`.
// The test uses a regular expression configuration
// to only execute assessment with regex form `add-*`
// which would cause assessment `repeat-msg` to be skipped.
func TestHello_WithFilters(t *testing.T) {
	env := env.NewWithConfig(envconf.New().WithAssessmentRegex("add-*"))
	feat := features.New("Hello Feature").
		Assess("add-bazz", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello("bazz")
			if result != "Hello bazz" {
				t.Error("unexpected message")
			}
			return ctx
		}).
		Assess("repeat-msg", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello(Hello("bat"))
			if result != "Hello Hello bat" {
				t.Error("unexpected message")
			}
			return ctx
		}).
		Assess("add-bat", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello("bat")
			if result != "Hello bat" {
				t.Error("unexpected message")
			}
			return ctx
		}).
		Feature()

	env.Test(t, feat)
}
