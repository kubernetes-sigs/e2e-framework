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

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	kindsupport "sigs.k8s.io/e2e-framework/support/kind"
)

var (
	testenv env.Environment
)

func TestMain(m *testing.M) {
	testenv = env.New()
	testenv.Setup(CreateKindFunc(envconf.RandomName("my-cluster", 16))).Finish(CleanupKindFunc())
	os.Exit(testenv.Run(m))
}

// CreateKindFunc returns an EnvFunc
// that creates kind cluster, and propagate
// the kind instance via the environment context
func CreateKindFunc(name string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		kind := kindsupport.NewCluster(name)
		kubecfg, err := kind.Create()
		if err != nil {
			return ctx, fmt.Errorf("kind setup: %w", err)
		}

		// stall, give pods time
		time.Sleep(7 * time.Second)

		// update envconfig  with kubeconfig
		if _, err := cfg.WithKubeconfigFile(kubecfg); err != nil {
			return ctx, fmt.Errorf("kind setup: envconfig: %w", err)
		}

		// forward cluster name and kubecfg file name (via context)
		// for cleanup later
		return context.WithValue(ctx, "kindcluster", kind), nil
	}
}

// CleanupKindFunc returns an EnvFunc that
// retrieves the kind cluster instance from the forwarded context
// then deletes the cluster.
func CleanupKindFunc() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		kind := ctx.Value("kindcluster").(*kindsupport.Cluster)
		if err := kind.Destroy(); err != nil {
			return ctx, fmt.Errorf("kind cleanup: %w", err)
		}
		return ctx, nil
	}
}
