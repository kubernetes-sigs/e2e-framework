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

package ko

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/vladimirvivien/gexe"
	log "k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/pkg/utils"
)

type localImageContextKey string

type Opts struct {
	// LocalKindName is used to indicate the local kind cluster to publish.
	LocalKindName string
	// Platforms is used to indicate the platforms to build and publish.
	Platforms []string
	// ConfigFile is used to indicate the ko config file path.
	ConfigFile string

	baseCmdAndOpt string
}

type Manager struct {
	e    *gexe.Echo
	path string
}

type Option func(*Opts)

const (
	missingKo = "'ko' command is missing. Please ensure the tool exists before using the ko manager"
)

// WithConfigFile is used to configure the ko config file path.
func WithConfigFile(configFile string) Option {
	return func(opts *Opts) {
		opts.ConfigFile = configFile
	}
}

// WithPlatforms is used to configure the platform to use when pulling
// a multi-platform base at the building phase. When platform="all",
// it will build and push an image for all platforms supported by the
// configured base image.
// platform string format: all | <os>[/<arch>[/<variant>]][,platform]*
func WithPlatforms(platforms ...string) Option {
	return func(opts *Opts) {
		opts.Platforms = append(opts.Platforms, platforms...)
	}
}

// WithLocalKind is used to configure the build and publish target as
// a local kind cluster.
func WithLocalKindName(name string) Option {
	return func(opts *Opts) {
		opts.LocalKindName = name
	}
}

// processOpts is used to generate the Opts resource that will be used to generate
// the actual helm command to be run using the getCommand helper
func (m *Manager) processOpts(opts ...Option) *Opts {
	option := &Opts{}
	for _, op := range opts {
		op(option)
	}
	return option
}

// getCommand is used to convert the Opts into a ko suitable command to be run
func (m *Manager) getCommand(opt *Opts) string {
	commandParts := []string{m.path, opt.baseCmdAndOpt}

	if len(opt.Platforms) != 0 {
		commandParts = append(commandParts, "--platform", strings.Join(opt.Platforms, ","))
	}

	return strings.Join(commandParts, " ")
}

// getEnvs is used to convert the Opts into environment variable that ko need.
func (m *Manager) getEnvs(opt *Opts) map[string]string {
	envs := map[string]string{}

	if opt.ConfigFile != "" {
		envs["KO_CONFIG_PATH"] = opt.ConfigFile
	}

	if opt.LocalKindName != "" {
		envs["KO_DOCKER_REPO"] = "kind.local"
		envs["KIND_CLUSTER_NAME"] = opt.LocalKindName
	}

	return envs
}

// Install install ko with `go install` if ko not found in PATH
func (m *Manager) Install(version string) error {
	path, err := utils.FindOrInstallGoBasedProvider(m.path, "ko", "github.com/google/ko", version)
	if path != "" {
		m.path = path
	}

	return err
}

// BuildLocal builds container image from the given packagePath and publishes it to a
// local repository supported by ko. It returns the container image ID within the ctx.
func (m *Manager) BuildLocal(ctx context.Context, packagePath string, opts ...Option) (context.Context, error) {
	o := m.processOpts(opts...)
	o.baseCmdAndOpt = fmt.Sprintf("build %s", packagePath)

	image, err := m.run(o)
	if err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, localImageContextKey(packagePath), image), nil
}

// GetLocalImage returns the previously built container image ID for packagePath from ctx.
func (m *Manager) GetLocalImage(ctx context.Context, packagePath string) (string, error) {
	var image string

	imgVal := ctx.Value(localImageContextKey(packagePath))
	if imgVal == nil {
		return "", fmt.Errorf("container image not found for packagePath %s", packagePath)
	}

	if img, ok := imgVal.(string); ok {
		image = img
	}

	return image, nil
}

// run method is used to invoke a ko command to perform a suitable operation.
// Please make sure to configure the right Opts using the Option helpers
func (m *Manager) run(opts *Opts) (out string, err error) {
	log.V(4).InfoS("Determining if ko binary is available or not", "executable", m.path)
	if m.e.Prog().Avail(m.path) == "" {
		return "", errors.New(missingKo)
	}

	envs := m.getEnvs(opts)
	command := m.getCommand(opts)

	var envsString string
	for k, v := range envs {
		envsString += k + "=" + v + " "
		m.e = m.e.SetEnv(k, v)
	}
	log.V(4).InfoS("Running Ko Operation", "envs", envsString, "command", command)
	proc := m.e.NewProc(command)

	var stderr bytes.Buffer
	var stdout bytes.Buffer
	proc.SetStderr(&stderr)
	proc.SetStdout(&stdout)

	result := proc.Run().Result()
	log.V(4).Info("Ko Command output \n", result)

	if !proc.IsSuccess() {
		return "", fmt.Errorf("%s: %w", strings.TrimSuffix(stderr.String(), "\n"), proc.Err())
	}

	return strings.TrimSuffix(stdout.String(), "\n"), nil
}

// WithPath is used to provide a custom path where the `ko` executable command
// can be found. This is useful in case if your binary is in a non standard location
// and you want to framework to use that instead of returning an error.
func (m *Manager) WithPath(path string) *Manager {
	m.path = path
	return m
}

// New creates a ko Manager.
func New() *Manager {
	return &Manager{
		path: "ko",
		e:    gexe.New(),
	}
}
