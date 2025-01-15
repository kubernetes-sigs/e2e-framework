/*
Copyright 2024 The Kubernetes Authors.

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

package k3d

import (
	"context"
	"fmt"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/stepfuncs"
	"sigs.k8s.io/e2e-framework/support"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func newDeployment(namespace string, name string, replicaCount int32) *appsv1.Deployment {
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "my-container",
				Image: "nginx",
			},
		},
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": "test-app"}},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test-app"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test-app"}},
				Spec:       podSpec,
			},
		},
	}
}

func TestK3DCluster(t *testing.T) {
	deploymentFeature := features.New("Should be able to create a new deployment in the k3d cluster").
		Assess("Create a new deployment", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			deployment := newDeployment(c.Namespace(), "test-deployment", 1)
			if err := c.Client().Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			var dep appsv1.Deployment
			if err := c.Client().Resources().Get(ctx, "test-deployment", c.Namespace(), &dep); err != nil {
				t.Fatal(err)
			}
			err := wait.For(conditions.New(c.Client().Resources()).DeploymentConditionMatch(&dep, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*3))
			if err != nil {
				t.Fatal(err)
			}
			return context.WithValue(ctx, "test-deployment", &dep)
		}).
		Feature()

	nodeAddFeature := features.New("Should be able to add a new node to the k3d cluster").
		Setup(stepfuncs.PerformNodeOperation(support.AddNode, &support.Node{
			Name:    fmt.Sprintf("%s-agent", clusterName),
			Cluster: clusterName,
			Role:    "agent",
		})).
		Assess("Check if the node is added to the cluster", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var node corev1.Node
			if err := c.Client().Resources().Get(ctx, fmt.Sprintf("k3d-%s-agent-0", clusterName), c.Namespace(), &node); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testEnv.Test(t, deploymentFeature, nodeAddFeature)
}
