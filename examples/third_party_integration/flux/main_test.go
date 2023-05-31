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
	"os"
	"sigs.k8s.io/e2e-framework/third_party/flux"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var (
	testEnv         env.Environment
	namespace       string
	kindClusterName string
)

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	kindClusterName = envconf.RandomName("flux", 10)
	namespace = envconf.RandomName("flux", 10)
	gitRepoName := "hello-world"
	ksName := "hello-world"
	testEnv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		envfuncs.CreateNamespace(namespace),
		flux.InstallFlux(),
		flux.CreateGitRepo(gitRepoName, "https://github.com/matrus2/go-hello-world", flux.WithBranch("main")),
		flux.CreateKustomization(ksName, "GitRepository/"+gitRepoName+".flux-system", flux.WithPath("template"), flux.WithArgs("--target-namespace", namespace, "--prune")),
	)

	testEnv.Finish(
		flux.DeleteKustomization(ksName),
		flux.DeleteGitRepo(gitRepoName),
		flux.UninstallFlux(),
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyKindCluster(kindClusterName),
	)
	os.Exit(testEnv.Run(m))
}
