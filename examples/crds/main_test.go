/*
Copyright 2022 The Kubernetes Authors.

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

package crds

import (
	"context"
	"os"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
)

var (
	testEnv         env.Environment
	kindClusterName string
	namespace       string
)

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	kindClusterName = envconf.RandomName("crdtest-", 16)
	namespace = envconf.RandomName("my-ns", 10)

	testEnv.Setup(
		envfuncs.CreateCluster(kind.NewProvider(), kindClusterName),
		envfuncs.SetupCRDs("./testdata/crds", "*"),
		envfuncs.CreateNamespace(namespace),
		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			// I have notiecd a lot timing issue with this tests where the actual assessment fails because
			// unable to retrieve the complete list of server APIs: stable.example.com/v1: the server could not find the requested resource
			// But when you manually check, the resource exists. Just that the tests are running a bit too quick and API server has not
			// yet started serving the request it seems like. Since there is no clear way today for us to use wait condition directly to
			// wait for the CRD to be showing up in the list of resource being served by API server, just adding a delay here seem to take
			// care of the problem.
			// TODO: add a wait.For conditional helper that can check and wait for the existence of a CRD resource
			time.Sleep(2 * time.Second)
			return ctx, nil
		},
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.TeardownCRDs("./testdata/crds", "*"),
		envfuncs.DestroyCluster(kindClusterName),
	)

	os.Exit(testEnv.Run(m))
}
