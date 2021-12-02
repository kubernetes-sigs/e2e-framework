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

package decoder

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var (
	testenv env.Environment

	namespace string = envconf.RandomName("decoder-ns", 16)
)

var initYAML string = `
apiVersion: v1
kind: Namespace
metadata:
  name: %s
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: mysa
  namespace: %s
`

func TestMain(m *testing.M) {
	testenv = env.New()
	kindClusterName := envconf.RandomName("decoder", 16)
	initYAML = fmt.Sprintf(initYAML, namespace, namespace)

	testenv.Setup(
		envfuncs.CreateKindCluster(kindClusterName),
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				return ctx, err
			}
			// decode and create a stream of YAML or JSON documents from an io.Reader
			decoder.DecodeEach(ctx, strings.NewReader(initYAML), decoder.CreateHandler(r))
			return ctx, nil
		},
	)
	testenv.Finish(
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyKindCluster(kindClusterName),
	)
	os.Exit(testenv.Run(m))
}
