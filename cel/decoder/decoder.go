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

// Package decoder bridges klient/decoder with cel so CEL assertions
// can run over YAML or JSON manifests decoded through the framework's
// decoder. Two shapes are exposed:
//
//   - HandlerFunc factories (AssertHandler, PolicyHandler) that plug into
//     decoder.DecodeEach, DecodeEachFile, DecodeURL.
//   - One-shot helpers (AssertYAML, AssertYAMLAll) for single-document and
//     multi-document inputs when a caller doesn't need streaming.
package decoder

import (
	"context"
	"fmt"
	"io"

	"sigs.k8s.io/e2e-framework/cel"
	"sigs.k8s.io/e2e-framework/cel/policy"
	kdecoder "sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

// AssertHandler returns a decoder.HandlerFunc that asserts expr against
// every decoded object, binding the object to the `object` CEL variable.
// The first assertion failure halts decoding and surfaces as the return
// error of the DecodeEach / DecodeEachFile / DecodeURL call.
func AssertHandler(ev *cel.Evaluator, expr string) kdecoder.HandlerFunc {
	return func(_ context.Context, obj k8s.Object) error {
		return ev.Assert(expr, cel.ObjectBinding(obj))
	}
}

// PolicyHandler returns a decoder.HandlerFunc that runs pol against every
// decoded object. All validations in pol are evaluated (not short-circuit)
// per object; a per-object failure halts decoding with a joined error.
func PolicyHandler(ev *cel.Evaluator, pol policy.Policy) kdecoder.HandlerFunc {
	return func(_ context.Context, obj k8s.Object) error {
		if res := pol.Check(ev, obj); !res.Passed() {
			return res.Err()
		}
		return nil
	}
}

// AssertYAML decodes a single-document YAML or JSON manifest from manifest
// and asserts expr against the resulting object.
func AssertYAML(ev *cel.Evaluator, expr string, manifest io.Reader) error {
	obj, err := kdecoder.DecodeAny(manifest)
	if err != nil {
		return fmt.Errorf("cel: decode: %w", err)
	}
	return ev.Assert(expr, cel.ObjectBinding(obj))
}

// AssertYAMLAll decodes a multi-document YAML or JSON stream and asserts
// expr against every object it contains. The first failure returns its
// error; callers wanting all failures at once should use PolicyHandler
// with a single-validation Policy and DecodeEach directly.
func AssertYAMLAll(ctx context.Context, ev *cel.Evaluator, expr string, manifest io.Reader) error {
	return kdecoder.DecodeEach(ctx, manifest, AssertHandler(ev, expr))
}
