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
	"fmt"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/features"
)

func Hello(name string) string {
	return fmt.Sprintf("Hello %s", name)
}

func TestHello(t *testing.T) {
	name := ctx.Value(1).(string)
	feat := features.New("Hello Feature").
		WithLabel("type", "simple").
		Assess("test message", func(ctx context.Context, t *testing.T) context.Context{
			result := Hello(name)
			if result != "Hello bazz" {
				t.Error("unexpected message")
			}
			return ctx
		}).Feature()

	testenv.Test(ctx, t, feat)
}
