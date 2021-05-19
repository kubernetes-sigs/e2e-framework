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

package testing

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/vladimirvivien/gexe"
)

var kindVersion = "v0.11.0"

type KindCluster struct {
	name string
	e    *gexe.Echo
}

func NewKindCluster(name string) *KindCluster {
	return &KindCluster{name: name, e: gexe.New()}
}

func (k *KindCluster) Create() error {
	log.Println("Creating kind cluster ", k.name)
	// is kind program available
	if err := findOrInstallKind(k.e); err != nil {
		return err
	}

	if strings.Contains(k.e.Run("kind get clusters"), k.name) {
		log.Println("Skipping KindCluster.Create: cluster already created: ", k.name)
		return nil
	}

	// create kind cluster using kind-cluster-docker.yaml config file
	log.Println("launching: kind create cluster --name", k.name)
	p := k.e.RunProc(fmt.Sprintf(`kind create cluster --name %s`, k.name))
	if p.Err() != nil {
		return fmt.Errorf("failed to create kind cluster: %s : %s", p.Err(), p.Result())
	}

	clusters := k.e.Run("kind get clusters")
	log.Println("kind clusters available: ", clusters)

	return nil
}

func (k *KindCluster) GetKubeConfig() (io.Reader, error) {
	fmt.Println("Retrieving kind kubeconfig for cluster: ", k.name)
	p := k.e.RunProc(fmt.Sprintf(`kind get kubeconfig --name %s`, k.name))
	if p.Err() != nil {
		return nil, p.Err()
	}

	return p.Out(), nil
}

func (k *KindCluster) MakeKubeConfigFile(path string) error {
	fmt.Println("Creating kind kubeconfig file: ", path)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to initialize kind kubeconfig file: %s", err)
	}
	defer f.Close()

	reader, err := k.GetKubeConfig()
	if err != nil {
		return fmt.Errorf("failed to generate kind kubeconfig: %s", err)
	}

	var num int64
	if num, err = io.Copy(f, reader); err != nil {
		return fmt.Errorf("failed to write kind kubeconfig file: %s", err)
	}

	fmt.Println("written bytes", num)
	return nil
}

func (k *KindCluster) GetKubeCtlContext() string {
	return fmt.Sprintf("kind-%s", k.name)
}

func (k *KindCluster) Destroy() error {
	fmt.Println("Destroying kind cluster :", k.name)
	if err := findOrInstallKind(k.e); err != nil {
		return err
	}
	// deleteting kind cluster
	p := k.e.RunProc(fmt.Sprintf(`kind delete cluster --name %s`, k.name))
	if p.Err() != nil {
		return fmt.Errorf("failed to install kind: %s: %s", p.Err(), p.Result())
	}

	fmt.Println("Kind cluster destroyed")

	clusters := k.e.Run("kind get clusters")
	fmt.Println("Available kind clusters: ", clusters)

	return nil
}

func findOrInstallKind(e *gexe.Echo) error {
	if e.Prog().Avail("kind") == "" {
		fmt.Println(`kind not found, installing with GO111MODULE="on" go get sigs.k8s.io/kind@v0.11.0`)
		if err := installKind(e); err != nil {
			return err
		}
	}
	return nil
}

func installKind(e *gexe.Echo) error {
	fmt.Println("installing: go get sigs.k8s.io/kind@", kindVersion)
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

	fmt.Println("ls result:", p.Result())

	p = e.RunProc("echo $PATH:$GOPATH/bin")
	if p.Err() != nil {
		return fmt.Errorf("failed to install kind: %s", p.Err())
	}

	fmt.Println("path environment detail:", p.Result())
	e.SetEnv("PATH", p.Result())

	return nil
}
