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

package resources_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	log "k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	dep          *appsv1.Deployment
	clientset    kubernetes.Interface
	count        uint64
	replicaCount int32 = 2
	ctx                = context.TODO()
	cfg          *rest.Config
	namespace    *corev1.Namespace
)

func TestMain(m *testing.M) {
	testenv := env.New()
	kindClusterName := envconf.RandomName("resources-test-", 16)
	testenv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			cfg = c.Client().RESTConfig()
			var err error
			clientset, err = kubernetes.NewForConfig(cfg)
			if err != nil {
				return ctx, err
			}
			initializeResObjects()
			return ctx, nil
		},
	)
	testenv.Finish(
		func(ctx context.Context, c *envconf.Config) (context.Context, error) {
			deleteDeployment(ctx, dep, namespace.Name)
			deleteNamespace(ctx, namespace)
			return ctx, nil
		},
		envfuncs.DestroyCluster(kindClusterName),
	)

	os.Exit(testenv.Run(m))
}

func deleteDeployment(ctx context.Context, dep *appsv1.Deployment, ns string) {
	_, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
	if err == nil {
		err = clientset.AppsV1().Deployments(ns).Delete(ctx, dep.Name, metav1.DeleteOptions{})
		if err != nil {
			log.ErrorS(err, "error while deleting deployment")
		}
	}
}

func deleteNamespace(ctx context.Context, ns *corev1.Namespace) {
	ns, err := clientset.CoreV1().Namespaces().Get(ctx, ns.Name, metav1.GetOptions{})
	if err != nil {
		return
	}

	err = clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{})
	if err != nil {
		log.ErrorS(err, "error while deleting namespace")
	}
}

func initializeResObjects() {
	namespace = &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test"}}
	dep = getDeployment(fmt.Sprintf("deployment-name-%v", count))
}

func getDeployment(name string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace.Name, Labels: map[string]string{"app": "test-app"}},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicaCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"foo": "bar"}},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "nginx", Image: "nginx"}}},
			},
		},
	}
}
