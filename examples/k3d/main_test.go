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
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/k3d"
)

var (
	testEnv     env.Environment
	clusterName string
)

func TestMain(m *testing.M) {
	testEnv = env.New()
	clusterName = envconf.RandomName("test", 16)
	namespace := envconf.RandomName("k3d-ns", 16)

	testEnv.Setup(
		envfuncs.CreateClusterWithOpts(k3d.NewProvider(), clusterName, k3d.WithImage("rancher/k3s:v1.29.6-k3s1")),
		envfuncs.CreateNamespace(namespace),
		envfuncs.LoadImageToCluster(clusterName, "rancher/k3s:v1.29.6-k3s1", "--verbose", "--mode", "direct"),
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyCluster(clusterName),
	)

	os.Exit(testEnv.Run(m))
}
