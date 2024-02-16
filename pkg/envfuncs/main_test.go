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

package envfuncs_test

import (
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

// nsTestenv is used to test the ns_funcs
var nsTestenv env.Environment

func TestMain(m *testing.M) {
	nsClusterName := envconf.RandomName("ns-cluster", 16)
	nsTestenv = env.New()

	// Use the same cluster for all ns_funcs tests
	nsTestenv.
		Setup(envfuncs.CreateCluster(kind.NewProvider(), nsClusterName)).
		Finish(envfuncs.DestroyCluster(nsClusterName))

	os.Exit(nsTestenv.Run(m))
}
