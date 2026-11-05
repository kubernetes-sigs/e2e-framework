/*
Copyright 2026 The Kubernetes Authors.

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

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoBinDir(t *testing.T) {
	gobin := t.TempDir()
	gopath1 := t.TempDir()
	gopath2 := t.TempDir()

	tests := []struct {
		name   string
		gobin  string
		gopath string
		want   string
	}{
		{
			name:   "GOBIN set",
			gobin:  gobin,
			gopath: gopath1,
			want:   gobin,
		},
		{
			name:   "GOBIN unset",
			gopath: gopath1,
			want:   filepath.Join(gopath1, "bin"),
		},
		{
			name:   "GOPATH with several entries",
			gopath: gopath1 + string(os.PathListSeparator) + gopath2,
			want:   filepath.Join(gopath1, "bin"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GOBIN", tt.gobin)
			t.Setenv("GOPATH", tt.gopath)
			if got := goBinDir(); got != tt.want {
				t.Errorf("goBinDir() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGoBinDirDefaultGOPATH(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("GOBIN", "")
	t.Setenv("GOPATH", "")
	os.Unsetenv("GOPATH")

	want := filepath.Join(home, "go", "bin")
	if got := goBinDir(); got != want {
		t.Errorf("goBinDir() = %q, want %q", got, want)
	}
}
