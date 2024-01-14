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

package envconf

import (
	"flag"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
)

func TestConfig_New(t *testing.T) {
	cfg := New()
	if cfg.client != nil {
		t.Errorf("client should be nil")
	}
	if cfg.namespace != "" {
		t.Errorf("namespace should be empty")
	}
	if cfg.featureRegex != nil || cfg.assessmentRegex != nil {
		t.Errorf("regex filters should be nil")
	}
	if cfg.ParallelTestEnabled() {
		t.Errorf("parallel test should be disabled by default")
	}
}

func TestConfig_New_WithParallel(t *testing.T) {
	os.Args = []string{"test-binary", "-parallel"}
	flag.CommandLine = &flag.FlagSet{}
	cfg, err := NewFromFlags()
	if err != nil {
		t.Error("failed to parse args", err)
	}
	if !cfg.ParallelTestEnabled() {
		t.Error("expected parallel test to be enabled when -parallel argument is provided")
	}
}

func TestConfig_New_WithDryRun(t *testing.T) {
	os.Args = []string{"test-binary", "--dry-run"}
	flag.CommandLine = &flag.FlagSet{}
	cfg, err := NewFromFlags()
	if err != nil {
		t.Error("failed to parse args", err)
	}
	if !cfg.DryRunMode() {
		t.Errorf("expected dryRun mode to be enabled with invoked with --dry-run arguments")
	}
}

func TestConfig_New_WithFailFastAndIgnoreFinalize(t *testing.T) {
	flag.CommandLine = &flag.FlagSet{}
	os.Args = []string{"test-binary", "-fail-fast"}
	cfg, err := NewFromFlags()
	if err != nil {
		t.Error("failed to parse args", err)
	}
	if !cfg.FailFast() {
		t.Error("expected fail-fast mode to be enabled when -fail-fast argument is passed")
	}
}

func TestConfig_New_WithIgnorePanicRecovery(t *testing.T) {
	flag.CommandLine = &flag.FlagSet{}
	os.Args = []string{"test-binary", "-disable-graceful-teardown"}
	cfg, err := NewFromFlags()
	if err != nil {
		t.Error("failed to parse args", err)
	}
	if !cfg.DisableGracefulTeardown() {
		t.Error("expected ignore-panic-recovery mode to be enabled when -disable-graceful-teardown argument is passed")
	}
}

func TestRandomName(t *testing.T) {
	t.Run("no prefix yields random name without dash", func(t *testing.T) {
		out := RandomName("", 16)
		if strings.Contains(out, "-") {
			t.Errorf("random name %q shouldn't contain a dash when no prefix provided", out)
		}
	})

	t.Run("non empty prefix yields random name with dash", func(t *testing.T) {
		out := RandomName("abc", 16)
		if !strings.Contains(out, "-") {
			t.Errorf("random name %q should contain a dash when prefix provided", out)
		}
	})
}

func TestClientIsGorountineSafe(t *testing.T) {
	t.Parallel()

	cfg := New()
	wg := sync.WaitGroup{}
	count := 10000
	wgCount := runtime.NumCPU()
	wg.Add(wgCount)

	for i := 0; i < wgCount; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < count; i++ {

				client := cfg.Client()
				if client == nil {
					t.Errorf("client should not be nil")
				}
			}
		}()
	}
	wg.Wait()
}
