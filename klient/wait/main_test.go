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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"sigs.k8s.io/e2e-framework/klient/internal/testutil"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sync"
	"testing"

	"k8s.io/client-go/rest"
)

var (
	tc *testutil.TestCluster
	cfg        *rest.Config
	resourceManager *resources.Resources
	namespace = "wait-test"
	resourceManagerOnce sync.Once
)

func TestMain(m *testing.M) {
	tc = testutil.SetupTestCluster("")
	cfg = tc.RESTConfig
	setup()
	exitCode := m.Run()
	tearDown()
	os.Exit(exitCode)
}

func setup() {
	createNamespace()
}

func getResourceManager() *resources.Resources {
	resourceManagerOnce.Do(func() {
		resourceMgr, err := resources.New(cfg)
		if err != nil {
			log.Fatalln("failed to create a resource manager instance", err)
		}
		resourceManager = resourceMgr.WithNamespace(namespace)
	})
	return resourceManager
}

func tearDown() {
	deleteNamespace()
	tc.DestroyTestCluster()
}

func deleteNamespace() {
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	_ = getResourceManager().Delete(context.TODO(), namespace)
}

func createNamespace() {
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}
	err := getResourceManager().Create(context.TODO(), namespace)
	if err != nil {
		log.Fatalln("failed to create test namespace for wait helper test", err)
	}
}

func createPod(name string, t *testing.T) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": name}},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: name, Image: "nginx"},
			},
		},
	}
	err := getResourceManager().Create(context.TODO(), pod)
	if err != nil {
		t.Error("failed to create pod due to an error", err)
	}
	return pod
}

func createDeployment(name string, replicas int32, t *testing.T) *appsv1.Deployment {
	 deployment := &appsv1.Deployment{
		 ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": name}},
		 Spec: appsv1.DeploymentSpec{
			 Replicas: &replicas,
			 Selector: &metav1.LabelSelector{
				 MatchLabels: map[string]string{"app": name},
			 },
			 Template: corev1.PodTemplateSpec{
				 ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				 Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: name, Image: "nginx"}}},
			 },
		 },
	 }
	 err := getResourceManager().Create(context.TODO(), deployment)
	 if err != nil {
		 t.Error("failed to create deployment due to an error", err)
	 }
	 return deployment
}

func createJob(name, cmd, arg string, t *testing.T) *batchv1.Job {
	var backOff int32 = 1
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": name}},
		Spec: batchv1.JobSpec{
			BackoffLimit: &backOff,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": name}},
				Spec:       corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
					{Name: name, Image: "alpine", Command: []string{cmd}, Args: []string{arg}},
				}},
			},
		},
	}
	err := getResourceManager().Create(context.TODO(), job)
	if err != nil {
		t.Error("failed to create a job due to an error", err)
	}
	return job
}