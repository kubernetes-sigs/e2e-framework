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
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"k8s.io/client-go/util/homedir"
)

var kubeconfig string

func TestResolveKubeConfigFileFlag(t *testing.T) {
	filename := ResolveKubeConfigFile()
	if filename != kubeconfigpath {
		t.Errorf("unexpected config path: %s", filename)
	}
}

func TestResolveKubeConfigFileEnv(t *testing.T) {
	// NOTE: not considered safe to run in parallel with other tests thats
	// require the --kubeconfig and --kubecontext flags.
	clearKubeconfigFlags()
	defer setKubeconfigFlags()

	kubeConfigPath1 := filepath.Join(t.TempDir(), "config")
	if _, err := os.Create(kubeConfigPath1); err != nil {
		t.Errorf("failed to create kubeconfig: %v", err)
	}

	kubeConfigPath2 := filepath.Join(t.TempDir(), "config")
	if _, err := os.Create(kubeConfigPath2); err != nil {
		t.Errorf("failed to create kubeconfig: %v", err)
	}

	t.Run("WithEnvEmpty", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "")

		filename := ResolveKubeConfigFile()

		// this will fallback to the true home directory.
		if filename != filepath.Join(homedir.HomeDir(), ".kube", "config") {
			t.Errorf("unexpected config path: %s", filename)
		}
	})

	t.Run("WithEnvPath", func(t *testing.T) {
		t.Setenv("KUBECONFIG", kubeConfigPath1)

		filename := ResolveKubeConfigFile()

		if filename != kubeConfigPath1 {
			t.Errorf("unexpected config path: %s", filename)
		}
	})

	t.Run("WithEnvPathListAllExist", func(t *testing.T) {
		t.Setenv("KUBECONFIG", fmt.Sprintf("%s:%s", kubeConfigPath1, kubeConfigPath2))

		filename := ResolveKubeConfigFile()

		// if all exist then it will take the first.
		if filename != kubeConfigPath1 {
			t.Errorf("unexpected config path: %s", filename)
		}
	})

	t.Run("WithEnvPathListFirstExists", func(t *testing.T) {
		t.Setenv("KUBECONFIG", fmt.Sprintf("%s:fake", kubeConfigPath1))

		filename := ResolveKubeConfigFile()

		// if first exists then it will take the first.
		if filename != kubeConfigPath1 {
			t.Errorf("unexpected config path: %s", filename)
		}
	})

	t.Run("WithEnvPathListLastExists", func(t *testing.T) {
		t.Setenv("KUBECONFIG", fmt.Sprintf("%s:fake", kubeConfigPath1))

		filename := ResolveKubeConfigFile()

		// if only last exists then it will take the last.
		if filename != kubeConfigPath1 {
			t.Errorf("unexpected config path: %s", filename)
		}
	})

	t.Run("WithEnvPathListNoneExist", func(t *testing.T) {
		t.Setenv("KUBECONFIG", "fake-foo:fake-bar")

		filename := ResolveKubeConfigFile()

		// if none exist then it will take the last.
		if filename != "fake-bar" {
			t.Errorf("unexpected config path: %s", filename)
		}
	})
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
