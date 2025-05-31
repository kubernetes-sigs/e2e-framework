/*
Copyright 2025 The Kubernetes Authors.

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

package exec_in_deployment

import (
	"bytes"
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

type ctxKey string

func TestExecInDeployment(t *testing.T) {
	deploymentCtxKey := ctxKey("deployment")

	feature := features.New("ExecInDeployment").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			resources := c.Client().Resources()

			deployment := newDeployment(c.Namespace())
			if err := resources.Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}

			if err := wait.For(
				conditions.New(resources).DeploymentAvailable(deployment.Name, c.Namespace()),
				wait.WithTimeout(time.Minute*1)); err != nil {
				t.Fatal(err)
			}

			return context.WithValue(ctx, deploymentCtxKey, deployment)
		}).
		Assess("executes commands in an existing deployment", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			deployment := ctx.Value(deploymentCtxKey).(*appsv1.Deployment)
			message := "foo bar baz"

			var stdout, stderr bytes.Buffer
			cmd := []string{"echo", "-n", message}
			if err := c.Client().Resources().ExecInDeployment(ctx, c.Namespace(), deployment.Name, cmd, &stdout, &stderr); err != nil {
				t.Log(stderr.String())
				t.Fatal(err)
			}

			if stdout.String() != message {
				t.Fatalf("Expected %q, got %q", message, stdout.String())
			}

			return ctx
		}).
		Assess("returns error for non-existent deployments", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var stdout, stderr bytes.Buffer
			cmd := []string{"echo", "should not happen"}
			err := c.Client().Resources().ExecInDeployment(ctx, c.Namespace(), "does-not-exist", cmd, &stdout, &stderr)

			if err == nil {
				t.Fatal("Expected an error, got nil")
			}

			return ctx
		}).
		Feature()

	testEnv.Test(t, feature)
}

func newDeployment(namespace string) *appsv1.Deployment {
	labels := map[string]string{"app": "exec-in-deployment"}
	replicas := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{Containers: []corev1.Container{{
					Name:    "alpine",
					Image:   "alpine",
					Command: []string{"sleep", "infinity"},
				}}},
			},
		},
	}
}
