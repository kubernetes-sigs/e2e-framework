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
	"log"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/internal/types"
)

var envForTesting types.Environment

type ctxTestKeyInt struct{}

func TestMain(m *testing.M) {
	// setup new environment test with injected context value
	initialVal := 22
	env, err := NewWithContext(context.WithValue(context.Background(), &ctxTestKeyInt{}, initialVal), envconf.New())
	if err != nil {
		log.Fatalf("Test suite failed to start: %s", err)
	}
	envForTesting = env

	// defined env setup funcs
	// each func will update value inside context
	envForTesting.Setup(
		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			val, ok := ctx.Value(&ctxTestKeyInt{}).(int)
			if !ok {
				log.Fatal("context value was not of expected type int or nil")
			}
			val *= 2 // 44
			return context.WithValue(ctx, &ctxTestKeyInt{}, val), nil
		},
		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			val, ok := ctx.Value(&ctxTestKeyInt{}).(int)
			if !ok {
				log.Fatal("context value was not of expected type int or nil")
			}
			val *= 2 // 88
			return context.WithValue(ctx, &ctxTestKeyInt{}, val), nil
		},
	).BeforeEachTest(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
		// update before each test
		val, ok := ctx.Value(&ctxTestKeyInt{}).(int)
		if !ok {
			log.Fatal("context value was not of expected type int or nil")
		}
		val += 2 // 90
		return context.WithValue(ctx, &ctxTestKeyInt{}, val), nil
	}).AfterEachTest(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
		// update after the test
		val, ok := ctx.Value(&ctxTestKeyInt{}).(int)
		if !ok {
			log.Fatal("context value was not of expected type int or nil")
		}
		val--
		return context.WithValue(ctx, &ctxTestKeyInt{}, val), nil
	})

	os.Exit(envForTesting.Run(m))
}
