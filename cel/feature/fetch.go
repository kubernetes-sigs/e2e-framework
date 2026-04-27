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

package feature

import (
	"context"
	"fmt"

	"sigs.k8s.io/e2e-framework/cel"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

// Fetch returns a FetcherFunc that retrieves the named resource into target
// on every call. target is a non-nil pointer to a k8s.Object of the expected
// Kind; its fields are populated in place and the same pointer is returned.
//
// If namespace is empty, cfg.Namespace() is used. For cluster-scoped
// resources, pass an empty namespace explicitly and use cfg.Namespace()
// only when your test wants that default.
func Fetch(target k8s.Object, name, namespace string) FetcherFunc {
	return func(ctx context.Context, cfg *envconf.Config) (k8s.Object, error) {
		if target == nil {
			return nil, fmt.Errorf("cel: fetch: target is nil")
		}
		ns := namespace
		if ns == "" {
			ns = cfg.Namespace()
		}
		client, err := cfg.NewClient()
		if err != nil {
			return nil, fmt.Errorf("cel: fetch: new client: %w", err)
		}
		if err := client.Resources().Get(ctx, name, ns, target); err != nil {
			return nil, fmt.Errorf("cel: fetch %s/%s: %w", ns, name, err)
		}
		return target, nil
	}
}

// AsBinder adapts a FetcherFunc into a BinderFunc that binds the fetched
// object to the `object` CEL variable. Use when a caller wants to compose
// several bindings (for example, object plus request) — construct each
// binding with the helper of its choice and merge them with cel.Bind.
func AsBinder(f FetcherFunc) BinderFunc {
	return func(ctx context.Context, cfg *envconf.Config) (cel.Bindings, error) {
		obj, err := f(ctx, cfg)
		if err != nil {
			return nil, err
		}
		return cel.ObjectBinding(obj), nil
	}
}
