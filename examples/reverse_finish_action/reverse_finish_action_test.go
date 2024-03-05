/*
Copyright 2024 The Kubernetes Authors.

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

package reversefinishaction

import (
	"context"
	"log"
	"os"
	"testing"

	klog "k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	testEnv env.Environment

	setupActionTracker  []int
	finishActionTracker []int
)

func TestMain(m *testing.M) {
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		log.Fatalf("failed to build envconf from flags: %s", err)
	}
	testEnv = env.NewWithConfig(cfg)
	testEnv.Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		setupActionTracker = append(setupActionTracker, 1)
		return ctx, nil
	})
	testEnv.Finish(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		finishActionTracker = append(finishActionTracker, 1)
		return ctx, nil
	})

	testEnv.Setup(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		setupActionTracker = append(setupActionTracker, 2)
		return ctx, nil
	})
	testEnv.Finish(func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		finishActionTracker = append(finishActionTracker, 2)
		return ctx, nil
	})

	testResult := testEnv.Run(m)
	klog.InfoS("Action Trigger Ordering", "setupAction", setupActionTracker, "finishAction", finishActionTracker)
	os.Exit(testResult)
}

func TestReverseFinishAction(t *testing.T) {
	// Just a take logger here to ensure we have some test running
	t.Log("Running ReverseFinishAction")
}
