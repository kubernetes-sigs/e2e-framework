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

package flags

import (
	"context"
	"log"
	"os"
	"testing"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func greet(key string) string {
	salut := map[string]string{
		"en": "Hello",
		"fr": "Allo",
		"es": "Olá",
	}
	return salut[key]
}

var test env.Environment

// TestMain sets up a test suite and configures the test
// environment using CLI flags. Note that the test.* flags
// along with the framework flags will be available.
//
// Pass flags in couple of ways:
//
//   go test -v . -args --assess en
//
// Or, build a test binary first:
//
//   go test -c -o flags.test .
//
// Then, execute the test:
//
//  ./flags.test --assess en
func TestMain(m *testing.M) {
	// create config from flags (always in TestMain because it calls flag.Parse())
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		log.Fatalf("failed to build envconf from flags: %s", err)
	}
	test = env.NewWithConfig(cfg)
	os.Exit(test.Run(m))
}

func TestWithFlags(t *testing.T) {
	f := features.New("salutation").WithLabel("type", "lang")
	f.Assess("en", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
		if greet("en") != "Hello" {
			t.Error("unexpected message: en")
		}
		return ctx
	})
	f.Assess("es", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
		if greet("es") != "Olá" {
			t.Error("unexpected message: es")
		}
		return ctx
	})

	test.Test(t, f.Feature())
}
