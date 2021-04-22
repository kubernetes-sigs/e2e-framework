package env

import (
	"context"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/conf"
	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

func TestAction_Run(t *testing.T) {
	tests := []struct {
		name string
		ctx context.Context
		setup func (context.Context) (int, error)
		expected int
		shouldFail bool
	}{
		{
			name : "single-step action",
			ctx : context.WithValue(context.TODO(), 0, 1),
			setup: func(ctx context.Context) (val int, err error) {
				funcs := []types.EnvFunc{
					func(ctx context.Context, config types.Config) error {
						val = 12
						return nil
					},
				}
				err = action{role: roleSetup, funcs: funcs}.run(ctx, conf.New())
				return
			},
			expected: 12,
		},
		{
			name : "multi-step action",
			ctx : context.WithValue(context.TODO(), 0, 1),
			setup: func(ctx context.Context) (val int, err error) {
				funcs := []types.EnvFunc{
					func(ctx context.Context, config types.Config) error {
						val = 12
						return nil
					},
					func(ctx context.Context, config types.Config) error {
						val = val * 2
						return nil
					},
				}
				err = action{role: roleSetup, funcs: funcs}.run(ctx, conf.New())
				return
			},
			expected: 24,
		},
		{
			name : "read from context",
			ctx : context.WithValue(context.TODO(), 0, 1),
			setup: func(ctx context.Context) (val int, err error) {
				funcs := []types.EnvFunc{
					func(ctx context.Context, config types.Config) error {
						i := ctx.Value(0).(int) + 2
						val = i
						return nil
					},
					func(ctx context.Context, config types.Config) error {
						val = val + 3
						return nil
					},
				}
				err = action{role: roleSetup, funcs: funcs}.run(ctx, conf.New())
				return
			},
			expected: 6,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T){
			result, err := test.setup(test.ctx)
			if !test.shouldFail && err != nil{
				t.Fatalf("unexpected failure: %v",err)
			}
			if result != test.expected {
				t.Error("unexpected value:", result)
			}
		})
	}
}
