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
	"flag"
	"k8s.io/client-go/util/homedir"
	"log"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/support/kind"
)

func SetupTestCluster(path string) (kc *kind.Cluster, kubeConfig string, cfg *rest.Config, clientSet kubernetes.Interface) {
	if path == "" {
		home := homedir.HomeDir()
		path = filepath.Join(home, ".kube", "config")
	}

	var err error
	kc, err = setupKind()
	if err != nil {
		log.Fatalln("error while setting up the kind cluster", err)
	}
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Paths to a kubeconfig. Only required if out-of-cluster.")
	err = flag.Set("kubeconfig", path)
	if err != nil {
		log.Fatalln("unexpected error while setting the flag value for kubeconfig", err)
	}
	flag.Parse()

	cfg, err = conf.New(conf.ResolveKubeConfigFile())
	if err != nil {
		log.Fatalln("error while client connection trying to resolve kubeconfig", err)
	}
	clientSet, err = kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalln("failed to create new Client set for kind cluster", err)
	}
	return
}

func DestroyTestCluster(kc *kind.Cluster) {
	err := kc.Destroy()
	if err != nil {
		log.Println("error while deleting the cluster", err)
		return
	}
}

func setupKind() (kc *kind.Cluster, err error) {
	kc = kind.NewCluster("e2e-test-cluster")
	if _, err = kc.Create(); err != nil {
		return
	}

	waitPeriod := 10 * time.Second
	log.Println("Waiting for kind pods to be initlaized...")
	time.Sleep(waitPeriod)
	return
}
