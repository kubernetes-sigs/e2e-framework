/*
Copyright 2024 The Kubernetes Authors.

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
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"k8s.io/apimachinery/pkg/util/json"

	"k8s.io/client-go/rest"

	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/utils"
	"sigs.k8s.io/e2e-framework/support"

	log "k8s.io/klog/v2"
)

var k3dVersion = "v5.7.2"

type Cluster struct {
	path           string
	name           string
	kubeConfigFile string
	version        string
	image          string
	rc             *rest.Config
	args           []string
}

// k3dNode is a struct containing a subset of values that are part of the k3d node list -o json
// data. This currently contains only those fields that are of interest to generate the
// support.Node struct in return for performing node operations.
type k3dNode struct {
	Name string `json:"name"`
	Role string `json:"role"`
	IP   struct {
		IP string `json:"IP"`
	} `json:"IP"`
	State struct {
		Running bool   `json:"Running"`
		Status  string `json:"Status"`
	} `json:"State"`
}

var (
	_ support.E2EClusterProviderWithImageLoader = &Cluster{}
	_ support.E2EClusterProviderWithLifeCycle   = &Cluster{}
)

func WithArgs(args ...string) support.ClusterOpts {
	return func(c support.E2EClusterProvider) {
		k, ok := c.(*Cluster)
		if ok {
			k.args = append(k.args, args...)
		}
	}
}

func WithImage(image string) support.ClusterOpts {
	return func(c support.E2EClusterProvider) {
		k, ok := c.(*Cluster)
		if ok {
			k.image = image
		}
	}
}

func NewCluster(name string) *Cluster {
	return &Cluster{name: name}
}

func NewProvider() support.E2EClusterProvider {
	return &Cluster{}
}

func (c *Cluster) findOrInstallK3D() error {
	if c.version != "" {
		k3dVersion = c.version
	}
	path, err := utils.FindOrInstallGoBasedProvider(c.path, "k3d", "github.com/k3d-io/k3d/v5", k3dVersion)
	if path != "" {
		c.path = path
	}
	return err
}

func (c *Cluster) getKubeConfig() (string, error) {
	kubeCfg := fmt.Sprintf("%s-kubecfg", c.name)

	var stdout, stderr bytes.Buffer
	err := utils.RunCommandWithSeperatedOutput(fmt.Sprintf("%s kubeconfig get %s", c.path, c.name), &stdout, &stderr)
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %s", stderr.String())
	}
	log.V(4).Info("k3d kubeconfig get stderr \n", stderr.String())

	file, err := os.CreateTemp("", fmt.Sprintf("k3d-cluster-%s", kubeCfg))
	if err != nil {
		return "", fmt.Errorf("k3d kubeconfig file: %w", err)
	}
	defer file.Close()

	c.kubeConfigFile = file.Name()

	if n, err := io.WriteString(file, stdout.String()); n == 0 || err != nil {
		return "", fmt.Errorf("k3d kubeconfig file: bytes copied: %d: %w]", n, err)
	}

	return file.Name(), nil
}

func (c *Cluster) clusterExists(name string) (string, bool) {
	clusters := utils.FetchCommandOutput(fmt.Sprintf("%s cluster get --no-headers", c.path))
	for _, c := range strings.Split(clusters, "\n") {
		if strings.HasPrefix(c, name) {
			return clusters, true
		}
	}
	return clusters, false
}

func (c *Cluster) startCluster(name string) error {
	cmd := fmt.Sprintf("%s cluster start %s", c.path, name)
	log.V(4).InfoS("Starting k3d cluster", "command", cmd)
	p := utils.RunCommand(cmd)
	if p.Err() != nil {
		return fmt.Errorf("k3d: failed to start cluster %q: %s: %s", name, p.Err(), p.Result())
	}
	return nil
}

func (c *Cluster) initKubernetesAccessClients() error {
	cfg, err := conf.New(c.kubeConfigFile)
	if err != nil {
		return err
	}
	c.rc = cfg
	return nil
}

func (c *Cluster) WithName(name string) support.E2EClusterProvider {
	c.name = name
	return c
}

func (c *Cluster) WithVersion(version string) support.E2EClusterProvider {
	c.version = version
	return c
}

func (c *Cluster) WithPath(path string) support.E2EClusterProvider {
	c.path = path
	return c
}

func (c *Cluster) WithOpts(opts ...support.ClusterOpts) support.E2EClusterProvider {
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Cluster) Create(ctx context.Context, args ...string) (string, error) {
	log.V(4).InfoS("Creating k3d cluster", "name", c.name)
	if err := c.findOrInstallK3D(); err != nil {
		return "", fmt.Errorf("failed to find or install k3d: %w", err)
	}

	if _, ok := c.clusterExists(c.name); ok {
		// This is being done as an extra step to ensure that in case you have the cluster by the same name, but it is not up.
		// Starting an already started cluster won't cause any harm. So, we will just start it once before continuing
		// further down the line and process rest of the workflows
		if err := c.startCluster(c.name); err != nil {
			return "", err
		}
		log.V(4).InfoS("Skipping k3d cluster creation. Cluster already exists", "name", c.name)
		kConfig, err := c.getKubeConfig()
		if err != nil {
			return "", err
		}
		return kConfig, c.initKubernetesAccessClients()
	}

	if c.image != "" {
		args = append(args, "--image", c.image)
	}

	args = append(args, c.args...)
	cmd := fmt.Sprintf("%s cluster create %s", c.path, c.name)

	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	log.V(4).InfoS("Launching k3d cluster", "command", cmd)

	var stdout, stderr bytes.Buffer

	p := utils.RunCommandWithCustomWriter(cmd, &stdout, &stderr)
	if p.Err() != nil {
		return "", fmt.Errorf("k3d: failed to create cluster %q: %s: %s: %s %s", c.name, p.Err(), p.Result(), stdout.String(), stderr.String())
	}
	clusters, ok := c.clusterExists(c.name)
	if !ok {
		return "", fmt.Errorf("k3d cluster create: cluster %v still not in 'cluster list' after creation: %v", c.name, clusters)
	}
	log.V(4).Info("k3d clusters available: ", clusters)

	kConfig, err := c.getKubeConfig()
	if err != nil {
		return "", err
	}
	return kConfig, c.initKubernetesAccessClients()
}

func (c *Cluster) CreateWithConfig(ctx context.Context, configFile string) (string, error) {
	var args []string
	if configFile != "" {
		args = append(args, "--config", configFile)
	}
	return c.Create(ctx, args...)
}

func (c *Cluster) GetKubeconfig() string {
	return c.kubeConfigFile
}

func (c *Cluster) GetKubectlContext() string {
	return fmt.Sprintf("k3d-%s", c.name)
}

func (c *Cluster) ExportLogs(ctx context.Context, dest string) error {
	log.Warning("ExportLogs not implemented for k3d. Please use regular kubectl like commands to extract the logs from the cluster")
	return nil
}

func (c *Cluster) Destroy(ctx context.Context) error {
	log.V(4).InfoS("Destroying k3d cluster", "name", c.name)
	if err := c.findOrInstallK3D(); err != nil {
		return fmt.Errorf("failed to find or install k3d: %w", err)
	}

	if _, ok := c.clusterExists(c.name); !ok {
		log.V(4).InfoS("Skipping k3d cluster destruction. Cluster does not exist", "name", c.name)
		return nil
	}

	cmd := fmt.Sprintf("%s cluster delete %s", c.path, c.name)
	log.V(4).InfoS("Destroying k3d cluster", "command", cmd)
	p := utils.RunCommand(cmd)
	if p.Err() != nil {
		outBytes, err := io.ReadAll(p.Out())
		if err != nil {
			log.ErrorS(err, "failed to read data from the k3d cluster delete process output due to an error")
		}
		return fmt.Errorf("k3d: failed to delete cluster %q: %s: %s: %s", c.name, p.Err(), p.Result(), string(outBytes))
	}

	log.V(4).InfoS("Removing kubeconfig file", "configFile", c.kubeConfigFile)
	if err := os.RemoveAll(c.kubeConfigFile); err != nil {
		return fmt.Errorf("k3d: failed to remove kubeconfig file %q: %w", c.kubeConfigFile, err)
	}
	return nil
}

func (c *Cluster) SetDefaults() support.E2EClusterProvider {
	if c.path == "" {
		c.path = "k3d"
	}
	return c
}

func (c *Cluster) WaitForControlPlane(ctx context.Context, client klient.Client) error {
	log.V(4).Info(" k3d provider doesn't implement a WaitForControlPlane as the provider automatically wait for the control plane")
	return nil
}

func (c *Cluster) KubernetesRestConfig() *rest.Config {
	return c.rc
}

func (c *Cluster) LoadImage(ctx context.Context, image string, args ...string) error {
	log.V(4).InfoS("Performing Image load operation", "cluster", c.name, "image", image, "args", args)
	p := utils.RunCommand(fmt.Sprintf("%s image import --cluster %s %s %s", c.path, c.name, strings.Join(args, " "), image))
	if p.Err() != nil {
		return fmt.Errorf("k3d: load docker-image %v failed: %s: %s", image, p.Err(), p.Result())
	}
	return nil
}

func (c *Cluster) LoadImageArchive(ctx context.Context, imageArchive string, args ...string) error {
	return c.LoadImage(ctx, imageArchive, args...)
}

func (c *Cluster) AddNode(ctx context.Context, node *support.Node, args ...string) error {
	cmd := fmt.Sprintf("%s node create %s --cluster %s", c.path, node.Name, c.name)

	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	if node.Role != "" {
		cmd = fmt.Sprintf("%s --role %s", cmd, node.Role)
	}

	log.V(4).InfoS("Adding node to k3d cluster", "command", cmd)
	p, stdout, stderr := utils.FetchSeperatedCommandOutput(cmd)
	if p.Err() != nil || (p.Exited() && p.ExitCode() != 0) {
		log.ErrorS(p.Err(), "failed to add node to k3d cluster", "stdout", stdout.String(), "stderr", stderr.String())
		return fmt.Errorf("k3d: failed to add node %q to cluster %q: %s: %s", node.Name, c.name, p.Err(), p.Result())
	}
	return nil
}

func (c *Cluster) RemoveNode(ctx context.Context, node *support.Node, args ...string) error {
	cmd := fmt.Sprintf("%s node delete %s", c.path, node.Name)

	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}

	log.V(4).InfoS("Removing node from k3d cluster", "command", cmd)
	p, stdout, stderr := utils.FetchSeperatedCommandOutput(cmd)
	if p.Err() != nil || (p.Exited() && p.ExitCode() != 0) {
		log.ErrorS(p.Err(), "failed to remove node from k3d cluster", "stdout", stdout.String(), "stderr", stderr.String())
		return fmt.Errorf("k3d: failed to remove node %q from cluster %q: %s: %s", node.Name, c.name, p.Err(), p.Result())
	}
	return nil
}

func (c *Cluster) StartNode(ctx context.Context, node *support.Node, args ...string) error {
	cmd := fmt.Sprintf("%s node start %s", c.path, node.Name)
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	log.V(4).InfoS("Starting node in k3d cluster", "command", cmd)
	p, stdout, stderr := utils.FetchSeperatedCommandOutput(cmd)
	if p.Err() != nil || (p.Exited() && p.ExitCode() != 0) {
		log.ErrorS(p.Err(), "failed to start node in k3d cluster", "stdout", stdout.String(), "stderr", stderr.String())
		return fmt.Errorf("k3d: failed to start node %q in cluster %q: %s: %s", node.Name, c.name, p.Err(), p.Result())
	}
	return nil
}

func (c *Cluster) StopNode(ctx context.Context, node *support.Node, args ...string) error {
	cmd := fmt.Sprintf("%s node stop %s", c.path, node.Name)
	if len(args) > 0 {
		cmd = fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
	}
	log.V(4).InfoS("Stopping node in k3d cluster", "command", cmd)
	p, stdout, stderr := utils.FetchSeperatedCommandOutput(cmd)
	if p.Err() != nil || (p.Exited() && p.ExitCode() != 0) {
		log.ErrorS(p.Err(), "failed to stop node in k3d cluster", "stdout", stdout.String(), "stderr", stderr.String())
		return fmt.Errorf("k3d: failed to stop node %q in cluster %q: %s: %s", node.Name, c.name, p.Err(), p.Result())
	}
	return nil
}

func (c *Cluster) ListNode(ctx context.Context, args ...string) ([]support.Node, error) {
	cmd := fmt.Sprintf("%s node list -o json", c.path)
	p := utils.RunCommand(cmd)
	if p.Err() != nil || (p.Exited() && p.ExitCode() != 0) {
		return nil, fmt.Errorf("k3d: failed to list nodes: %s: %s", p.Err(), p.Result())
	}
	var nodeInfo []k3dNode
	if err := json.Unmarshal([]byte(p.Result()), &nodeInfo); err != nil {
		return nil, fmt.Errorf("k3d: failed to unmarshal node list: %s", err)
	}
	nodes := make([]support.Node, len(nodeInfo))
	for _, n := range nodeInfo {
		nodes = append(nodes, support.Node{
			Name:    n.Name,
			Role:    n.Role,
			IP:      net.ParseIP(n.IP.IP),
			State:   n.State.Status,
			Cluster: c.name,
		})
	}
	return nodes, nil
}
