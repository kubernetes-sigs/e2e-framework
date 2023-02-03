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
	"flag"
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name  string
		args  []string
		flags *EnvFlags
	}{
		{
			name:  "with all",
			args:  []string{"-assess", "volume test", "--feature", "beta", "--labels", "k0=v0, k0=v01, k1=v1, k1=v11, k2=v2", "--skip-labels", "k0=v0, k1=v1", "-skip-features", "networking", "-skip-assessment", "volume test", "-parallel", "--dry-run", "--disable-graceful-teardown"},
			flags: &EnvFlags{assess: "volume test", feature: "beta", labels: LabelsMap{"k0": {"v0", "v01"}, "k1": {"v1", "v11"}, "k2": {"v2"}}, skiplabels: LabelsMap{"k0": {"v0"}, "k1": {"v1"}}, skipFeatures: "networking", skipAssessments: "volume test"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = &flag.FlagSet{}
			testFlags, err := ParseArgs(test.args)
			if err != nil {
				t.Fatal(err)
			}

			if testFlags.Feature() != test.flags.Feature() {
				t.Errorf("unmatched feature: %s; %s", testFlags.Feature(), test.flags.Feature())
			}
			if testFlags.Assessment() != test.flags.Assessment() {
				t.Errorf("unmatched assessment: %s", testFlags.Assessment())
			}

			for k, v := range testFlags.Labels() {
				if !reflect.DeepEqual(test.flags.Labels()[k], v) {
					t.Errorf("unmatched labels %s=%v", k, test.flags.Labels()[k])
				}
			}

			for k, v := range testFlags.SkipLabels() {
				if !reflect.DeepEqual(test.flags.SkipLabels()[k], v) {
					t.Errorf("unmatched skip labels %s=%v", k, test.flags.Labels()[k])
				}
			}

			if testFlags.SkipFeatures() != test.flags.SkipFeatures() {
				t.Errorf("unmatched feature for skip: %s", testFlags.SkipFeatures())
			}

			if testFlags.SkipAssessment() != test.flags.SkipAssessment() {
				t.Errorf("unmatched assessment name for skip: %s", testFlags.SkipFeatures())
			}

			if !testFlags.Parallel() {
				t.Errorf("unmatched flag parsed. Expected parallel to be true.")
			}

			if !testFlags.DryRun() {
				t.Errorf("unmatched flag parsed. Expected dryRun to be true.")
			}

			if !testFlags.DisableGracefulTeardown() {
				t.Errorf("unmatched flag parsed. Expected disableGracefulTeardown to be true")
			}
		})
	}
}

func TestLabelsMap_Contains(t *testing.T) {
	type args struct {
		key string
		val string
	}
	tests := []struct {
		name string
		m    LabelsMap
		args args
		want bool
	}{
		{
			name: "empty map",
			m:    LabelsMap{},
			args: args{
				key: "somekey",
				val: "someval",
			},
			want: false,
		},
		{
			name: "key does not exist",
			m:    LabelsMap{"key0": {"val0"}},
			args: args{
				key: "key1",
				val: "val1",
			},
			want: false,
		},
		{
			// TODO (@embano1): #https://github.com/kubernetes-sigs/e2e-framework/issues/198
			name: "lower-case key for upper case key does not exist",
			m:    LabelsMap{"Key0": {"val0"}},
			args: args{
				key: "key1",
				val: "val1",
			},
			want: false,
		},
		{
			name: "value for existing key does not exist",
			m:    LabelsMap{"key0": {"val0"}},
			args: args{
				key: "key0",
				val: "val1",
			},
			want: false,
		},
		{
			name: "value for map with one key with one value exists",
			m:    LabelsMap{"key0": {"val0"}},
			args: args{
				key: "key0",
				val: "val0",
			},
			want: true,
		},
		{
			name: "value for map with one key with multiple values exists",
			m:    LabelsMap{"key0": {"val0", "val1"}},
			args: args{
				key: "key0",
				val: "val1",
			},
			want: true,
		},
		{
			name: "value for map with multiple keys and values exists",
			m:    LabelsMap{"key0": {"val0", "val1"}, "key1": {"val1"}},
			args: args{
				key: "key1",
				val: "val1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Contains(tt.args.key, tt.args.val); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
