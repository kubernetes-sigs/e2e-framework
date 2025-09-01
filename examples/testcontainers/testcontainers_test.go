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

package testcontainers

import (
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

// define custom type for context keys to avoid collisions
type customCtxKey string

// define context key for storing deployment reference
const testDeploymentKey customCtxKey = "test-deployment"

var defaultLabels = map[string]string{"app": "test-app"} // default labels for deployment and pods

func TestDeployment(t *testing.T) {
	// define our features
	deploymentFeature := features.New("deployment with busybox and nginx").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// create a new deployment called test-deployment
			deployment := newDeployment("test-deployment", cfg.Namespace(), 1, defaultLabels)
			if err := cfg.Client().Resources().Create(ctx, deployment); err != nil {
				t.Fatalf("failed to create deployment test-deployment: %s", err)
			}
			return ctx
		}).
		Assess("wait for deployment available", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatalf("failed to create new client: %s", err)
			}
			dep := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "test-deployment", Namespace: cfg.Namespace()},
			}
			// wait for the deployment to finish becoming available
			err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(&dep, appsv1.DeploymentAvailable, corev1.ConditionTrue),
				wait.WithTimeout(time.Minute*3))
			if err != nil {
				t.Fatalf("failed waiting for deployment available: %s", err)
			}
			t.Logf("deployment %s is available", dep.Name)
			return ctx
		}).
		Assess("read deployment object", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var dep appsv1.Deployment
			err := cfg.Client().Resources().Get(ctx, "test-deployment", cfg.Namespace(), &dep)
			if err != nil {
				t.Fatalf("unable to reterieve deployment %s", err)
			}
			t.Logf("successfully read deployment %s", dep.Name)
			// inject deployment into context for use in teardown
			return context.WithValue(ctx, testDeploymentKey, &dep)
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// retrieve deployment from context
			depVal := ctx.Value(testDeploymentKey)
			if depVal == nil {
				t.Logf("no deployment found in context, skipping teardown")
				return ctx
			}
			deploy, ok := (depVal).(*appsv1.Deployment)
			if !ok {
				t.Logf("deployment in context has wrong type, skipping teardown")
				return ctx
			}
			if err := cfg.Client().Resources().Delete(ctx, deploy); err != nil {
				t.Fatalf("failed to delete deployment %s", err)
			}
			return ctx
		}).Feature()

	// run our features
	testEnv.Test(t, deploymentFeature)
}

// newDeployment creates a new deployment with the specified number of replicas and labels
// as part of this test, we should have already loaded the nginx and busybox image into the cluster
func newDeployment(name, namespace string, replicaCount int32, labels map[string]string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx",
						},
						{
							Name:  "busybox",
							Image: "busybox",
							Command: []string{
								"sleep", "3800",
							},
						},
					},
				},
			},
		},
	}
}
