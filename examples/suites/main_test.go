package suites

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
)

var (
	global env.Environment
)
func TestMain(m *testing.M) {
	global = env.New(conf.New())
	ctx := context.WithValue(context.TODO(), 1, "bazz")
	global.BeforeTest(func(ctx context.Context, conf conf.Config) error {
		return nil
	})
	global.Run(ctx, m)
}
