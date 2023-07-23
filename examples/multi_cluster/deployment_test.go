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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

var curDir, _ = os.Getwd()

func checkPodStatus(t *testing.T, kubeConfig string, clusterName string) {
	t.Helper()
	client, err := klient.NewWithKubeConfigFile(kubeConfig)
	if err != nil {
		t.Errorf("ran into an error trying to create a client for Cluster %s", clusterName)
	}
	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{},
	}
	err = wait.For(conditions.New(client.Resources()).ResourceScaled(deployment, func(object k8s.Object) int32 {
		return object.(*appsv1.Deployment).Status.ReadyReplicas
	}, 1))
	if err != nil {
		t.Fatal("failed waiting for the Deployment to reach a ready state")
	}
}

func TestScenarioOne(t *testing.T) {
	feature := features.New("Scenario One").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			for _, clusterName := range clusterNames {
				cluster, ok := envfuncs.GetKindClusterFromContext(ctx, clusterName)
				if !ok {
					t.Fatalf("Failed to extract kind cluster %s from context", clusterName)
				}
				manager := helm.New(cluster.GetKubeconfig())
				err := manager.RunInstall(helm.WithName("example"), helm.WithNamespace("default"), helm.WithChart(filepath.Join(curDir, "testdata", "example_chart")), helm.WithWait(), helm.WithTimeout("10m"))
				if err != nil {
					t.Fatal("failed to invoke helm install operation due to an error", err)
				}
			}
			return ctx
		}).
		Assess(fmt.Sprintf("Deployment is running successfully - %s", clusterNames[0]), func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			cluster, ok := envfuncs.GetKindClusterFromContext(ctx, clusterNames[0])
			if !ok {
				t.Fatalf("Failed to extract kind cluster %s from context", clusterNames[0])
			}
			checkPodStatus(t, cluster.GetKubeconfig(), clusterNames[0])
			return ctx
		}).
		Assess(fmt.Sprintf("Deployment is running successfully - %s", clusterNames[1]), func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			cluster, ok := envfuncs.GetKindClusterFromContext(ctx, clusterNames[1])
			if !ok {
				t.Fatalf("Failed to extract kind cluster %s from context", clusterNames[1])
			}
			checkPodStatus(t, cluster.GetKubeconfig(), clusterNames[1])
			return ctx
		}).
		Feature()

	_ = testEnv.Test(t, feature)
}
