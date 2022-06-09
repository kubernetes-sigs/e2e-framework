/*
Copyright 2022 The Kubernetes Authors.

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

package fail_fast

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestExample(t *testing.T) {
	failFeature := features.New("fail-feature").
		Assess("1==2", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			if 1 != 2 {
				t.Log("1 != 2")
				t.FailNow() // mark test case as failed here, don't continue execution
			} else {
				t.Log("1 == 2")
			}
			return ctx
		}).
		Assess("print", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("THIS LINE SHOULDN'T BE PRINTED")
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			t.Log("This teardown should not be invoked")
			return ctx
		}).
 		Feature()

	nextFeature := features.New("next-feature").
		Assess("print", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("THIS LINE ALSO SHOULDN'T BE PRINTED")
			return ctx
		}).
		Feature()

	testenv.Test(t, failFeature, nextFeature)
}

// even if the previous testcase fails, execute this testcase
func TestNext(t *testing.T) {
	nextFeature := features.New("next-test-feature").
		Assess("print", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			t.Log("THIS LINE SHOULD BE PRINTED")
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			t.Log("This teardown should be invoked")
			return ctx
		}).
		Feature()

	testenv.Test(t, nextFeature)
}
