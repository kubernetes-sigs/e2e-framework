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

package wait

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"sigs.k8s.io/e2e-framework/internal/testutil"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"testing"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	cfg        *rest.Config
	kc         *kind.Cluster
	waitHelper *waiter
	resourceManager *resources.Resources
	namespace = "wait-test"
)

func TestMain(m *testing.M) {
	kc, _, cfg, _ = testutil.SetupTestCluster("")
	setup()
	exitCode := m.Run()
	tearDown()
	os.Exit(exitCode)
}

func setup() {
	resourceManager, err := resources.New(cfg)
	if err != nil {
		log.Fatalln("failed to create a resource manager instance", err)
	}

	resourceManager = resourceManager.WithNamespace(namespace)

	waitHelper = New(resourceManager)
	createNamespace(waitHelper.resources)
}

func tearDown() {
	deleteNamespace()
	testutil.DestroyTestCluster(kc)
}

func deleteNamespace() {
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	_ = waitHelper.resources.Delete(context.TODO(), namespace)
}

func createNamespace(r *resources.Resources) {
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	err := r.Create(context.TODO(), namespace)
	if err != nil {
		log.Fatalln("failed to create test namespace for wait helper test", err)
	}
}

func createPod(name string, r *resources.Resources) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": name}},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: name, Image: "nginx"},
			},
		},
	}
	return pod, r.Create(context.TODO(), pod)
}

func createDeployment(name string, r *resources.Resources) (*appsv1.Deployment, error) {
	var replica int32 = 1
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": name}},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replica,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: name, Image: "nginx"}}},
			},
		},
	}
	return deployment, r.Create(context.TODO(), deployment)
}