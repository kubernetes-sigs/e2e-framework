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
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"
)

var kubeconfig string

func TestResolveKubeConfigFile(t *testing.T) {
	home := homedir.HomeDir()
	filename := ResolveKubeConfigFile()

	if filename != filepath.Join(home, "test", ".kube", "config") {
		t.Errorf("unexpected config path: %s", filename)
	}
}

func TestNew(t *testing.T) {
	cfg, err := New(ResolveKubeConfigFile())
	if err != nil {
		t.Error("error while creating client connection", err)
	}

	if cfg == nil {
		t.Errorf("client config is nil")
	}
}

func TestNewWithContextName(t *testing.T) {
	cfg, err := NewWithContextName(ResolveKubeConfigFile(), DefaultClusterContext)
	if err != nil {
		t.Error("error while client connection", err)
	}

	if cfg == nil {
		t.Errorf("client config is nill")
	}
}
