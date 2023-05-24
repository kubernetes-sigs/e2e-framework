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

package k3d

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/vladimirvivien/gexe"
	log "k8s.io/klog/v2"
)

var k3dVersion = "v5.4.3"

type Cluster struct {
	name        string
	e           *gexe.Echo
	kubecfgFile string
	version     string
}

func NewCluster(name string) *Cluster {
	return &Cluster{name: name, e: gexe.New()}
}

// WithVersion set k3d version
func (k *Cluster) WithVersion(ver string) *Cluster {
	k.version = ver
	return k
}

func (k *Cluster) getKubeconfig() (string, error) {
	kubecfg := fmt.Sprintf("%s-kubecfg", k.name)

	p := k.e.RunProc(fmt.Sprintf(`k3d kubeconfig get %s`, k.name))
	if p.Err() != nil {
		return "", fmt.Errorf("k3d kubeconfig get: %w", p.Err())
	}
	var stdout bytes.Buffer
	if _, err := stdout.ReadFrom(p.Out()); err != nil {
		return "", fmt.Errorf("k3d kubeconfig stdout bytes: %w", err)
	}

	file, err := os.CreateTemp("", fmt.Sprintf("k3d-cluser-%s", kubecfg))
	if err != nil {
		return "", fmt.Errorf("k3d kubeconfig file: %w", err)
	}
	defer file.Close()

	k.kubecfgFile = file.Name()

	if n, err := io.Copy(file, &stdout); n == 0 || err != nil {
		return "", fmt.Errorf("k3d kubecfg file: bytes copied: %d: %w]", n, err)
	}

	return file.Name(), nil
}

func (k *Cluster) clusterExists(name string) (string, bool) {
	clustersList := []string{}
	clustersOutput := k.e.Run("k3d cluster list")
	for _, c := range strings.Split(clustersOutput, "\n")[1:] {
		// split on whitespace or tabs
		parts := strings.Fields(c)
		if len(parts) < 2 {
			continue
		}
		clustersList = append(clustersList, parts[0])
	}

	clustersListJoined := strings.Join(clustersList, " ")
	for _, c := range clustersList {
		if c == name {
			return clustersListJoined, true
		}
	}
	return clustersListJoined, false
}

func (k *Cluster) CreateWithConfig(imageName, k3dConfigFile string) (string, error) {
	return k.Create("--image", imageName, "--config", k3dConfigFile)
}

func (k *Cluster) Create(args ...string) (string, error) {
	log.V(4).Info("Creating k3d cluster ", k.name)
	if err := k.findOrInstallK3d(k.e); err != nil {
		return "", err
	}

	if _, ok := k.clusterExists(k.name); ok {
		log.V(4).Infof("Skipping k3d cluster creation: cluster %q already exists: ", k.name)
		return k.getKubeconfig()
	}

	command := fmt.Sprintf(`k3d cluster create --wait --no-lb %s`, k.name)
	if len(args) > 0 {
		command = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	}
	log.V(4).Infof("Launching %q", command)
	p := k.e.RunProc(command)
	if p.Err() != nil {
		return "", fmt.Errorf("failed to create k3d cluster: %s : %s", p.Err(), p.Result())
	}

	clusters, ok := k.clusterExists(k.name)
	if !ok {
		return "", fmt.Errorf("k3d Cluster.Create: cluster %v still not in 'cluster list' after creation: %v", k.name, clusters)
	}
	log.V(4).Info("k3d clusters available: ", clusters)

	// Grab kubeconfig file for cluster.
	return k.getKubeconfig()
}

// GetKubeconfig returns the path of the kubeconfig file
// associated with this k3d cluster
func (k *Cluster) GetKubeconfig() string {
	return k.kubecfgFile
}

func (k *Cluster) GetKubeCtlContext() string {
	return fmt.Sprintf("k3d-%s", k.name)
}

func (k *Cluster) Destroy() error {
	log.V(4).Info("Destroying k3d cluster ", k.name)
	if err := k.findOrInstallK3d(k.e); err != nil {
		return err
	}

	p := k.e.RunProc(fmt.Sprintf(`k3d cluster delete %s`, k.name))
	if p.Err() != nil {
		return fmt.Errorf("k3d: delete cluster failed: %s: %s", p.Err(), p.Result())
	}

	log.V(4).Info("Removing kubeconfig file ", k.kubecfgFile)
	if err := os.RemoveAll(k.kubecfgFile); err != nil {
		return fmt.Errorf("k3d: remove kubefconfig failed: %w", err)
	}

	return nil
}

func (k *Cluster) findOrInstallK3d(e *gexe.Echo) error {
	if e.Prog().Avail("k3d") == "" {
		log.V(4).Infof(`k3d not found, installing with go install go install github.com/k3d-io/k3d@%s`, k3dVersion)
		if err := k.installK3d(e); err != nil {
			return err
		}
	}
	return nil
}

func (k *Cluster) installK3d(e *gexe.Echo) error {
	if k.version != "" {
		k3dVersion = k.version
	}

	var err error

	// first try to install with "brew"
	if _, found := commandExists("brew"); found {
		err = k.installK3dWithBrew(e)
	}
	if err == nil {
		if _, found := commandExists("k3d"); found {
			return nil
		}
	}

	// try to install with "curl"
	if _, found := commandExists("curl"); found {
		err = k.installK3dWithCurl(e)
	}
	if err == nil {
		if _, found := commandExists("k3d"); found {
			return nil
		}
	}

	// try to install with "go"
	if _, found := commandExists("go"); found {
		err = k.installK3dWithGo(e)
	}
	if err == nil {
		if _, found := commandExists("k3d"); found {
			return nil
		}
	}

	return fmt.Errorf("k3d not available even after installation")
}

func (k *Cluster) installK3dWithBrew(e *gexe.Echo) error {
	log.V(4).Infof("Installing with brew")
	p := e.RunProc("brew install k3d")
	if p.Err() != nil {
		return fmt.Errorf("failed to install k3d with brew: %s", p.Err())
	}

	if !p.IsSuccess() || p.ExitCode() != 0 {
		return fmt.Errorf("failed to install k3d with brew: %s", p.Result())
	}
	return nil
}

func (k *Cluster) installK3dWithCurl(e *gexe.Echo) error {
	log.V(4).Infof("Installing with curl")
	p := e.RunProc("curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash")
	if p.Err() != nil {
		return fmt.Errorf("failed to install k3d with curl: %s", p.Err())
	}

	if !p.IsSuccess() || p.ExitCode() != 0 {
		return fmt.Errorf("failed to install k3d with curl: %s", p.Result())
	}
	return nil
}

func (k *Cluster) installK3dWithGo(e *gexe.Echo) error {
	if k.version != "" {
		k3dVersion = k.version
	}

	log.V(4).Infof("Installing: go install github.com/k3d-io/k3d@%s", k3dVersion)
	p := e.RunProc(fmt.Sprintf("go install github.com/k3d-io/k3d@%s", k3dVersion))
	if p.Err() != nil {
		return fmt.Errorf("failed to install k3d: %s", p.Err())
	}

	if !p.IsSuccess() || p.ExitCode() != 0 {
		return fmt.Errorf("failed to install k3d: %s", p.Result())
	}

	// PATH may already be set to include $GOPATH/bin so we don't need to.
	if k3dPath := e.Prog().Avail("k3d"); k3dPath != "" {
		log.V(4).Info("Installed k3d at", k3dPath)
		return nil
	}

	p = e.RunProc("ls $GOPATH/bin")
	if p.Err() != nil {
		return fmt.Errorf("failed to install k3d: %s", p.Err())
	}

	p = e.RunProc("echo $PATH:$GOPATH/bin")
	if p.Err() != nil {
		return fmt.Errorf("failed to install k3d: %s", p.Err())
	}

	log.V(4).Info(`Setting path to include $GOPATH/bin:`, p.Result())
	e.SetEnv("PATH", p.Result())

	return nil
}

// LoadDockerImage loads a docker image from the host into the k3d cluster
func (k *Cluster) LoadDockerImage(image string) error {
	log.V(4).Infof("Loading image %q in %q", image, k.name)
	p := k.e.RunProc(fmt.Sprintf(`k3d image import --cluster %s %s`, k.name, image))
	if p.Err() != nil {
		return fmt.Errorf("k3d: load docker image failed: %s: %s", p.Err(), p.Result())
	}
	log.V(4).Info("... image loaded")
	return nil
}

// LoadImageArchive loads a docker image TAR archive from the host into the k3d cluster
func (k *Cluster) LoadImageArchive(imageArchive string) error {
	p := k.e.RunProc(fmt.Sprintf(`k3d image import --cluster %s %s`, k.name, imageArchive))
	if p.Err() != nil {
		return fmt.Errorf("k3d: load image archive failed: %s: %s", p.Err(), p.Result())
	}
	return nil
}

func commandExists(cmd string) (string, bool) {
	path, err := exec.LookPath(cmd)
	return path, err == nil
}
