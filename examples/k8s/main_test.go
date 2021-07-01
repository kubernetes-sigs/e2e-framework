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

package k8s

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
	testenv = env.New()
	testenv.Setup(
		// env func: creates kind cluster, propagate kubeconfig file name
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			cluster := envconf.RandomName("my-cluster", 16)
			kubecfg, err := createKindCluster(cluster)
			if err != nil {
				return ctx, err
			}
			// stall a bit to allow most pods to come up
			time.Sleep(time.Second * 10)

			// propagate cluster name and kubeconfig file name
			return context.WithValue(context.WithValue(ctx, 1, kubecfg), 2, cluster), nil
		},
		// env func: creates a klient.Client for the envconfig.Config
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			kubecfg := ctx.Value(1).(string)
			// create a klient.Client and set it for the env config
			client, err := klient.NewWithKubeConfigFile(kubecfg)
			if err != nil {
				return ctx, fmt.Errorf("create klient.Client: %w", err)
			}
			cfg.WithClient(client) // set client in envconfig
			return ctx, nil
		},
	).Finish(
		// Teardown func: delete kind cluster and delete kubecfg file
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			kubecfg := ctx.Value(1).(string)
			cluster := ctx.Value(2).(string)
			if err := deleteKindCluster(cluster, kubecfg); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
	)

	os.Exit(testenv.Run(m))
}
