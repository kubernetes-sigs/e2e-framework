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

package e2e

import (
	// If testing with a cloud vendor managed cluster uncomment one of the below dependencies to properly get authorised.
	//_ "k8s.io/client-go/plugin/pkg/client/auth/azure" // auth for AKS clusters
	//_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"   // auth for GKE clusters
	//_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"  // auth for OIDC
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var testenv env.Environment

func TestMain(m *testing.M) {
	testenv = env.New()
	namespace := envconf.RandomName("sample-ns", 16)
	if os.Getenv("REAL_CLUSTER") == "true" {
		path := conf.ResolveKubeConfigFile()
		cfg := envconf.NewWithKubeConfig(path)
		testenv = env.NewWithConfig(cfg)

		testenv.Setup(
			envfuncs.CreateNamespace(namespace),
		)
		testenv.Finish(
			envfuncs.DeleteNamespace(namespace),
		)
	} else {
		kindClusterName := envconf.RandomName("kind-with-config", 16)

		testenv.Setup(
			envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
			envfuncs.CreateNamespace(namespace),
		)

		testenv.Finish(
			envfuncs.DeleteNamespace(namespace),
			envfuncs.DestroyCluster(kindClusterName),
		)
	}

	os.Exit(testenv.Run(m))
}
