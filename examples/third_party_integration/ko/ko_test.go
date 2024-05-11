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

package ko

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/ko"
)

var packagePath = "./" + filepath.Join(curDir, "testdata", "example_goapp")

func TestBuildLocalKind(t *testing.T) {
	feature := features.New("ko build with local kind cluster").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			var err error

			// Apple Silicon with Podman as kind provider need to build with linux/arm64
			platform := fmt.Sprintf("linux/%s", runtime.GOARCH)

			manager := ko.New()
			err = manager.Install("latest")
			if err != nil {
				t.Fatalf("failed to install ko: %v", err)
			}

			ctx, err = manager.BuildLocal(ctx, packagePath, ko.WithLocalKindName(kindClusterName), ko.WithPlatforms(platform))
			if err != nil {
				t.Fatalf("failed to build with local kind: %v", err)
			}

			return ctx
		}).
		Assess("Deployment is running successfully", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			manager := ko.New()
			goappImage, err := manager.GetLocalImage(ctx, packagePath)
			if err != nil {
				t.Fatalf("failed to get previous built image: %v", err)
			}

			deployment := newDeployment(config.Namespace(), goappImage, 1)
			client, err := config.NewClient()
			if err != nil {
				t.Fatalf("failed to init k8s client: %v", err)
			}
			if err = client.Resources().Create(ctx, deployment); err != nil {
				t.Fatalf("failed to create deployment: %v", err)
			}
			err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*5))
			if err != nil {
				t.Fatalf("failed to wait for deployment to be ready: %v", err)
			}

			return ctx
		}).Feature()

	_ = testEnv.Test(t, feature)
}

func newDeployment(namespace, image string, replicas int32) *appsv1.Deployment {
	labels := map[string]string{"app": "goapp"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "goapp", Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "goapp", Image: image}}},
			},
		},
	}
}
