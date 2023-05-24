package test

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestK3D(t *testing.T) {
	countNamespaces := features.New("count namespaces").
		WithLabel("type", "ns-count").
		Assess("namespace exist", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			var nspaces corev1.NamespaceList
			err := cfg.Client().Resources().List(context.TODO(), &nspaces)
			if err != nil {
				t.Fatal(err)
			}
			if len(nspaces.Items) == 1 {
				t.Fatal("no other namespace")
			}
			return ctx
		}).Feature()

	// test feature
	testenv.Test(t, countNamespaces)
}
