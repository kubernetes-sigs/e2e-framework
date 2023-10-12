/*
Copyright 2023 The Kubernetes Authors.

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

package cilium

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/third_party/helm"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	testEnv env.Environment
)

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	kindClusterName := "kind-with-cni"
	namespace := "cilium-test"
	testEnv.Setup(
		// Create kind cluster with custom config
		envfuncs.CreateClusterWithConfig(
			kind.NewProvider(),
			kindClusterName,
			"kind-config.yaml",
			kind.WithImage("kindest/node:v1.22.2")),
		// Create random namespace
		envfuncs.CreateNamespace(namespace),
		// Install Cilium via Helm
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			manager := helm.New(cfg.KubeconfigFile())
			err := manager.RunRepo(helm.WithArgs("add", "cilium", "https://helm.cilium.io/"))
			if err != nil {
				return nil, err
			}

			err = manager.RunInstall(
				helm.WithChart("cilium/cilium"),
				helm.WithNamespace("kube-system"),
				helm.WithArgs("--generate-name", "--set", "image.pullPolicy=IfNotPresent", "--set", "ipam.mode=kubernetes", "--wait"))
			if err != nil {
				return nil, err
			}

			// Wait for a worker node to be ready
			client, err := cfg.NewClient()
			if err != nil {
				return nil, err
			}

			node := &corev1.Node{
				ObjectMeta: metav1.ObjectMeta{Name: kindClusterName + "-worker"},
			}

			wait.For(conditions.New(client.Resources()).ResourceMatch(node, func(object k8s.Object) bool {
				d := object.(*corev1.Node)
				status := false
				for _, v := range d.Status.Conditions {
					if v.Type == "Ready" && v.Status == "True" {
						status = true
					}
				}
				return status
			}), wait.WithTimeout(time.Minute*2))
			return ctx, nil
		})

	testEnv.Finish(
		// Uninstall Cilium
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			manager := helm.New(cfg.KubeconfigFile())
			err := manager.RunRepo(helm.WithArgs("remove", "cilium"))
			if err != nil {
				return nil, err
			}
			return ctx, nil
		},
		envfuncs.DeleteNamespace(namespace),
		envfuncs.ExportClusterLogs(kindClusterName, "./logs"),
		envfuncs.DestroyCluster(kindClusterName),
	)
	os.Exit(testEnv.Run(m))
}
