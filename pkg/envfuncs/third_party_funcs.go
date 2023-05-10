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

package envfuncs

import (
	"context"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/third_party/flux"
)

func InstallFlux() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		manager := flux.New(c.KubeconfigFile())
		manager.InstallFlux()
		return ctx, nil
	}
}

func UninstallFlux() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		manager := flux.New(c.KubeconfigFile())
		manager.UninstallFlux()
		return ctx, nil
	}
}
