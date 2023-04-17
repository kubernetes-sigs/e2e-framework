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

package testutil

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/support/kind"
)

type TestCluster struct {
	KindCluster *kind.Cluster
	Kubeconfig  string
	RESTConfig  *rest.Config
	Clientset   kubernetes.Interface
}

func SetupTestCluster(path string) *TestCluster {
	if path == "" {
		path = conf.ResolveKubeConfigFile()
	}

	tc := &TestCluster{}
	var err error
	kc, err := setupKind()
	if err != nil {
		log.Fatalln("error while setting up the kind cluster", err)
	}
	tc.KindCluster = kc

	cfg, err := conf.New(path)
	if err != nil {
		log.Fatalln("error while client connection trying to resolve kubeconfig", err)
	}
	tc.RESTConfig = cfg
	clientSet, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalln("failed to create new Client set for kind cluster", err)
	}
	tc.Clientset = clientSet
	if err := waitForControlPlane(clientSet); err != nil {
		log.Fatalln("failed to wait for Kind Cluster control-plane components", err)
	}
	return tc
}

func (t *TestCluster) DestroyTestCluster() {
	err := t.KindCluster.Destroy()
	if err != nil {
		log.ErrorS(err, "error while deleting the cluster")
		return
	}
}

func setupKind() (kc *kind.Cluster, err error) {
	kc = kind.NewCluster("e2e-test-cluster")
	if _, err = kc.Create(); err != nil {
		return
	}
	return
}

func waitForControlPlane(c kubernetes.Interface) error {
	selector, err := metav1.LabelSelectorAsSelector(
		&metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "component", Operator: metav1.LabelSelectorOpIn, Values: []string{"etcd", "kube-apiserver", "kube-controller-manager", "kube-scheduler"}},
			},
		},
	)
	if err != nil {
		return err
	}
	options := metav1.ListOptions{LabelSelector: selector.String()}
	log.Info("Waiting for kind control-plane pods to be initialized...")
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (done bool, err error) {
		pods, err := c.CoreV1().Pods("kube-system").List(ctx, options)
		if err != nil {
			return false, err
		}
		running := 0
		for i := range pods.Items {
			if pods.Items[i].Status.Phase == v1.PodRunning {
				running++
			}
		}
		// a kind cluster with one control-plane node will have 4 pods running the core apiserver components
		return running >= 4, nil
	})
	if err != nil {
		return err
	}

	selector, err = metav1.LabelSelectorAsSelector(
		&metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{Key: "k8s-app", Operator: metav1.LabelSelectorOpIn, Values: []string{"kindnet", "kube-dns", "kube-proxy"}},
			},
		},
	)
	if err != nil {
		return err
	}
	options = metav1.ListOptions{LabelSelector: selector.String()}
	log.Info("Waiting for kind networking pods to be initialized...")
	err = wait.PollUntilContextTimeout(context.Background(), 5*time.Second, 2*time.Minute, false, func(ctx context.Context) (done bool, err error) {
		pods, err := c.CoreV1().Pods("kube-system").List(ctx, options)
		if err != nil {
			return false, err
		}
		running := 0
		for i := range pods.Items {
			if pods.Items[i].Status.Phase == v1.PodRunning {
				running++
			}
		}
		// a kind cluster with one control-plane node will have 4 k8s-app pods running networking components
		return running >= 4, nil
	})
	return err
}
