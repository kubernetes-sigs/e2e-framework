/*
Copyright 2023 The Kubernetes Authors.

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

package kwok

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	log "k8s.io/klog/v2"

	"github.com/vladimirvivien/gexe"
)

var kwokVersion = "v0.2.0"

type Cluster struct {
	name         string
	e            *gexe.Echo
	kubecfgFile  string
	version      string
	waitDuration time.Duration
}

func NewCluster(name string) *Cluster {
	return &Cluster{name: name, e: gexe.New(), waitDuration: 1 * time.Minute}
}

// WithVersion set kwok version
func (k *Cluster) WithVersion(ver string) *Cluster {
	k.version = ver
	return k
}

// WithWaitDuration set duration to wait for kwok cluster to be ready
func (k *Cluster) WithWaitDuration(d time.Duration) *Cluster {
	k.waitDuration = d
	return k
}

func (k *Cluster) getKubeconfig() (string, error) {
	kubecfg := fmt.Sprintf("%s-kubecfg", k.name)

	p := k.e.RunProc(fmt.Sprintf(`kwokctl get kubeconfig --name %s`, k.name))
	if p.Err() != nil {
		return "", fmt.Errorf("kwokctl get kubeconfig: %w", p.Err())
	}

	var stdout bytes.Buffer
	if _, err := stdout.ReadFrom(p.Out()); err != nil {
		return "", fmt.Errorf("kwokctl kubeconfig stdout bytes: %w", err)
	}

	file, err := os.CreateTemp("", fmt.Sprintf("kwok-cluster-%s", kubecfg))
	if err != nil {
		return "", fmt.Errorf("kwok kubeconfig file: %w", err)
	}
	defer file.Close()

	k.kubecfgFile = file.Name()

	if n, err := io.Copy(file, &stdout); n == 0 || err != nil {
		return "", fmt.Errorf("kwok kubecfg file: bytes copied: %d: %w]", n, err)
	}

	return file.Name(), nil
}

func (k *Cluster) clusterExists(name string) (string, bool) {
	clusters := k.e.Run("kwokctl get clusters")
	for _, c := range strings.Split(clusters, "\n") {
		if c == name {
			return clusters, true
		}
	}
	return clusters, false
}

func (k *Cluster) CreateWithConfig(kwokConfigFile string) (string, error) {
	return k.Create("--config", kwokConfigFile)
}

func (k *Cluster) Create(args ...string) (string, error) {
	log.V(4).Info("Creating kwok cluster ", k.name)
	if err := k.findOrInstallKwok(k.e); err != nil {
		return "", err
	}

	if _, ok := k.clusterExists(k.name); ok {
		log.V(4).Info("Skipping kwok Cluster.Create: cluster already created: ", k.name)
		return k.getKubeconfig()
	}

	command := fmt.Sprintf(`kwokctl create cluster --name %s --wait %s`, k.name, k.waitDuration.String())
	if len(args) > 0 {
		command = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	}
	log.V(4).Info("Launching:", command)
	p := k.e.RunProc(command)
	if p.Err() != nil {
		return "", fmt.Errorf("failed to create kwok cluster: %s : %s", p.Err(), p.Result())
	}

	clusters, ok := k.clusterExists(k.name)
	if !ok {
		return "", fmt.Errorf("kwok Cluster.Create: cluster %v still not in 'cluster list' after creation: %v", k.name, clusters)
	}
	log.V(4).Info("kwok cluster available: ", clusters)

	// Grab kubeconig file for cluster.
	return k.getKubeconfig()
}

// GetKubeconfig returns the path of the kubeconfig file
// associated with this kwok cluster
func (k *Cluster) GetKubeconfig() string {
	return k.kubecfgFile
}

func (k *Cluster) GetKubeCtlContext() string {
	return fmt.Sprintf("kwok-%s", k.name)
}

func (k *Cluster) Destroy() error {
	log.V(4).Info("Destroying kwok cluster ", k.name)
	if err := k.findOrInstallKwok(k.e); err != nil {
		return err
	}

	p := k.e.RunProc(fmt.Sprintf(`kwokctl delete cluster --name %s`, k.name))
	if p.Err() != nil {
		return fmt.Errorf("kwok: delete cluster failed: %s: %s", p.Err(), p.Result())
	}

	log.V(4).Info("Removing kubeconfig file ", k.kubecfgFile)
	if err := os.RemoveAll(k.kubecfgFile); err != nil {
		return fmt.Errorf("kwok: remove kubefconfig failed: %w", err)
	}
	return nil
}

func (k *Cluster) findOrInstallKwok(e *gexe.Echo) error {
	if k.version != "" {
		kwokVersion = k.version
	}
	// installing kwok means installing kwokctl binary

	if e.Prog().Avail("kwokctl") == "" {
		log.V(4).Info(`kwokctl not found, installing version @%s`, kwokVersion)
		if err := k.installKwokCtl(e); err != nil {
			return err
		}
	}
	// PATH may already be set to include $GOPATH/bin so we don't need to.
	if kwokCtlPath := e.Prog().Avail("kwokctl"); kwokCtlPath != "" {
		log.V(4).Info("Installed kwokctl at", kwokCtlPath)
		return nil
	}

	p := e.RunProc("ls $GOPATH/bin")
	if p.Err() != nil {
		return fmt.Errorf("failed to install kwokctl: %s", p.Err())
	}

	p = e.RunProc("echo $PATH:$GOPATH/bin")
	if p.Err() != nil {
		return fmt.Errorf("failed to install kwokctl: %s", p.Err())
	}

	log.V(4).Info(`Setting path to include $GOPATH/bin:`, p.Result())
	e.SetEnv("PATH", p.Result())

	if kwokCtlPath := e.Prog().Avail("kwokctl"); kwokCtlPath != "" {
		log.V(4).Info("Installed kwokctl at", kwokCtlPath)
		return nil
	}
	return fmt.Errorf("kwokctl not available even after installation")
}

func (k *Cluster) installKwokCtl(e *gexe.Echo) error {
	log.V(4).Info("Installing: go install sigs.k8s.io/kwok/cmd/kwokctl@%s", kwokVersion)
	installKwokCtlCmd := fmt.Sprintf("go install sigs.k8s.io/kwok/cmd/kwokctl@%s", kwokVersion)
	log.V(4).Info("%s", installKwokCtlCmd)
	p := e.RunProc(installKwokCtlCmd)
	if p.Err() != nil {
		return fmt.Errorf("failed to install kwokctl: %s %s", installKwokCtlCmd, p.Err())
	}

	if !p.IsSuccess() || p.ExitCode() != 0 {
		return fmt.Errorf("failed to install kwokctl: %s", p.Result())
	}
	return nil
}
