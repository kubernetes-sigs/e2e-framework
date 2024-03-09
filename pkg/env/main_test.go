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
	"os"
	"testing"

	log "k8s.io/klog/v2"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/types"
)

var envForTesting types.Environment

type ctxTestKeyString struct{}

func TestMain(m *testing.M) {
	// setup new environment test with injected context value
	initialVal := []string{}
	env, err := NewWithContext(context.WithValue(context.Background(), &ctxTestKeyString{}, initialVal), envconf.New())
	if err != nil {
		log.Fatalf("Test suite failed to start: %s", err)
	}
	envForTesting = env

	// defined env setup funcs
	// each func will update value inside context
	envForTesting.Setup(
		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
			if !ok {
				log.Fatal("context value was not of expected type []string or nil")
			}
			val = append(val, "setup-1")
			return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
		},
		func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
			val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
			if !ok {
				log.Fatal("context value was not of expected type []string or nil")
			}
			val = append(val, "setup-2")
			return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
		},
	).BeforeEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
		// update before each test
		val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
		if !ok {
			log.Fatal("context value was not of expected type []string or nil")
		}
		val = append(val, "before-each-test")
		return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
	}).AfterEachTest(func(ctx context.Context, _ *envconf.Config, t *testing.T) (context.Context, error) {
		// update after the test
		val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
		if !ok {
			log.Fatal("context value was not of expected type []string] or nil")
		}
		val = append(val, "after-each-test")
		return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
	}).Finish(func(ctx context.Context, _ *envconf.Config) (context.Context, error) {
		// update after the test suite
		val, ok := ctx.Value(&ctxTestKeyString{}).([]string)
		if !ok {
			log.Fatal("context value was not of expected type []string] or nil")
		}
		// this will only be accessible after the whole suite run
		val = append(val, "finish")
		return context.WithValue(ctx, &ctxTestKeyString{}, val), nil
	})

	os.Exit(envForTesting.Run(m))
}
