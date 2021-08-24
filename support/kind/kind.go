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

package kind

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/vladimirvivien/gexe"
)

var kindVersion = "v0.11.0"

type Cluster struct {
	name        string
	e           *gexe.Echo
	kubecfgFile string
	version     string
}

func NewCluster(name string) *Cluster {
	return &Cluster{name: name, e: gexe.New()}
}

// WithVersion set kind version
func (k *Cluster) WithVersion(ver string) *Cluster {
	k.version = ver
	return k
}

func (k *Cluster) Create() (string, error) {
	log.Println("Creating kind cluster ", k.name)
	// is kind program available
	if err := k.findOrInstallKind(k.e); err != nil {
		return "", err
	}

	if strings.Contains(k.e.Run("kind get clusters"), k.name) {
		log.Println("Skipping Kind Cluster.Create: cluster already created: ", k.name)
		return "", nil
	}

	// create kind cluster using kind-cluster-docker.yaml config file
	log.Println("launching: kind create cluster --name", k.name)
	p := k.e.RunProc(fmt.Sprintf(`kind create cluster --name %s`, k.name))
	if p.Err() != nil {
		return "", fmt.Errorf("failed to create kind cluster: %s : %s", p.Err(), p.Result())
	}

	clusters := k.e.Run("kind get clusters")
	log.Println("kind clusters available: ", clusters)

	// grab kubeconfig file for cluster
	kubecfg := fmt.Sprintf("%s-kubecfg", k.name)
	p = k.e.StartProc(fmt.Sprintf(`kind get kubeconfig --name %s`, k.name))
	if p.Err() != nil {
		return "", fmt.Errorf("kind get kubeconfig: %s: %w", p.Result(), p.Err())
	}

	file, err := ioutil.TempFile("", fmt.Sprintf("kind-cluser-%s", kubecfg))
	if err != nil {
		return "", fmt.Errorf("kind kubeconfig file: %w", err)
	}
	defer file.Close()

	k.kubecfgFile = file.Name()

	if n, err := io.Copy(file, p.Out()); n == 0 || err != nil {
		return "", fmt.Errorf("kind kubecfg file: bytes copied: %d: %w]", n, err)
	}

	return file.Name(), nil
}

// GetKubeconfig returns the path of the kubeconfig file
// associated with this kind cluster
func (k *Cluster) GetKubeconfig() string {
	return k.kubecfgFile
}

func (k *Cluster) GetKubeCtlContext() string {
	return fmt.Sprintf("kind-%s", k.name)
}

func (k *Cluster) Destroy() error {
	if err := k.findOrInstallKind(k.e); err != nil {
		return err
	}

	// deleteting kind cluster
	p := k.e.RunProc(fmt.Sprintf(`kind delete cluster --name %s`, k.name))
	if p.Err() != nil {
		return fmt.Errorf("failed to install kind: %s: %s", p.Err(), p.Result())
	}

	if err := os.RemoveAll(k.kubecfgFile); err != nil {
		return fmt.Errorf("kind: remove kubefconfig failed: %w", err)
	}

	return nil
}

func (k *Cluster) findOrInstallKind(e *gexe.Echo) error {
	if e.Prog().Avail("kind") == "" {
		log.Println(`kind not found, installing with GO111MODULE="on" go get sigs.k8s.io/kind@v0.11.0`)
		if err := k.installKind(e); err != nil {
			return err
		}
	}
	return nil
}

func (k *Cluster) installKind(e *gexe.Echo) error {
	if k.version != "" {
		kindVersion = k.version
	}

	log.Println("installing: go get sigs.k8s.io/kind@", kindVersion)
	p := e.SetEnv("GO111MODULE", "on").RunProc(fmt.Sprintf("go get sigs.k8s.io/kind@%s", kindVersion))
	if p.Err() != nil {
		return fmt.Errorf("failed to install kind: %s", p.Err())
	}

	if !p.IsSuccess() || p.ExitCode() != 0 {
		return fmt.Errorf("failed to install kind: %s", p.Result())
	}

	p = e.RunProc("ls $GOPATH/bin")
	if p.Err() != nil {
		return fmt.Errorf("failed to install kind: %s", p.Err())
	}

	p = e.RunProc("echo $PATH:$GOPATH/bin")
	if p.Err() != nil {
		return fmt.Errorf("failed to install kind: %s", p.Err())
	}

	e.SetEnv("PATH", p.Result())

	return nil
}
