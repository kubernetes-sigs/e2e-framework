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

package kyverno

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
	"sigs.k8s.io/e2e-framework/third_party/flux"
)

var (
	testEnv         env.Environment
	namespace       string
	kindClusterName string
)

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	kindClusterName = envconf.RandomName("flux", 10)
	namespace = envconf.RandomName("flux", 10)
	helmRepoName := "kyverno"
	hrNameKyverno := "kyverno"
	hrNamePolicies := "kyverno-policies"
	testEnv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		envfuncs.CreateNamespace(namespace),
		flux.InstallFlux(),
		flux.CreateHelmRepository(helmRepoName, "https://kyverno.github.io/kyverno/"),
		flux.CreateHelmRelease(hrNameKyverno, "HelmRepository/"+helmRepoName, "kyverno", flux.WithArgs("--target-namespace", "kyverno", "--create-target-namespace")),
		flux.CreateHelmRelease(hrNamePolicies, "HelmRepository/"+helmRepoName, "kyverno-policies", flux.WithArgs("--target-namespace", "kyverno", "--values", "kyverno-values.yaml")),
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			return ctx, waitForKyvernoDeployments(cfg)
		},
	)

	testEnv.Finish(
		flux.DeleteHelmRelease(hrNamePolicies),
		flux.DeleteHelmRelease(hrNameKyverno),
		flux.DeleteHelmRepo(helmRepoName),
		flux.UninstallFlux(),
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyCluster(kindClusterName),
	)
	os.Exit(testEnv.Run(m))
}

// waitForKyvernoDeployments - waits for all the deployments related to kyverno to be available
func waitForKyvernoDeployments(c *envconf.Config) error {
	client, err := c.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	kyvernoDeployments := []string{
		"kyverno-admission-controller",
		"kyverno-background-controller",
		"kyverno-cleanup-controller",
		"kyverno-reports-controller",
	}

	for _, deploymentName := range kyvernoDeployments {
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      deploymentName,
				Namespace: "kyverno",
			},
		}

		err := wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionTrue),
			wait.WithTimeout(time.Minute*2), wait.WithInterval(time.Second*5))
		if err != nil {
			return fmt.Errorf("deployment %s did not become available: %w", deploymentName, err)
		}
	}

	return nil
}
