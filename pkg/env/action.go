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

package env

import (
	"context"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

const (
	roleSetup = iota
	roleBeforeTest
	roleBeforeFeature
	roleAfterFeature
	roleAfterTest
	roleFinish
)

// action a group env functions
type action struct {
	role  actionRole
	funcs []types.EnvFunc
}

func (a action) run(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
	for _, f := range a.funcs {
		if f == nil {
			continue
		}

		var err error
		ctx, err = f(ctx, cfg)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}
