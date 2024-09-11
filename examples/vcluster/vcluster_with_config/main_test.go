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

package vcluster

import (
	"context"
	"log"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support"
	"sigs.k8s.io/e2e-framework/support/kind"
	"sigs.k8s.io/e2e-framework/third_party/vcluster"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	opts := []support.ClusterOpts{}
	// vcluster requires a "host" cluster to install into, so we should resolve one
	if os.Getenv("REAL_CLUSTER") == "true" {
		cfg := conf.ResolveKubeConfigFile()
		opts = append(opts, vcluster.WithHostKubeConfig(cfg))
	} else {
		// create a kind cluster to use as the vcluster "host"
		cfg, err := kind.NewProvider().WithName("kind-vc-host").Create(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts, vcluster.WithHostKubeConfig(cfg))
	}

	testenv, _ = env.NewFromFlags()
	vclusterName := envconf.RandomName("vcluster-with-config", 16)
	namespace := envconf.RandomName("vcluster-ns", 16)
	opts = append(opts, vcluster.WithNamespace(namespace))
	testenv.Setup(
		envfuncs.CreateNamespace(namespace),
		envfuncs.CreateClusterWithConfig(vcluster.NewProvider(), vclusterName, "values.yaml", opts...),
	)

	testenv.Finish(
		envfuncs.DestroyCluster(vclusterName),
		envfuncs.DeleteNamespace(namespace),
	)

	if os.Getenv("REAL_CLUSTER") != "true" {
		// cleanup the vcluster "host"-kind-cluster
		testenv.Finish(
			func(ctx context.Context, c *envconf.Config) (context.Context, error) {
				if err := kind.NewProvider().WithName("kind-vc-host").Destroy(ctx); err != nil {
					return ctx, err
				}
				return ctx, nil
			},
		)
	}

	os.Exit(testenv.Run(m))
}
