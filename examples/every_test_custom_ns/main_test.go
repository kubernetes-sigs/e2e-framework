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

package every_test_custom_ns

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/support/kind"
)

type (
	NamespaceCtxKey string
	ClusterCtxKey   string
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()

	// Specifying a run ID so that multiple runs wouldn't collide.
	runID := envconf.RandomName("ns", 4)

	testenv.Setup(
		// Step: creates kind cluster, propagate kind cluster object
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			name := envconf.RandomName("my-cluster", 16)
			cluster := kind.NewCluster(name)
			kubeconfig, err := cluster.Create(ctx)
			if err != nil {
				return ctx, err
			}
			// stall a bit to allow most pods to come up
			time.Sleep(time.Second * 10)

			// update environment with kubecofig file
			cfg.WithKubeconfigFile(kubeconfig)

			// propagate cluster value
			return context.WithValue(ctx, ClusterCtxKey("cluster"), cluster), nil
		}).Finish(
		// Teardown func: delete kind cluster
		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			cluster := ctx.Value(ClusterCtxKey("cluster")).(*kind.Cluster) // nil should be tested
			if cluster == nil {
				return ctx, fmt.Errorf("error getting kind cluster from context")
			}
			if err := cluster.Destroy(ctx); err != nil {
				return ctx, err
			}
			return ctx, nil
		},
	)

	testenv.BeforeEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		return createNSForTest(ctx, cfg, t, runID)
	})
	testenv.AfterEachTest(func(ctx context.Context, cfg *envconf.Config, t *testing.T) (context.Context, error) {
		return deleteNSForTest(ctx, cfg, t, runID)
	})

	os.Exit(testenv.Run(m))
}

// CreateNSForTest creates a random namespace with the runID as a prefix. It is stored in the context
// so that the deleteNSForTest routine can look it up and delete it.
func createNSForTest(ctx context.Context, cfg *envconf.Config, t *testing.T, runID string) (context.Context, error) {
	ns := envconf.RandomName(runID, 10)
	ctx = context.WithValue(ctx, GetNamespaceKey(t), ns)

	t.Logf("Creating NS %v for test %v", ns, t.Name())
	nsObj := v1.Namespace{}
	nsObj.Name = ns
	return ctx, cfg.Client().Resources().Create(ctx, &nsObj)
}

// DeleteNSForTest looks up the namespace corresponding to the given test and deletes it.
func deleteNSForTest(ctx context.Context, cfg *envconf.Config, t *testing.T, _ string) (context.Context, error) {
	ns := fmt.Sprint(ctx.Value(GetNamespaceKey(t)))
	t.Logf("Deleting NS %v for test %v", ns, t.Name())

	nsObj := v1.Namespace{}
	nsObj.Name = ns
	return ctx, cfg.Client().Resources().Delete(ctx, &nsObj)
}

// GetNamespaceKey returns the context key for a given test
func GetNamespaceKey(t *testing.T) NamespaceCtxKey {
	// When we pass t.Name() from inside an `assess` step, the name is in the form TestName/Features/Assess
	if strings.Contains(t.Name(), "/") {
		return NamespaceCtxKey(strings.Split(t.Name(), "/")[0])
	}

	// When pass t.Name() from inside a `testenv.BeforeEachTest` function, the name is just TestName
	return NamespaceCtxKey(t.Name())
}
