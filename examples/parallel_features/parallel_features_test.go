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

package parallel_features

import (
	"context"
	"os"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	clusterName string
	namespace   string
	testEnv     env.Environment
)

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)

	clusterName = envconf.RandomName("kind-parallel", 16)
	namespace = envconf.RandomName("kind-ns", 16)

	testEnv.Setup(
		envfuncs.CreateKindCluster(clusterName),
		envfuncs.CreateNamespace(namespace),
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyKindCluster(clusterName),
	)
	os.Exit(testEnv.Run(m))
}

func newDeployment(namespace string, name string, replicaCount int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": name}},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: name, Image: "nginx"}}},
			},
		},
	}
}

func TestPodBringUp(t *testing.T) {
	featureOne := features.New("Feature One").
		Assess("Create Nginx Deployment 1", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			deployment := newDeployment(namespace, "deployment-1", 2)
			err := config.Client().Resources().Create(ctx, deployment)
			if err != nil {
				t.Error("failed to create test pod for deployment-1")
			}
			ctx = context.WithValue(ctx, "DEPLOYMENT", deployment)
			return ctx
		}).
		Assess("Wait for Nginx Deployment 1 to be scaled up", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			deployment := ctx.Value("DEPLOYMENT").(*appsv1.Deployment)
			err := wait.For(conditions.New(config.Client().Resources()).ResourceScaled(deployment, func(object k8s.Object) int32 {
				return object.(*appsv1.Deployment).Status.ReadyReplicas
			}, 2))
			if err != nil {
				t.Error("failed waiting for deployment to be scaled up")
			}
			return ctx
		}).Feature()

	featureTwo := features.New("Feature Two").
		Assess("Create Nginx Deployment 2", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			deployment := newDeployment(namespace, "deployment-2", 2)
			err := config.Client().Resources().Create(ctx, deployment)
			if err != nil {
				t.Error("failed to create test pod for deployment-2")
			}
			ctx = context.WithValue(ctx, "DEPLOYMENT", deployment)
			return ctx
		}).
		Assess("Wait for Nginx Deployment 2 to be scaled up", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			deployment := ctx.Value("DEPLOYMENT").(*appsv1.Deployment)
			err := wait.For(conditions.New(config.Client().Resources()).ResourceScaled(deployment, func(object k8s.Object) int32 {
				return object.(*appsv1.Deployment).Status.ReadyReplicas
			}, 2))
			if err != nil {
				t.Error("failed waiting for deployment to be scaled up")
			}
			return ctx
		}).Feature()

	_ = testEnv.TestInParallel(t, featureOne, featureTwo)
}
