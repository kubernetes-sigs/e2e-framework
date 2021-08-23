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

package resources

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/homedir"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	kubeconfig   string
	dep          *appsv1.Deployment
	clientset    *kubernetes.Clientset
	count        uint64
	replicaCount int32 = 2
	ctx                = context.TODO()
	cfg          *rest.Config
	namespace    *corev1.Namespace
	kc           *kind.Cluster
)

func TestMain(m *testing.M) {
	setup()
	initializeResObjects()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	home := homedir.HomeDir()
	path := filepath.Join(home, ".kube", "config")

	// set up kind cluster
	err := setupKindCluster()
	if err != nil {
		log.Println("error while setting up kind cluster", err)
		return
	}

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Paths to a kubeconfig. Only required if out-of-cluster.")

	// set --kubeconfig flag
	err = flag.Set("kubeconfig", path)
	if err != nil {
		log.Println("unexpected error while setting flag value", err)
		return
	}

	flag.Parse()

	cfg, err = conf.New(conf.ResolveKubeConfigFile())
	if err != nil {
		log.Println("error while client connection", err)
		return
	}

	clientset, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Println("error while client set connection", err)
		return
	}
}

// setupKindCluster
func setupKindCluster() error {
	kc = kind.NewCluster("e2e-test-cluster")
	if _, err := kc.Create(); err != nil {
		return err
	}

	// stall to wait for kind pods initialization
	waitTime := time.Second * 10
	log.Println("waiting for kind pods to initialize...", waitTime)
	time.Sleep(waitTime)

	return nil
}

func teardown() {
	deleteDeployment(ctx, dep, namespace.Name)
	deleteNamespace(ctx, namespace)

	// delete kind cluster
	err := kc.Destroy()
	if err != nil {
		log.Println("error while deleting the cluster", err)
		return
	}
}

func deleteDeployment(ctx context.Context, dep *appsv1.Deployment, ns string) {
	_, err := clientset.AppsV1().Deployments(ns).Get(ctx, dep.Name, metav1.GetOptions{})
	if err == nil {
		err = clientset.AppsV1().Deployments(ns).Delete(ctx, dep.Name, metav1.DeleteOptions{})
		if err != nil {
			log.Println("error while deleting deployment", err)
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
		log.Println("error while deleting namespace", err)
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
