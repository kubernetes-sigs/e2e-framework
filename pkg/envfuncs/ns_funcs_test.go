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
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/support/kind"
)

func TestCreateNamespace(t *testing.T) {
	var err error
	var ns corev1.Namespace
	kindClusterName := envconf.RandomName("my-cluster", 16)
	namespace := envconf.RandomName("myns", 16)
	labels := map[string]string{"label": "myns-label"}
	annotations := map[string]string{"ann": "myns-ann"}

	feat := features.New("CreateNamespaceWithLabelsAndAnn").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			ctx, err = envfuncs.CreateCluster(kind.NewProvider(), kindClusterName)(ctx, cfg)
			if err != nil {
				t.Fatal("Error creating kind cluster", err)
			}
			ctx, err = envfuncs.CreateNamespace(namespace, envfuncs.WithAnnotations(annotations), envfuncs.WithLabels(labels))(ctx, cfg)
			if err != nil {
				t.Fatal("Error creating namespace", err)
			}
			return ctx
		}).
		Assess("namespace created with custom labels and annotations", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// Get namespace
			err := cfg.Client().Resources().Get(ctx, namespace, namespace, &ns)
			if err != nil {
				t.Fatal("error getting namespace", err)
			}

			// Check labels and annotations
			// No DeepEqual for labels because k8s adds the `kubernetes.io/metadata.name` label on creation
			if labels["label"] != ns.Labels["label"] {
				t.Errorf("namespace labels do not match. Expected:\n%v but got:\n%v", labels, ns.Labels)
			}
			if !reflect.DeepEqual(annotations, ns.Annotations) {
				t.Errorf("namespace annotations do not match. Expected:\n%v but got:\n%v", annotations, ns.Annotations)
			}

			// Check namespace is stored in config
			if ns.Name != cfg.Namespace() {
				t.Errorf("namespace stored in config does not match the one created. Expected:\n%v but got:\n%v", ns.Name, cfg.Namespace())
			}

			// Check namespace is stored in context
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
			ctx, err = envfuncs.DestroyCluster(kindClusterName)(ctx, cfg)
			if err != nil {
				t.Error("Error destroying kind cluster", err)
			}
			return ctx
		}).
		Feature()

	testenv.Test(t, feat)
}
