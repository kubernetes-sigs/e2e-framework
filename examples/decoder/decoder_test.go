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

package decoder

import (
	"context"
	"os"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestDecoder(t *testing.T) {
	// use files matching the Glob testdata/* in the following tests
	testdata := os.DirFS("testdata")
	pattern := "*"

	decodeAll := features.New("decoding objects").Assess("Single File", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		f, err := os.Open("testdata/configmap.yaml")
		if err != nil {
			t.Fatal(err)
		}
		obj, err := decoder.DecodeAny(f)
		if err != nil {
			t.Fatal(err)
		}
		configMap, ok := obj.(*v1.ConfigMap)
		if !ok {
			t.Fatal("object decoded to unexpected type")
		}
		t.Log(configMap)
		return ctx
	}).Assess("Multiple Files", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		objects, err := decoder.DecodeAllFiles(ctx, testdata, pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, obj := range objects {
			t.Log(obj.GetObjectKind(), obj.GetNamespace(), obj.GetName())
		}
		return ctx
	}).Feature()

	decoderHandlerFuncs := features.New("setup and teardown").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}

			if err := decoder.DecodeEachFile(ctx, testdata, pattern,
				decoder.CreateHandler(r),           // try to CREATE objects after decoding
				decoder.MutateNamespace(namespace), // inject a namespace into decoded objects, before calling CreateHandler
			); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("objects created", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			// check for decoded object object creation
			if err := client.Resources(namespace).Get(ctx, "my-config", namespace, &v1.ConfigMap{}); err != nil {
				t.Fatal(err)
			}
			if err := client.Resources(namespace).Get(ctx, "myapp", namespace, &appsv1.Deployment{}); err != nil {
				t.Fatal(err)
			}
			if err := client.Resources(namespace).Get(ctx, "myapp", namespace, &v1.Service{}); err != nil {
				t.Fatal(err)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// remove test resources before exiting
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}
			if err := decoder.DecodeEachFile(ctx, testdata, pattern,
				decoder.DeleteHandler(r),           // try to DELETE objects after decoding
				decoder.MutateNamespace(namespace), // inject a namespace into decoded objects, before calling DeleteHandler
			); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testenv.Test(t, decodeAll, decoderHandlerFuncs)
}
