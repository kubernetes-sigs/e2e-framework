/*
Copyright 2026 The Kubernetes Authors.

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

package env

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestDeepCopyConfig_WithFakeClient(t *testing.T) {
	// Create a fake controller-runtime client
	fakeClient := fake.NewClientBuilder().
		WithObjects(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
		}).
		Build()

	// Create a klient.Client from the fake client
	klClient := klient.NewFromCRClient(fakeClient)

	// Create an envconf.Config with the fake-backed client
	cfg := envconf.New()
	cfg.WithClient(klClient)

	// Create a test environment with the config
	testEnv := NewWithConfig(cfg)

	// Create a simple feature to trigger deepCopyConfig via processTests
	// This will create a child environment which calls deepCopyConfig
	// and should not panic even though the client has a nil RESTConfig
	feature := features.New("test feature").
		Assess("test assessment", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// Verify the client is still available
			if cfg.Client() == nil {
				t.Error("expected client to be available in copied config")
			}
			return ctx
		}).
		Feature()

	// This will trigger deepCopyConfig internally
	testEnv.Test(t, feature)
}

func TestDeepCopyConfig_WithFakeClientAndNamespace(t *testing.T) {
	// Create a fake controller-runtime client
	fakeClient := fake.NewClientBuilder().
		WithObjects(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "test-ns",
			},
		}).
		Build()

	// Create a klient.Client from the fake client
	klClient := klient.NewFromCRClient(fakeClient)

	// Create an envconf.Config with the fake-backed client and a namespace
	cfg := envconf.New()
	cfg.WithClient(klClient)
	cfg.WithNamespace("test-ns")

	// Create a test environment with the config
	testEnv := NewWithConfig(cfg)

	// Create a feature to trigger deepCopyConfig
	feature := features.New("test feature with namespace").
		Assess("test assessment", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// Verify the namespace is preserved in the copied config
			if cfg.Namespace() != "test-ns" {
				t.Errorf("expected namespace test-ns, got %s", cfg.Namespace())
			}
			// Verify the client is still available
			if cfg.Client() == nil {
				t.Error("expected client to be available in copied config")
			}
			return ctx
		}).
		Feature()

	// This will trigger deepCopyConfig internally
	testEnv.Test(t, feature)
}
