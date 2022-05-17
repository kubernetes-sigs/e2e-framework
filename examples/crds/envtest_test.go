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

	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/examples/crds/testdata/crontabs"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestCRDSetup(t *testing.T) {
	feature := features.New("Custom Controller").
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			crontabs.AddToScheme(r.GetScheme())
			r.WithNamespace(namespace)
			decoder.DecodeEachFile(
				ctx, os.DirFS("./testdata/crs"), "*",
				decoder.CreateHandler(r),
				decoder.MutateNamespace(namespace),
			)
			return ctx
		}).
		Assess("Check If Resource created", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			r, err := resources.New(c.Client().RESTConfig())
			if err != nil {
				t.Fail()
			}
			r.WithNamespace(namespace)
			crontabs.AddToScheme(r.GetScheme())
			ct := &crontabs.CronTab{}
			err = r.Get(ctx, "my-new-cron-object", namespace, ct)
			if err != nil {
				t.Fail()
			}
			klog.InfoS("CR Details", "cr", ct)
			return ctx
		}).Feature()

	testEnv.Test(t, feature)
}
