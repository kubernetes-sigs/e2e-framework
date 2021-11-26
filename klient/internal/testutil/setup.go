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
	"time"

	log "k8s.io/klog/v2"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

	waitPeriod := 10 * time.Second
	log.Info("Waiting for kind pods to be initialized...")
	time.Sleep(waitPeriod)
	return
}
