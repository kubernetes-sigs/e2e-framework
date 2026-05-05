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
	authenticationv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/e2e-framework/klient/k8s"
)

// Bindings is the variable map passed to CEL at evaluation time. The keys
// correspond to CEL variable names (e.g. "object", "oldObject", "self").
type Bindings map[string]any

// AdmissionRequest mirrors the subset of admissionv1.AdmissionRequest that
// admission CEL exposes under the `request` variable. Callers populate only
// the fields their assertions reference.
type AdmissionRequest struct {
	Kind               schema.GroupVersionKind
	Resource           schema.GroupVersionResource
	SubResource        string
	RequestKind        *schema.GroupVersionKind
	RequestResource    *schema.GroupVersionResource
	RequestSubResource string
	Name               string
	Namespace          string
	Operation          string // CREATE, UPDATE, DELETE, CONNECT
	UserInfo           authenticationv1.UserInfo
	DryRun             bool
}

// ObjectBinding binds obj to the `object` variable used by admission CEL.
func ObjectBinding(obj k8s.Object) Bindings {
	return Bindings{"object": toUnstructured(obj)}
}

// OldObjectBinding binds obj to the `oldObject` variable.
func OldObjectBinding(obj k8s.Object) Bindings {
	return Bindings{"oldObject": toUnstructured(obj)}
}

// ParamsBinding binds the param resource a ValidatingAdmissionPolicyBinding
// references to the `params` variable.
func ParamsBinding(params k8s.Object) Bindings {
	return Bindings{"params": toUnstructured(params)}
}

// NamespaceBinding binds the namespace of the object under admission to the
// `namespaceObject` variable.
func NamespaceBinding(ns *corev1.Namespace) Bindings {
	return Bindings{"namespaceObject": toUnstructured(ns)}
}

// RequestBinding binds a test-constructed AdmissionRequest to the `request`
// variable. Only the fields populated on req are surfaced to CEL.
// A nil req binds `request` to nil.
func RequestBinding(req *AdmissionRequest) Bindings {
	if req == nil {
		return Bindings{"request": nil}
	}
	return Bindings{"request": admissionRequestToMap(req)}
}

// SelfBinding binds self to the `self` variable (CRD environment only).
// Using SelfBinding against an admission-environment Evaluator produces a
// CEL compile error at expression evaluation time.
func SelfBinding(self k8s.Object) Bindings {
	return Bindings{"self": toUnstructured(self)}
}

// OldSelfBinding binds oldSelf to the `oldSelf` variable (CRD environment).
func OldSelfBinding(oldSelf k8s.Object) Bindings {
	return Bindings{"oldSelf": toUnstructured(oldSelf)}
}

// Bind composes multiple Bindings. When keys collide, later values win.
func Bind(parts ...Bindings) Bindings {
	out := Bindings{}
	for _, p := range parts {
		for k, v := range p {
			out[k] = v
		}
	}
	return out
}

// variableDecls returns the cel.Variable declarations for a given Env.
func variableDecls(e Env) []cel.EnvOption {
	switch e {
	case EnvCRD:
		return []cel.EnvOption{
			cel.Variable("self", cel.DynType),
			cel.Variable("oldSelf", cel.DynType),
		}
	default: // EnvAdmission
		return []cel.EnvOption{
			cel.Variable("object", cel.DynType),
			cel.Variable("oldObject", cel.DynType),
			cel.Variable("request", cel.DynType),
			cel.Variable("params", cel.DynType),
			cel.Variable("namespaceObject", cel.DynType),
			cel.Variable("authorizer", cel.DynType),
			cel.Variable("variables", cel.DynType),
		}
	}
}

// toUnstructured converts a k8s.Object into the map[string]any shape CEL
// sees. A nil object becomes a nil binding, matching admission semantics
// where `object` is null on DELETE and `oldObject` is null on CREATE.
func toUnstructured(obj k8s.Object) any {
	if obj == nil {
		return nil
	}
	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		// Surface the error at evaluation time rather than at binding time;
		// returning a sentinel map lets CEL fail the assertion cleanly.
		return map[string]any{"__binding_error__": err.Error()}
	}
	return u
}

func admissionRequestToMap(req *AdmissionRequest) map[string]any {
	m := map[string]any{
		"kind":        gvkToMap(req.Kind),
		"resource":    gvrToMap(req.Resource),
		"subResource": req.SubResource,
		"name":        req.Name,
		"namespace":   req.Namespace,
		"operation":   req.Operation,
		"userInfo":    userInfoToMap(req.UserInfo),
		"dryRun":      req.DryRun,
	}
	if req.RequestKind != nil {
		m["requestKind"] = gvkToMap(*req.RequestKind)
	}
	if req.RequestResource != nil {
		m["requestResource"] = gvrToMap(*req.RequestResource)
	}
	if req.RequestSubResource != "" {
		m["requestSubResource"] = req.RequestSubResource
	}
	return m
}

func gvkToMap(gvk schema.GroupVersionKind) map[string]any {
	return map[string]any{"group": gvk.Group, "version": gvk.Version, "kind": gvk.Kind}
}

func gvrToMap(gvr schema.GroupVersionResource) map[string]any {
	return map[string]any{"group": gvr.Group, "version": gvr.Version, "resource": gvr.Resource}
}

func userInfoToMap(u authenticationv1.UserInfo) map[string]any {
	m := map[string]any{
		"username": u.Username,
		"uid":      u.UID,
		"groups":   stringSliceOrEmpty(u.Groups),
	}
	if len(u.Extra) > 0 {
		extra := make(map[string]any, len(u.Extra))
		for k, v := range u.Extra {
			extra[k] = []string(v)
		}
		m["extra"] = extra
	}
	return m
}

func stringSliceOrEmpty(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
