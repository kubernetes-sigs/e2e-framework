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

package conf

import (
	"regexp"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/e2e-framework/pkg/flags"
)

type Filter struct {
	Assessment string
	Feature    string
	Labels     map[string]string
}

type Config struct {
	// kube
	namespace string
	kubecfg   *rest.Config

	assessmentRegex *regexp.Regexp
	featureRegex    *regexp.Regexp
	labels          map[string]string
}

func New() *Config {
	return &Config{}
}

// NewFromFlags creates a Config with parsed
// flag values pre-populated.
func NewFromFlags() (*Config, error) {
	newFlags, err := flags.Parse()
	if err != nil {
		return nil, err
	}

	cfg := New()
	if newFlags.Assessment() != "" {
		cfg.assessmentRegex = regexp.MustCompile(newFlags.Assessment())
	}
	if newFlags.Feature() != "" {
		cfg.featureRegex = regexp.MustCompile(newFlags.Feature())
	}
	cfg.labels = newFlags.Labels()

	return cfg, nil
}

// NewWithKubeCfgFile is a convenience constructor that will
// create a Kubernetes *rest.Config from a file
func NewWithKubeCfgFile(filePath string) (*Config, error) {
	return nil, nil
}

func (c *Config) WithKubeConfig(cfg *rest.Config) *Config {
	c.kubecfg = cfg
	return c
}

func (c *Config) KubeConfig() *rest.Config {
	return c.kubecfg
}

func (c *Config) WithNamespace(ns string) *Config {
	c.namespace = ns
	return c
}

func (c *Config) Namespace() string {
	return c.namespace
}

func (c *Config) WithAssessmentRegex(regex string) *Config {
	c.assessmentRegex = regexp.MustCompile(regex)
	return c
}

func (c *Config) AssessmentRegex() *regexp.Regexp {
	return c.assessmentRegex
}

func (c *Config) WithFeatureRegex(regex string) *Config {
	c.featureRegex = regexp.MustCompile(regex)
	return c
}

func (c *Config) FeatureRegex() *regexp.Regexp {
	return c.featureRegex
}

func (c *Config) Labels() map[string]string {
	return c.labels
}
