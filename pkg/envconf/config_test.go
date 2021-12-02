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
	"os"
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
	cfg, err := NewFromFlags()
	if err != nil {
		t.Error("failed to parse args", err)
	}
	if !cfg.ParallelTestEnabled() {
		t.Error("expected parallel test to be enabled when -parallel argument is provided")
	}
}
