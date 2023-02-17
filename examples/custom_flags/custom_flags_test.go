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

package custom_flags

import (
	"context"
	"flag"
	"os"
	"testing"

	log "k8s.io/klog/v2"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	test         env.Environment
	myCustomFlag string
)

func TestMain(m *testing.M) {
	flag.StringVar(&myCustomFlag, "my-custom-flag", "", "my custom flag for my tests")
	// create config from flags (always in TestMain or init handler of the package before calling envconf.NewFromFlags())
	cfg, err := envconf.NewFromFlags()
	if err != nil {
		log.Fatalf("failed to build envconf from flags: %s", err)
	}
	test = env.NewWithConfig(cfg)
	os.Exit(test.Run(m))
}

func TestWithCustomFlag(t *testing.T) {
	f := features.New("feature")
	f.Assess("custom flag", func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
		t.Logf("Custom flag my-custom-flag: %v", myCustomFlag)
		return ctx
	})
	test.Test(t, f.Feature())
}
