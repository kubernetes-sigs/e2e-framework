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

package kwok

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

func TestKwokCluster(t *testing.T) {
	deploymentFeature := features.New("appsv1/deployment").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// setup node
			node := newNode("kwok-node")
			if err := cfg.Client().Resources().Create(ctx, node); err != nil {
				t.Fatal(err)
			}
			// start a deployment
			deployment := newDeployment(cfg.Namespace(), "test-deployment", 1)
			if err := cfg.Client().Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("deployment creation", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var dep appsv1.Deployment
			if err := cfg.Client().Resources().Get(ctx, "test-deployment", cfg.Namespace(), &dep); err != nil {
				t.Fatal(err)
			}
			// wait for the deployment to finish becoming available
			err := wait.For(conditions.New(cfg.Client().Resources()).DeploymentConditionMatch(&dep, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatal(err)
			}
			return context.WithValue(ctx, "test-deployment", &dep)
		}).
		Assess("pod ready", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			pods := &corev1.PodList{}
			err := cfg.Client().Resources(cfg.Namespace()).List(context.TODO(), pods)
			if err != nil || pods.Items == nil {
				t.Error("error while getting pods", err)
			}
			if len(pods.Items) != 1 {
				t.Fatal("could not find any pod in the namespace")
			}
			pod := pods.Items[0]
			// wait for the pod to be ready
			err = wait.For(conditions.New(cfg.Client().Resources()).PodConditionMatch(&pod, corev1.PodReady, corev1.ConditionTrue), wait.WithTimeout(time.Minute*1))
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			dep := ctx.Value("test-deployment").(*appsv1.Deployment)
			if err := cfg.Client().Resources().Delete(ctx, dep); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()
	testenv.Test(t, deploymentFeature)
}

func newDeployment(namespace string, name string, replicaCount int32) *appsv1.Deployment {
	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:  "my-container",
				Image: "nginx",
			},
		},
		Affinity: &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{
						{
							MatchExpressions: []corev1.NodeSelectorRequirement{
								{
									Key:      "type",
									Operator: corev1.NodeSelectorOpIn,
									Values:   []string{"kwok"},
								},
							},
						},
					},
				},
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

func newNode(nodeName string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nodeName,
			Labels: map[string]string{"beta.kubernetes.io/arch": "amd64", "beta.kubernetes.io/os": "linux", "type": "kwok"},
		},
		Spec: corev1.NodeSpec{},
	}
}
