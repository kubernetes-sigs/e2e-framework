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

package cel

import (
	"github.com/google/cel-go/cel"
	"k8s.io/apiserver/pkg/cel/library"
)

// Library identifies a CEL library that the Kubernetes API server registers
// for admission CEL. NewEvaluator wires in every library by default; use
// WithLibraries to select a narrower set.
type Library int

const (
	LibAuthz Library = iota
	LibAuthzSelectors
	LibQuantity
	LibURLs
	LibIP
	LibCIDR
	LibRegex
	LibLists
	LibFormat
	LibSemver
	LibJSONPatch
)

// librarySet is the set of selected libraries, in a deterministic order.
type librarySet struct {
	libs []Library
}

func newLibrarySet(libs []Library) librarySet {
	return librarySet{libs: libs}
}

// allLibraries returns a librarySet containing every supported library.
func allLibraries() librarySet {
	return librarySet{libs: []Library{
		LibAuthz,
		LibAuthzSelectors,
		LibQuantity,
		LibURLs,
		LibIP,
		LibCIDR,
		LibRegex,
		LibLists,
		LibFormat,
		LibSemver,
		LibJSONPatch,
	}}
}

// envOptions translates the selected libraries into cel.EnvOption values.
func (s librarySet) envOptions() []cel.EnvOption {
	out := make([]cel.EnvOption, 0, len(s.libs))
	for _, l := range s.libs {
		if opt := libraryEnvOption(l); opt != nil {
			out = append(out, opt)
		}
	}
	return out
}

func libraryEnvOption(l Library) cel.EnvOption {
	switch l {
	case LibAuthz:
		return library.Authz()
	case LibAuthzSelectors:
		return library.AuthzSelectors()
	case LibQuantity:
		return library.Quantity()
	case LibURLs:
		return library.URLs()
	case LibIP:
		return library.IP()
	case LibCIDR:
		return library.CIDR()
	case LibRegex:
		return library.Regex()
	case LibLists:
		return library.Lists()
	case LibFormat:
		return library.Format()
	case LibSemver:
		return library.SemverLib()
	case LibJSONPatch:
		return library.JSONPatch()
	}
	return nil
}
