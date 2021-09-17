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

package custom_env_funcs

import (
	"context"
	"os"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	testenv.Setup(
		// Step: creates kind cluster, propagate kind cluster object
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			name := envconf.RandomName("my-cluster", 16)
			cluster := kind.NewCluster(name)
			kubeconfig, err := cluster.Create()
			if err != nil {
				return ctx, err
			}
			// stall a bit to allow most pods to come up
			time.Sleep(time.Second * 10)

			// update environment with kubecofig file
			cfg.WithKubeconfigFile(kubeconfig)
			// propagate cluster value
			return context.WithValue(ctx, 1, cluster), nil
		},
	).Finish(
		// Teardown func: delete kind cluster
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			cluster := ctx.Value(1).(*kind.Cluster) // nil should be tested
			if err := cluster.Destroy(); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
	)

	os.Exit(testenv.Run(m))
}
