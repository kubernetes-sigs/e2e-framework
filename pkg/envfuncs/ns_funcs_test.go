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
	"context"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestCreateNamespace(t *testing.T) {
	labels := map[string]string{"label": "myns-label"}
	annotations := map[string]string{"ann": "myns-ann"}

	tests := []struct {
		name                string
		opts                []envfuncs.CreateNamespaceOpts
		expectedLabels      map[string]string
		expectedAnnotations map[string]string
	}{
		{
			name:           "CreateBasicNamespace",
			expectedLabels: map[string]string{},
		},
		{
			name:                "CreateNamespaceWithLabelsAndAnn",
			opts:                []envfuncs.CreateNamespaceOpts{envfuncs.WithAnnotations(annotations), envfuncs.WithLabels(labels)},
			expectedLabels:      labels,
			expectedAnnotations: annotations,
		},
	}

	feats := make([]features.Feature, 0, len(tests))
	for _, test := range tests {
		var ns corev1.Namespace
		namespace := envconf.RandomName("create-ns", 16)
		feat := features.New(test.name).
			Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx, err := envfuncs.CreateNamespace(namespace, test.opts...)(ctx, cfg)
				if err != nil {
					t.Fatal("Error creating namespace", err)
				}
				return ctx
			}).
			Assess("namespace created", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				err := cfg.Client().Resources().Get(ctx, namespace, namespace, &ns)
				if err != nil {
					t.Fatal("error getting namespace", err)
				}
				return ctx
			}).
			Assess("namespace labels and annotations", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				// K8s adds the `kubernetes.io/metadata.name` label on namespace creation
				test.expectedLabels["kubernetes.io/metadata.name"] = namespace
				if !reflect.DeepEqual(test.expectedLabels, ns.Labels) {
					t.Errorf("namespace labels do not match. Expected:\n%v but got:\n%v", test.expectedLabels, ns.Labels)
				}
				if !reflect.DeepEqual(test.expectedAnnotations, ns.Annotations) {
					t.Errorf("namespace annotations do not match. Expected:\n%v but got:\n%v", test.expectedAnnotations, ns.Annotations)
				}
				return ctx
			}).
			Assess("namespace stored in config", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				if ns.Name != cfg.Namespace() {
					t.Errorf("namespace stored in config does not match the one created. Expected:\n%v but got:\n%v", ns.Name, cfg.Namespace())
				}
				return ctx
			}).
			Assess("namespace stored in context", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				nsVal := ctx.Value(envfuncs.NamespaceContextKey(ns.Name))
				nsFromCtx, ok := nsVal.(corev1.Namespace)
				if !ok {
					t.Errorf("error casting namespace from context to namespace object. Value:%+v", nsVal)
				} else if !reflect.DeepEqual(nsFromCtx, ns) {
					t.Errorf("namespace stored in context does not match the one created. Expected:\n%v but got:\n%v", &ns, nsFromCtx)
				}

				return ctx
			}).
			Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
				ctx, err := envfuncs.DeleteNamespace(namespace)(ctx, cfg)
				if err != nil {
					t.Error("Error deleting namespace", err)
				}
				return ctx
			}).
			Feature()

		feats = append(feats, feat)
	}

	nsTestenv.Test(t, feats...)
}

func TestDeleteNamespace(t *testing.T) {
	var ns corev1.Namespace
	namespace := envconf.RandomName("delete-ns", 16)
	feat := features.New("DeleteNamespace").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, err := envfuncs.CreateNamespace(namespace)(ctx, cfg)
			if err != nil {
				t.Fatal("Error creating namespace", err)
			}
			ctx, err = envfuncs.DeleteNamespace(namespace)(ctx, cfg)
			if err != nil {
				t.Fatal("Unexpected error deleting namespace", err)
			}
			r, err := resources.New(cfg.Client().RESTConfig())
			if err != nil {
				t.Fatal("Error creating new resources", err)
			}
			err = wait.For(conditions.New(r).ResourceDeleted(&corev1.Namespace{
				ObjectMeta: v1.ObjectMeta{
					Name: namespace,
				},
			}),
				wait.WithImmediate())
			if err != nil {
				t.Fatal("Error waiting for namespace deletion", err)
			}
			return ctx
		}).
		Assess("namespace deleted", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			err := cfg.Client().Resources().Get(ctx, namespace, namespace, &ns)
			if !errors.IsNotFound(err) {
				t.Error("unexpected error when checking if namespace is deleted", err)
			}
			return ctx
		}).
		Feature()

	nsTestenv.Test(t, feat)
}
