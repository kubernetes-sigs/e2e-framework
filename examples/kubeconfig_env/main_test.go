package kubeconfigenv

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var testEnv env.Environment

func TestMain(m *testing.M) {
	// skips flag parsing
	testEnv = env.New()

	os.Exit(testEnv.Run(m))
}

func TestKubeconfig(t *testing.T) {
	testEnv.Test(t,
		features.New("Kubeconfig").
			Assess("Read Kubeconfig", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
				assert.True(t, strings.HasPrefix(os.Getenv("KUBECONFIG"), ".kube/config"+string(filepath.ListSeparator)))
				assert.Equal(t, ".kube/config", conf.ResolveKubeConfigFile())
				return ctx
			}).
			Feature(),
	)
}
