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

package flux

import (
	"fmt"
	"strings"

	"github.com/vladimirvivien/gexe"
	log "k8s.io/klog/v2"
)

type Opts struct {
	name      string
	source    string
	namespace string
	mode      string
	url       string
	branch    string
	tag       string
	commit    string
	path      string
	interval  string
	args      []string
}

type Source string

const (
	Git    Source = "git"
	Bucket Source = "bucket"
	Helm   Source = "helm"
	Oci    Source = "oci"
)

type Manager struct {
	e          *gexe.Echo
	kubeConfig string
}

type Option func(*Opts)

func (m *Manager) processOpts(opts ...Option) *Opts {
	option := &Opts{}
	for _, op := range opts {
		op(option)
	}
	return option
}

// WithNamespace is used to specify the namespace of flux installation
func WithNamespace(namespace string) Option {
	return func(opts *Opts) {
		opts.namespace = namespace
	}
}

// WithCommit is used to target a source with a specific commit SHA
func WithCommit(commit string) Option {
	return func(opts *Opts) {
		opts.commit = commit
	}
}

// WithTag is used to target a source with a specific tag
func WithTag(tag string) Option {
	return func(opts *Opts) {
		opts.tag = tag
	}
}

// WithBranch is used to target a source with a specific branch
func WithBranch(branch string) Option {
	return func(opts *Opts) {
		opts.branch = branch
	}
}

// WithPath is used to specify a path for reconciliation
func WithPath(path string) Option {
	return func(opts *Opts) {
		opts.path = path
	}
}

// WithInterval is used to specify how often flux should check for changes in a source
func WithInterval(interval string) Option {
	return func(opts *Opts) {
		opts.interval = interval
	}
}

// WithArgs is used to pass any additional parameter to Flux command
func WithArgs(args ...string) Option {
	return func(opts *Opts) {
		opts.args = args
	}
}

func (m *Manager) run(opts *Opts) (err error) {
	if m.e.Prog().Avail("flux") == "" {
		err = fmt.Errorf("'flux' command is missing. Please ensure the tool exists before using the flux manager")
		return
	}
	command := m.getCommand(opts)

	log.V(4).InfoS("Running Flux Operation", "command", command)
	proc := m.e.RunProc(command)
	result := proc.Result()
	log.V(4).Info("Flux Command output \n", result)
	if proc.IsSuccess() {
		return nil
	} else {
		return proc.Err()
	}
}

func New(kubeConfig string) *Manager {
	return &Manager{e: gexe.New(), kubeConfig: kubeConfig}
}

func (m *Manager) getCommand(opt *Opts) string {
	commandParts := []string{"flux", opt.mode}

	if opt.name != "" {
		commandParts = append(commandParts, opt.name)
	}
	if opt.source != "" {
		commandParts = append(commandParts, "--source", opt.source)
	}
	if opt.url != "" {
		commandParts = append(commandParts, "--url", opt.url)
	}
	if opt.namespace != "" {
		commandParts = append(commandParts, "--namespace", opt.namespace)
	}
	if opt.branch != "" {
		commandParts = append(commandParts, "--branch", opt.branch)
	}
	if opt.tag != "" {
		commandParts = append(commandParts, "--tag", opt.tag)
	}
	if opt.commit != "" {
		commandParts = append(commandParts, "--commit", opt.commit)
	}
	if opt.path != "" {
		commandParts = append(commandParts, "--path", opt.path)
	}
	if opt.interval != "" {
		commandParts = append(commandParts, "--interval", opt.interval)
	}

	commandParts = append(commandParts, opt.args...)
	commandParts = append(commandParts, "--kubeconfig", m.kubeConfig)
	return strings.Join(commandParts, " ")
}

func (m *Manager) installFlux(opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "install"
	return m.run(o)
}

func (m *Manager) uninstallFlux(opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "uninstall -s"
	return m.run(o)
}

func (m *Manager) createSource(sourceType Source, name, url string, opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = string("create source " + sourceType)
	o.name = name
	o.url = url
	return m.run(o)
}

func (m *Manager) deleteSource(sourceType Source, name string, opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = string("delete source " + sourceType + " -s")
	o.name = name
	return m.run(o)
}

func (m *Manager) createKustomization(name, source string, opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "create ks"
	o.name = name
	o.source = source
	return m.run(o)
}

func (m *Manager) deleteKustomization(name string, opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "delete ks -s"
	o.name = name
	return m.run(o)
}
