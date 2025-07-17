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

package kubeconfigenv

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var testEnv env.Environment

func TestMain(m *testing.M) {
	// skips flag parsing
	testEnv = env.New()

	os.Exit(testEnv.Run(m))
}

func TestKubeconfig(t *testing.T) {
	testEnv.Test(t,
		features.New("Kubeconfig").
			Assess("Read Kubeconfig", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				assert.NoError(t, os.Setenv("KUBECONFIG", ".kube/config:/root/.kube/config"))
				assert.Equal(t, ".kube/config", conf.ResolveKubeConfigFile())
				return ctx
			}).
			Feature(),
	)
}
