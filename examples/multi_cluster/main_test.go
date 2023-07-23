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

package multi_cluster

import (
	"context"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	testEnv      env.Environment
	clusterNames []string
)

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)

	clusterNames = []string{
		envconf.RandomName("cluster-one", 16),
		envconf.RandomName("cluster-two", 16),
	}

	testEnv.Setup(
		func(ctx context.Context, config *envconf.Config) (context.Context, error) {
			var err error
			for _, cluster := range clusterNames {
				ctx, err = envfuncs.CreateCluster(kind.NewProvider(), cluster)(ctx, config)
				if err != nil {
					return ctx, err
				}
			}
			return ctx, nil
		},
	).Finish(
		func(ctx context.Context, config *envconf.Config) (context.Context, error) {
			var err error
			for _, cluster := range clusterNames {
				ctx, err = envfuncs.DestroyCluster(cluster)(ctx, config)
				if err != nil {
					return ctx, err
				}
			}
			return ctx, nil
		},
	)

	os.Exit(testEnv.Run(m))
}
