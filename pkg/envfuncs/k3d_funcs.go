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

package envfuncs

import (
	"context"
	"fmt"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/support/k3d"
)

type k3dContextKey string

// CreateK3dCluster returns an env.Func that is used to
// create a k3d cluster that is then injected in the context
// using the name as a key.
//
// NOTE: the returned function will update its env config with the
// kubeconfig file for the config client.
func CreateK3dCluster(clusterName string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		k := k3d.NewCluster(clusterName)
		kubecfg, err := k.Create()
		if err != nil {
			return ctx, err
		}

		// update envconfig  with kubeconfig
		cfg.WithKubeconfigFile(kubecfg)

		// store entire cluster value in ctx for future access using the cluster name
		return context.WithValue(ctx, k3dContextKey(clusterName), k), nil
	}
}

// CreateK3dClusterWithConfig returns an env.Func that is used to
// create a k3d cluster that is then injected in the context
// using the name as a key.
//
// NOTE: the returned function will update its env config with the
// kubeconfig file for the config client.
func CreateK3dClusterWithConfig(clusterName, image, configFilePath string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		k := k3d.NewCluster(clusterName)
		kubecfg, err := k.CreateWithConfig(image, configFilePath)
		if err != nil {
			return ctx, err
		}

		// update envconfig  with kubeconfig
		cfg.WithKubeconfigFile(kubecfg)

		// store entire cluster value in ctx for future access using the cluster name
		return context.WithValue(ctx, k3dContextKey(clusterName), k), nil
	}
}

// DestroyK3dCluster returns an EnvFunc that
// retrieves a previously saved k3d Cluster in the context (using the name), then deletes it.
//
// NOTE: this should be used in a Environment.Finish step.
func DestroyK3dCluster(name string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		clusterVal := ctx.Value(k3dContextKey(name))
		if clusterVal == nil {
			return ctx, fmt.Errorf("destroy k3d cluster func: context cluster is nil")
		}

		cluster, ok := clusterVal.(*k3d.Cluster)
		if !ok {
			return ctx, fmt.Errorf("destroy k3d cluster func: unexpected type for cluster value")
		}

		if err := cluster.Destroy(); err != nil {
			return ctx, fmt.Errorf("destroy k3d cluster: %w", err)
		}

		return ctx, nil
	}
}
