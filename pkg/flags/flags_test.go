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
			args:  []string{"-assess", "volume test", "--feature", "beta", "--labels", "k0=v0, k1=v1, k2=v2"},
			flags: &EnvFlags{assess: "volume test", feature: "beta", labels: LabelsMap{"k0": "v0", "k1": "v1", "k2": "v2"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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
				if test.flags.Labels()[k] != v {
					t.Errorf("unmatched label %s=%s", k, test.flags.Labels()[k])
				}
			}
		})
	}
}
