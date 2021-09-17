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

package envconf

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/flags"
)

// Config represents and environment configuration
type Config struct {
	client          klient.Client
	namespace       string
	assessmentRegex *regexp.Regexp
	featureRegex    *regexp.Regexp
	labels          map[string]string
}

// New creates and initializes an empty environment configuration
func New() *Config {
	return &Config{}
}

// NewWithKubeconfig is a convenience constructor function
// that creates a new environment configuration using a kubeconfig file.
func NewWithKubeconfig(kubecfg string) (*Config, error) {
	client, err := klient.NewWithKubeConfigFile(kubecfg)
	if err != nil {
		return nil, err
	}
	cfg := New()
	cfg.client = client
	return cfg, nil
}

// NewFromFlags initializes an environment config using values
// from command-line argument flags. See package flags for supported flags.
func NewFromFlags() (*Config, error) {
	flagset, err := flags.Parse()
	if err != nil {
		return nil, err
	}
	e := New()
	if flagset.Assessment() != "" {
		e.assessmentRegex = regexp.MustCompile(flagset.Assessment())
	}
	if flagset.Feature() != "" {
		e.featureRegex = regexp.MustCompile(flagset.Feature())
	}

	// setup EnvConfig
	e.labels = flagset.Labels()
	e.namespace = flagset.Namespace()

	kubecfg := flagset.Kubeconfig()
	if kubecfg == "" {
		kubecfg = conf.ResolveKubeConfigFile()
	}

	// if there is a kubeconfig
	if kubecfg != "" {
		c, err := klient.NewWithKubeConfigFile(kubecfg)
		if err != nil {
			return nil, err
		}
		e.client = c
	}

	return e, nil
}

// WithKubeconfigFile creates a new klient.Client and injects it in the cfg
func (c *Config) WithKubeconfigFile(kubecfg string) (*Config, error) {
	client, err := klient.NewWithKubeConfigFile(kubecfg)
	if err != nil {
		return nil, err
	}
	c.client = client
	return c, nil
}

// WithClient used to update the environment klient.Client
func (c *Config) WithClient(client klient.Client) *Config {
	c.client = client
	return c
}

// Client returns the environment klient.Client
func (c *Config) Client() klient.Client {
	return c.client
}

// WithNamespace updates the environment namespace value
func (c *Config) WithNamespace(ns string) *Config {
	c.namespace = ns
	return c
}

// WithRandomNamespace sets the environment's namespace
// to a random value
func (c *Config) WithRandomNamespace() *Config {
	c.namespace = randNS()
	return c
}

// Namespace returns the namespace for the environment
func (c *Config) Namespace() string {
	return c.namespace
}

// WithAssessmentRegex sets the environment assessment regex filter
func (c *Config) WithAssessmentRegex(regex string) *Config {
	c.assessmentRegex = regexp.MustCompile(regex)
	return c
}

// AssessmentRegex returns the environment assessment filter
func (c *Config) AssessmentRegex() *regexp.Regexp {
	return c.assessmentRegex
}

// WithFeatureRegex sets the environment's feature regex filter
func (c *Config) WithFeatureRegex(regex string) *Config {
	c.featureRegex = regexp.MustCompile(regex)
	return c
}

// FeatureRegex returns the environment's feature regex filter
func (c *Config) FeatureRegex() *regexp.Regexp {
	return c.featureRegex
}

// WithLabels sets the environment label filters
func (c *Config) WithLabels(lbls map[string]string) *Config {
	c.labels = lbls
	return c
}

// Labels returns the environment's label filters
func (c *Config) Labels() map[string]string {
	return c.labels
}

func randNS() string {
	return RandomName("testns-", 32)
}

// RandomName generates a random name of n length with the provided
// prefix. If prefix is omitted, the then entire name is random char.
func RandomName(prefix string, n int) string {
	if n == 0 {
		n = 32
	}
	if len(prefix) >= n {
		return prefix
	}
	rand.Seed(time.Now().UnixNano())
	p := make([]byte, n)
	rand.Read(p)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(p))[:n]
}
