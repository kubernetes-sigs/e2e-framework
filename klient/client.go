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

package klient

import (
	"k8s.io/client-go/rest"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

// Client stores values to interact with the
// API-server.
type Client interface {
	RESTConfig() *rest.Config
	Resources() *resources.Resources
}

type client struct {
	cfg       *rest.Config
	resources *resources.Resources
}

// New returns a new Client value
func New(cfg *rest.Config) (Client, error) {
	res, err := resources.New(cfg)
	if err != nil {
		return nil, err
	}
	return &client{cfg: cfg, resources: res}, nil
}

// NewWithKubeConfigFile creates a client using the kubeconfig filePath
func NewWithKubeConfigFile(filePath string) (Client, error) {
	cfg, err := conf.New(filePath)
	if err != nil {
		return nil, err
	}
	return New(cfg)
}

func (c *client) RESTConfig() *rest.Config {
	return c.cfg
}

func (c *client) Resources() *resources.Resources {
	return c.resources
}
