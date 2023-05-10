package flux

import (
	"fmt"
	"github.com/vladimirvivien/gexe"
	log "k8s.io/klog/v2"
	"strings"
)

type Opts struct {
	name      string
	source    string
	namespace string
	mode      string
	url       string
	branch    string
	tag       string
	path      string
	interval  string
	args      []string
}

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

func (m *Manager) InstallFlux(opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "install"
	return m.run(o)
}

func (m *Manager) UninstallFlux(opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "uninstall -s"
	return m.run(o)
}

func WithNamespace(namespace string) Option {
	return func(opts *Opts) {
		opts.namespace = namespace
	}
}

func WithTag(tag string) Option {
	return func(opts *Opts) {
		opts.tag = tag
	}
}

func WithBranch(branch string) Option {
	return func(opts *Opts) {
		opts.branch = branch
	}
}

func WithPath(path string) Option {
	return func(opts *Opts) {
		opts.path = path
	}
}

func WithInterval(interval string) Option {
	return func(opts *Opts) {
		opts.interval = interval
	}
}

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
	command, err := m.getCommand(opts)
	if err != nil {
		return
	}
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

func (m *Manager) getCommand(opt *Opts) (string, error) {
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
	if opt.path != "" {
		commandParts = append(commandParts, "--path", opt.path)
	}
	if opt.interval != "" {
		commandParts = append(commandParts, "--interval", opt.interval)
	}

	commandParts = append(commandParts, opt.args...)
	commandParts = append(commandParts, "--kubeconfig", m.kubeConfig)
	return strings.Join(commandParts, " "), nil
}

func (m *Manager) CreateGitRepo(name string, url string, opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "create source git"
	o.name = name
	o.url = url
	return m.run(o)
}

func (m *Manager) DeleteGitRepo(name string, opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "delete source git -s"
	o.name = name
	return m.run(o)
}

func (m *Manager) CreateKustomization(name string, source string, opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "create ks"
	o.name = name
	o.source = source
	return m.run(o)
}

func (m *Manager) DeleteKustomization(name string, opts ...Option) error {
	o := m.processOpts(opts...)
	o.mode = "delete ks -s"
	o.name = name
	return m.run(o)
}
