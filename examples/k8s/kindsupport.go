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

package k8s

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/vladimirvivien/gexe"
)

var (
	e           = gexe.New()
	kindVersion = "v0.11.0"
)

func createKindCluster(clusterName string) (string, error) {
	if e.Prog().Avail("kind") == "" {
		log.Println(`kind may not be installed, attempting to install...`)
		p := e.SetEnv("GO111MODULE", "on").RunProc(fmt.Sprintf("go get sigs.k8s.io/kind@%s", kindVersion))
		if p.Err() != nil {
			return "", fmt.Errorf("install kind failed: %s: %w", p.Result(), p.Err())
		}
		e.SetEnv("PATH", e.Run("echo $PATH:$GOPATH/bin"))
	}

	// create cluster
	p := e.RunProc(fmt.Sprintf(`kind create cluster --name %s`, clusterName))
	if p.Err() != nil {
		return "", fmt.Errorf("kind create cluster: %s: %w", p.Result(), p.Err())
	}

	// grab kubeconfig file for cluster
	kubecfg := fmt.Sprintf("%s-kubecfg", clusterName)
	p = e.StartProc(fmt.Sprintf(`kind get kubeconfig --name %s`, clusterName))
	if p.Err() != nil {
		return "", fmt.Errorf("kind get kubeconfig: %s: %w", p.Result(), p.Err())
	}

	file, err := ioutil.TempFile("", fmt.Sprintf("kind-cluser-%s", kubecfg))
	if err != nil {
		return "", fmt.Errorf("kind kubeconfig file: %w", err)
	}
	defer file.Close()
	if n, err := io.Copy(file, p.Out()); n == 0 || err != nil {
		return "", fmt.Errorf("kind kubecfg file: bytes copied: %d: %w]", n, err)
	}
	return file.Name(), nil
}

func deleteKindCluster(clusterName, kubeconfig string) error {
	p := e.RunProc(fmt.Sprintf(`kind delete cluster --name %s`, clusterName))
	if p.Err() != nil {
		return fmt.Errorf("kind delete cluster: %s: %w", p.Result(), p.Err())
	}
	if err := os.RemoveAll(kubeconfig); err != nil {
		return fmt.Errorf("kind: remove kubefconfig failed: %w", err)
	}
	return nil
}
