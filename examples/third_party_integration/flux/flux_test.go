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

package flux

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/flux"
	"testing"
	"time"
)

var curDir, _ = os.Getwd()

func TestFluxRepoWorkflow(t *testing.T) {
	gitRepoName := "hello-world"
	kustomizationName := "hello-world"

	feature := features.New("Install resources by flux").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			manager := flux.New(c.KubeconfigFile())
			err := manager.CreateGitRepo(gitRepoName, "https://github.com/matrus2/go-hello-world", flux.WithBranch("main"))
			if err != nil {
				t.Fatal("failed to create git repository", err)
			}
			// Set --prune so that once deleteKustomization will be called all corresponding resources will be removed
			err = manager.CreateKustomization(kustomizationName, "GitRepository/hello-world.flux-system", flux.WithPath("template"), flux.WithArgs("--target-namespace", c.Namespace(), "--prune"))
			if err != nil {
				t.Fatal("failed to create kustomization", err)
			}

			return ctx
		}).
		Assess("check if deployment was successful", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			deployment := &appsv1.Deployment{
				ObjectMeta: v1.ObjectMeta{
					Name:      "hello-app",
					Namespace: c.Namespace(),
				},
				Spec: appsv1.DeploymentSpec{},
			}

			err := wait.For(conditions.New(c.Client().Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionStatus(v1.ConditionTrue)), wait.WithTimeout(time.Minute*5))
			if err != nil {
				t.Fatal("Error deployment not found", err)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			manager := flux.New(c.KubeconfigFile())
			err := manager.DeleteGitRepo(gitRepoName)
			if err != nil {
				t.Fatal("failed to delete git repository", err)
			}
			err = manager.DeleteKustomization(kustomizationName)
			if err != nil {
				t.Fatal("failed to delete kustomization", err)
			}

			return ctx
		}).Feature()

	testEnv.Test(t, feature)
}
