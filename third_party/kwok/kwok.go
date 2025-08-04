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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/client-go/rest"
	klog "k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/utils"
	"sigs.k8s.io/e2e-framework/support"
)

var kwokVersion = "v0.5.0"

type Cluster struct {
	name         string
	path         string
	kubecfgFile  string
	version      string
	waitDuration time.Duration
	rc           *rest.Config
}

var _ support.E2EClusterProvider = &Cluster{}

func NewCluster(name string) *Cluster {
	return &Cluster{name: name, waitDuration: 1 * time.Minute}
}

func NewProvider() support.E2EClusterProvider {
	return &Cluster{}
}

func WithPath(path string) support.ClusterOpts {
	return func(c support.E2EClusterProvider) {
		k, ok := c.(*Cluster)
		if ok {
			k.path = path
		}
	}
}

func WithWaitDuration(waitDuration time.Duration) support.ClusterOpts {
	return func(c support.E2EClusterProvider) {
		k, ok := c.(*Cluster)
		if ok {
			k.waitDuration = waitDuration
		}
	}
}

func (k *Cluster) findOrInstallKwokCtl() error {
	if k.version != "" {
		kwokVersion = k.version
	}
	path, err := utils.FindOrInstallGoBasedProvider(k.path, "kwokctl", "sigs.k8s.io/kwok/cmd/kwokctl", kwokVersion)
	if path != "" {
		k.path = path
	}
	return err
}

func (k *Cluster) clusterExists(name string) (string, bool) {
	clusters := utils.FetchCommandOutput(fmt.Sprintf("%s get clusters", k.path))
	for _, c := range strings.Split(clusters, "\n") {
		if c == name {
			return clusters, true
		}
	}
	return clusters, false
}

func (k *Cluster) getKubeconfig(args ...string) (string, error) {
	kubecfg := fmt.Sprintf("%s-kubecfg", k.name)

	var stdout, stderr bytes.Buffer
	cmd := fmt.Sprintf(`%s get kubeconfig %s --name %s`, k.path, strings.Join(args, " "), k.name)
	err := utils.RunCommandWithSeperatedOutput(cmd, &stdout, &stderr)
	if err != nil {
		return "", fmt.Errorf("kwokctl get kubeconfig: stderr: %s: %w", stderr.String(), err)
	}

	file, err := os.CreateTemp("", fmt.Sprintf("kwok-cluster-%s", kubecfg))
	if err != nil {
		return "", fmt.Errorf("kwok kubeconfig file: %w", err)
	}
	defer file.Close()

	k.kubecfgFile = file.Name()

	if n, err := io.WriteString(file, stdout.String()); n == 0 || err != nil {
		return "", fmt.Errorf("kwok kubecfg file: bytes copied: %d: %w]", n, err)
	}

	return file.Name(), nil
}

func (k *Cluster) initKubernetesAccessClients() error {
	cfg, err := conf.New(k.kubecfgFile)
	if err != nil {
		return err
	}
	k.rc = cfg
	return nil
}

func (k *Cluster) Create(ctx context.Context, args ...string) (string, error) {
	klog.V(4).Info("Creating a kwok cluster ", k.name)
	if err := k.findOrInstallKwokCtl(); err != nil {
		return "", err
	}
	if _, ok := k.clusterExists(k.name); ok {
		klog.V(4).Info("Skipping Kwok Cluster creation. Cluster already created ", k.name)
		kConfig, err := k.getKubeconfig()
		if err != nil {
			return "", err
		}
		return kConfig, k.initKubernetesAccessClients()
	}

	command := fmt.Sprintf(`%s create cluster --name %s --wait %s`, k.path, k.name, k.waitDuration.String())
	if len(args) > 0 {
		command = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	}
	klog.V(4).Info("Launching:", command)
	p := utils.RunCommand(command)
	if p.Err() != nil {
		outBytes, err := io.ReadAll(p.Out())
		if err != nil {
			klog.ErrorS(err, "failed to read data from the kwok create process output due to an error")
		}
		return "", fmt.Errorf("kwok: failed to create cluster %q: %s: %s: %s", k.name, p.Err(), p.Result(), string(outBytes))
	}

	clusters, ok := k.clusterExists(k.name)
	if !ok {
		return "", fmt.Errorf("kwok Cluster.Create: cluster %v still not in 'cluster list' after creation: %v", k.name, clusters)
	}
	klog.V(4).Info("kwok cluster available: ", clusters)

	kConfig, err := k.getKubeconfig()
	if err != nil {
		return "", err
	}
	return kConfig, k.initKubernetesAccessClients()
}

func (k *Cluster) CreateWithConfig(ctx context.Context, configFile string) (string, error) {
	if configFile == "" {
		return k.Create(ctx)
	}

	return k.Create(ctx, "--config", configFile)
}

func (k *Cluster) Destroy(ctx context.Context) error {
	klog.V(4).Info("Destroying kwok cluster ", k.name)
	if err := k.findOrInstallKwokCtl(); err != nil {
		return err
	}

	p := utils.RunCommand(fmt.Sprintf(`%s delete cluster --name %s`, k.path, k.name))
	if p.Err() != nil {
		outBytes, err := io.ReadAll(p.Out())
		if err != nil {
			klog.ErrorS(err, "failed to read data from the kwok delete process output due to an error")
		}
		return fmt.Errorf("kwok: failed to delete cluster %q: %s: %s: %s", k.name, p.Err(), p.Result(), string(outBytes))
	}

	klog.V(4).Info("Removing kubeconfig file ", k.kubecfgFile)
	if err := os.RemoveAll(k.kubecfgFile); err != nil {
		return fmt.Errorf("kwok: remove kubefconfig failed: %w", err)
	}
	return nil
}

func (k *Cluster) ExportLogs(ctx context.Context, dest string) error {
	if err := k.findOrInstallKwokCtl(); err != nil {
		return err
	}
	// In kwokctl 0.3.0 and above, there is a new kwokctl export logs feature that has been added which can
	// simplify the workf of exporting the logs for us. Let us check if the CLI has that command and if so
	// let us use that to export logs. Otherwise, we can fallback to exporting individual items.
	p := utils.RunCommand(fmt.Sprintf("%s export logs --help", k.path))
	if p.ExitCode() == 0 {
		return utils.RunCommand(fmt.Sprintf("%s --name %s export logs %s", k.path, k.name, dest)).Err()
	}

	// TODO: Get Rid of this if we decide to enforce a min version of the kwokctl at some point
	for _, component := range []string{"audit", "etcd", "kube-apiserver", "kube-controller-manager", "kube-scheduler", "kwok-controller", "prometheus"} {
		command := fmt.Sprintf("%s logs %s", k.path, component)
		p := utils.RunCommand(command)
		if p.Err() != nil {
			klog.ErrorS(p.Err(), "ran into an error trying to export the log", "component", component)
			continue
		}
		var stdout bytes.Buffer
		if _, err := stdout.ReadFrom(p.Out()); err != nil {
			return fmt.Errorf("kwokctl logs %s stdout bytes: %w", component, err)
		}
		file, err := os.Create(filepath.Join(dest, fmt.Sprintf("%s.log", component)))
		if err != nil {
			klog.ErrorS(err, "ran into an error trying to create file to export logs", "component", component)
			continue
		}
		if n, err := io.Copy(file, &stdout); n == 0 || err != nil {
			klog.ErrorS(err, "kwokctl logs %s file: bytes copied: %d: %w]", component, n, err)
		}
	}
	return nil
}

func (k *Cluster) GetKubectlContext() string {
	return fmt.Sprintf("kwok-%s", k.name)
}

func (k *Cluster) GetKubeconfig() string {
	return k.kubecfgFile
}

func (k *Cluster) SetDefaults() support.E2EClusterProvider {
	if k.path == "" {
		k.path = "kwokctl"
	}
	return k
}

func (k *Cluster) WaitForControlPlane(ctx context.Context, client klient.Client) error {
	klog.V(4).Info("kwokctl doesn't implement a WaitForControlPlane handler. The --wait argument passed to the `kwokctl` should take care of this already")
	return nil
}

func (k *Cluster) WithName(name string) support.E2EClusterProvider {
	k.name = name
	return k
}

func (k *Cluster) WithOpts(opts ...support.ClusterOpts) support.E2EClusterProvider {
	for _, o := range opts {
		o(k)
	}
	return k
}

func (k *Cluster) WithPath(path string) support.E2EClusterProvider {
	k.path = path
	return k
}

func (k *Cluster) WithVersion(version string) support.E2EClusterProvider {
	k.version = version
	return k
}

func (k *Cluster) KubernetesRestConfig() *rest.Config {
	return k.rc
}

func (k *Cluster) GenerateKubeconfig(args ...string) (string, error) {
	return k.getKubeconfig(args...)
}
