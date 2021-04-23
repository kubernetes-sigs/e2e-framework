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
