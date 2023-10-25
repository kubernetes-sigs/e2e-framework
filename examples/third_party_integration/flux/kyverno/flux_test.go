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

func TestFluxRepoWorkflow(t *testing.T) {
	anyPod := features.New("Check creation of tenant resources under kyverno cluster policies").
		Assess("ensure privileged containers can't be deployed in the cluster", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			deployment := newDeployment("nginx-1", map[string]string{"app": "nginx"})
			client, err := c.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Second*30))
			t.Log(err)
			if err == nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()
	adminPod := features.New("Check creation of admin resources under kyverno cluster policies").
		Assess("ensure cluster admins can provision privileged containers", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			deployment := newDeployment("nginx-2", map[string]string{"app": "admin"})
			client, err := c.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Second*30))
			if err != nil {
				t.Fatal(err)
			}

			return ctx
		}).Feature()

	testEnv.Test(t, anyPod, adminPod)
}

func newDeployment(name string, labels map[string]string) *appsv1.Deployment {
	privileged := new(bool)
	*privileged = true
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "nginx",
						Image: "nginx",
						SecurityContext: &corev1.SecurityContext{
							Privileged: privileged,
						},
					}},
				},
			},
		},
	}
}
