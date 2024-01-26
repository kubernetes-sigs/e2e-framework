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

package dry_run

import (
	"context"
	"testing"

	klog "k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestDryRunOne(t *testing.T) {
	f1 := features.New("F1").
		Assess("Assessment One", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Perform Some assessment
			return ctx
		}).Feature()

	f2 := features.New("F2").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			klog.Info("Do not run this when in dry-run mode")
			return ctx
		}).
		Assess("Assessment One", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Perform Some assessment
			return ctx
		}).
		Assess("Assessment Two", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Perform Some assessment
			return ctx
		}).Feature()

	_ = testEnv.TestInParallel(t, f1, f2)
}

func TestDryRunTwo(t *testing.T) {
	f1 := features.New("F1").
		Assess("Assessment One", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			// Perform Some assessment
			return ctx
		}).Feature()

	testEnv.Test(t, f1)
}
