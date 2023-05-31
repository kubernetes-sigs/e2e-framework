/*
Copyright 2023 The Kubernetes Authors.

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

package flux

import (
	"context"
	"fmt"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

var manager *Manager

const NoFluxInstallationFoundMsg = "flux needs to be installed within a cluster first"

func InstallFlux() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		manager = New(c.KubeconfigFile())
		err := manager.installFlux()
		if err != nil {
			return ctx, fmt.Errorf("installation of flux failed: %w", err)
		}
		return ctx, nil
	}
}

func CreateGitRepo(gitRepoName string, gitRepoUrl string, opts ...Option) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		if manager == nil {
			return ctx, fmt.Errorf(NoFluxInstallationFoundMsg)
		}
		err := manager.createSource(Git, gitRepoName, gitRepoUrl, opts...)
		if err != nil {
			return ctx, fmt.Errorf("git reporistory creation failed: %w", err)
		}
		return ctx, nil
	}
}

func CreateKustomization(kustomizationName string, sourceRef string, opts ...Option) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		if manager == nil {
			return ctx, fmt.Errorf(NoFluxInstallationFoundMsg)
		}
		err := manager.createKustomization(kustomizationName, sourceRef, opts...)
		if err != nil {
			return ctx, fmt.Errorf("kustomization creation failed: %w", err)
		}
		return ctx, nil
	}
}

func UninstallFlux() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		if manager == nil {
			return ctx, fmt.Errorf(NoFluxInstallationFoundMsg)
		}
		err := manager.uninstallFlux()
		if err != nil {
			return ctx, fmt.Errorf("uninstallation of flux failed: %w", err)
		}
		return ctx, nil
	}
}

func DeleteKustomization(kustomizationName string, opts ...Option) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		if manager == nil {
			return ctx, fmt.Errorf(NoFluxInstallationFoundMsg)
		}
		err := manager.deleteKustomization(kustomizationName, opts...)
		if err != nil {
			return ctx, fmt.Errorf("kustomization creation failed: %w", err)
		}
		return ctx, nil
	}
}

func DeleteGitRepo(gitRepoName string, opts ...Option) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		if manager == nil {
			return ctx, fmt.Errorf(NoFluxInstallationFoundMsg)
		}
		err := manager.deleteSource(Git, gitRepoName, opts...)
		if err != nil {
			return ctx, fmt.Errorf("git reporistory deletion failed: %w", err)
		}
		return ctx, nil
	}
}
