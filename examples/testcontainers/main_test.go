/*
Copyright 2025 The Kubernetes Authors.

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

package testcontainers

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/k3s"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

const (
	k3sImage = `rancher/k3s:v1.32.1-k3s1` // upstream k3s image
)

// Define custom type for context key to avoid collisions
type contextKey string

// define context key for storing k3s container reference
// this will point to  *k3s.K3sContainer
const k3sContainerKey contextKey = "k3sContainer"

// various variables for our test
var (
	testEnv   env.Environment
	namespace string
	images    = []string{"nginx:latest", "busybox:latest"} // list of images to load into the k3s cluster
)

func TestMain(m *testing.M) {
	// initialize the test environment first
	testEnv = env.New()

	// random namespace for each test run
	namespace = envconf.RandomName("testcontainers-k3s-ns", 11)

	testEnv.Setup(
		setupClusterAndLoadImages(images...), // setup k3s cluster and load images
		envfuncs.CreateNamespace(namespace),  // create our namespace for tests
	)

	testEnv.Finish(
		envfuncs.DeleteNamespace(namespace), // delete the namespace we created
		terminateK3sContainer(),             // terminate the k3s container, uses context key to find it
	)

	os.Exit(testEnv.Run(m))
}

// setupClusterAndLoadImages returns an environment function that sets up k3s cluster
func setupClusterAndLoadImages(images ...string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		// start k3s cluster in a Container using testcontainers
		k3sContainer, err := k3s.Run(ctx, k3sImage)
		if err != nil {
			return ctx, fmt.Errorf("failed to start k3s container: %w", err)
		}

		// store container reference in context for subsequent use
		ctx = context.WithValue(ctx, k3sContainerKey, k3sContainer)

		// we need to pull images first before loading them into k3s
		// try docker provider first
		provider, err := testcontainers.ProviderDocker.GetProvider()
		if err != nil {
			// try podman as a fallback
			provider, err = testcontainers.ProviderPodman.GetProvider()
			if err != nil {
				return ctx, fmt.Errorf("failed to get provider, tried both docker and podman")
			}
		}

		// pull images
		for _, image := range images {
			if err := provider.PullImage(ctx, image); err != nil {
				return ctx, fmt.Errorf("failed to pull image %s: %w", image, err)
			}
		}

		// load various images into the k3s cluster using testcontainers
		if err := k3sContainer.LoadImages(ctx, images...); err != nil {
			return ctx, fmt.Errorf("failed to load images into k3s cluster: %w", err)
		}

		// grab kubeconfig YAML from testcontainers
		// this is actual content of kubeconfig file in []byte format
		// and NOT a path to the kubeconfig file
		kubeCfgYaml, err := k3sContainer.GetKubeConfig(ctx)
		if err != nil {
			return ctx, fmt.Errorf("failed to get kubeconfig: %w", err)
		}

		// grab the rest config from the kubeconfig YAML
		restConfig, err := clientcmd.RESTConfigFromKubeConfig(kubeCfgYaml)
		if err != nil {
			return ctx, fmt.Errorf("failed to create rest config from kubeconfig: %w", err)
		}

		// creat a new klient from the rest config
		client, err := klient.New(restConfig)
		if err != nil {
			return ctx, fmt.Errorf("failed to create klient from rest config: %w", err)
		}

		// update the current config with the new klient client
		cfg.WithClient(client)

		return ctx, nil
	}
}

// terminateK3sContainer returns an environment function that terminates the container
// created by testcontainers
func terminateK3sContainer() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		// retrieve k3s container reference from context we
		// stored during setup
		containerVal := ctx.Value(k3sContainerKey)
		if containerVal != nil {
			if k3sContainer, ok := containerVal.(*k3s.K3sContainer); ok {
				if err := k3sContainer.Terminate(ctx); err != nil {
					return ctx, fmt.Errorf("failed to terminate k3s container: %w", err)
				}
			}
		}
		return ctx, nil
	}
}
