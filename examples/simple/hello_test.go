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

package simple

import (
	"context"
	"fmt"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func Hello(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

// TestHello shows an example of a test environment
// that uses a simple setup to assess a feature (test)
// in a test function directly (outside of test suite TestMain)
func TestHello(t *testing.T) {
	e := env.NewWithConfig(envconf.New())
	feat := features.New("Hello Feature").
		WithLabel("type", "simple").
		Assess("test message", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello("foo")
			if result != "Hello foo" {
				t.Error("unexpected message")
			}
			return ctx
		})

	e.Test(t, feat.Feature())
}

// The following shows an example of a simple
// test function that uses feature with a setup
// step.
func TestHello_WithSetup(t *testing.T) {
	e := env.NewWithConfig(envconf.New())
	var name string
	feat := features.New("Hello Feature").
		WithLabel("type", "simple").
		Setup(func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			name = "foobar"
			return ctx
		}).
		Assess("test message", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
			result := Hello(name)
			if result != "Hello foobar" {
				t.Error("unexpected message")
			}
			return ctx
		}).Feature()

	e.Test(t, feat)
}
