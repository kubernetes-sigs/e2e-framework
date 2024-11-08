/*
Copyright 2022 The Kubernetes Authors.

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

package tester

import (
	"testing"

	"github.com/octago/sflags/gen/gpflag"
)

func TestBuildFlags(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		tester *Tester
	}{
		{
			name:   "empty flags",
			args:   []string{},
			tester: &Tester{},
		},
		{
			name: "all flags",
			args: []string{
				"--packages", ".",
				"--assess", "volume test",
				"--feature", "beta",
				"--labels", "k0=v0, k1=v1, k2=v2",
				"--skip-labels", "k0=v0, k1=v1",
				"--skip-features", "networking",
				"--skip-assessment", "volume test",
				"--parallel",
				"--dry-run",
				"--disable-graceful-teardown",
			},
			tester: &Tester{
				Packages:                ".",
				Assess:                  "volume test",
				Feature:                 "beta",
				Labels:                  "k0=v0, k1=v1, k2=v2",
				SkipLabels:              "k0=v0, k1=v1",
				SkipFeatures:            "networking",
				SkipAssessment:          "volume test",
				Parallel:                true,
				DryRun:                  true,
				DisableGracefulTeardown: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tester := &Tester{}
			fs, err := gpflag.Parse(tester)
			if err != nil {
				t.Fatal(err)
			}
			if err := fs.Parse(test.args); err != nil {
				t.Fatalf("failed to parse flags: %s", err)
			}
			if tester.Packages != test.tester.Packages {
				t.Errorf("flag 'Package' not matched: expecting %s, got %s", test.tester.Packages, tester.Packages)
			}
			if tester.Assess != test.tester.Assess {
				t.Errorf("flag 'Assess' not matched: expecting %s, got %s", test.tester.Assess, tester.Assess)
			}
			if tester.Feature != test.tester.Feature {
				t.Errorf("flag 'Feature' not matched: expecting %s, got %s", test.tester.Feature, tester.Feature)
			}
			if tester.Labels != test.tester.Labels {
				t.Errorf("flag 'Labels' not matched: expecting %s, got %s", test.tester.Labels, tester.Labels)
			}
			if tester.SkipLabels != test.tester.SkipLabels {
				t.Errorf("flag 'SkipLabels' not matched: expecting %s, got %s", test.tester.SkipLabels, tester.SkipLabels)
			}
			if tester.SkipFeatures != test.tester.SkipFeatures {
				t.Errorf("flag 'SkipFeatures' not matched: expecting %s, got %s", test.tester.SkipFeatures, tester.SkipFeatures)
			}
			if tester.SkipAssessment != test.tester.SkipAssessment {
				t.Errorf("flag 'SkipAssessment' not matched: expecting %s, got %s", test.tester.SkipAssessment, tester.SkipAssessment)
			}
			if tester.Parallel != test.tester.Parallel {
				t.Errorf("flag 'Parallel' not matched: expecting %t, got %t", test.tester.Parallel, tester.Parallel)
			}
			if tester.DryRun != test.tester.DryRun {
				t.Errorf("flag 'DryRun' not matched: expecting %t, got %t", test.tester.DryRun, tester.DryRun)
			}
			if tester.DisableGracefulTeardown != test.tester.DisableGracefulTeardown {
				t.Errorf("flag 'DisableGracefulTeardown' not matched: expecting %t, got %t", test.tester.DisableGracefulTeardown, tester.DisableGracefulTeardown)
			}
		})
	}
}
