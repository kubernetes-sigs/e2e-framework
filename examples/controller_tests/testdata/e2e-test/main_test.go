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

package e2eframework

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
	"sigs.k8s.io/e2e-framework/support/kind"
	"sigs.k8s.io/e2e-framework/support/utils"
)

var (
	testEnv env.Environment

	dockerImage  = "cronjob-controller:v0.0.1"
	kustomizeVer = "v5.5.0"
	ctrlGenVer   = "v0.16.4"

	certMgrVer = "v1.13.1"
	certMgrUrl = fmt.Sprintf("https://github.com/jetstack/cert-manager/releases/download/%s/cert-manager.yaml", certMgrVer)
	promVer    = "v0.60.0"
	promUrl    = fmt.Sprintf("https://github.com/prometheus-operator/prometheus-operator/releases/download/%s/bundle.yaml", promVer)

	namespace = "cronjob"
)

func TestMain(m *testing.M) {
	testEnv = env.New()
	kindClusterName := "kind-test"
	kindCluster := kind.NewCluster(kindClusterName)

	// Use Environment.Setup to configure pre-test setup (i.e. create cluster, build code, allocate resources, etc)
	log.Println("Creating KinD cluster...")
	testEnv.Setup(
		envfuncs.CreateCluster(kindCluster, kindClusterName),
		envfuncs.CreateNamespace(namespace),

		// install cluster dependencies mgr
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			log.Println("Installing prometheus operator...")
			if p := utils.RunCommand(fmt.Sprintf("kubectl apply -f %s --server-side", promUrl)); p.Err() != nil {
				log.Printf("Failed to deploy prometheus: %s: %s", p.Err(), p.Out())
				return ctx, p.Err()
			}

			log.Println("Installing cert-manager...")
			client := cfg.Client()

			if p := utils.RunCommand(fmt.Sprintf("kubectl apply -f %s", certMgrUrl)); p.Err() != nil {
				log.Printf("Failed to deploy cert-manager: %s: %s", p.Err(), p.Out())
				return ctx, p.Err()
			}

			// wait for CertManager to be ready
			log.Println("Waiting for cert-manager deployment to be available...")
			if err := wait.For(
				conditions.New(client.Resources()).DeploymentAvailable("cert-manager-webhook", "cert-manager"),
				wait.WithTimeout(5*time.Minute),
				wait.WithInterval(10*time.Second),
			); err != nil {
				log.Printf("Timedout while waiting for cert-manager deployment: %s", err)
				return ctx, err
			}
			return ctx, nil
		},

		// install tool dependencies
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			log.Println("Installing bin tools...")
			if p := utils.RunCommand(fmt.Sprintf("go install sigs.k8s.io/kustomize/kustomize/v5@%s", kustomizeVer)); p.Err() != nil {
				log.Printf("Failed to install kustomize binary: %s: %s", p.Err(), p.Out())
				return ctx, p.Err()
			}
			if p := utils.RunCommand(fmt.Sprintf("go install sigs.k8s.io/controller-tools/cmd/controller-gen@%s", ctrlGenVer)); p.Err() != nil {
				log.Printf("Failed to install controller-gen binary: %s: %s", p.Err(), p.Out())
				return ctx, p.Err()
			}
			return ctx, nil
		},

		// generate and deploy resource configurations
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			log.Println("Building source components...")
			origWd, _ := os.Getwd()

			// change dir for Make file or it will fail
			if err := os.Chdir("../"); err != nil {
				log.Printf("Unable to set working directory: %s", err)
				return ctx, err
			}

			// gen manifest files
			log.Println("Generate manifests...")
			if p := utils.RunCommand(`controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases`); p.Err() != nil {
				log.Printf("Failed to generate manifests: %s: %s", p.Err(), p.Out())
				return ctx, p.Err()
			}

			// gen api objects
			log.Println("Generate API objects...")
			if p := utils.RunCommand(`controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."`); p.Err() != nil {
				log.Printf("Failed to generate API objects: %s: %s", p.Err(), p.Out())
				return ctx, p.Err()
			}

			// Build docker image
			log.Println("Building docker image...")
			if p := utils.RunCommand(fmt.Sprintf("docker build -t %s .", dockerImage)); p.Err() != nil {
				log.Printf("Failed to build docker image: %s: %s", p.Err(), p.Out())
				return ctx, p.Err()
			}

			// Load docker image into kind
			log.Println("Loading docker image into kind cluster...")
			if err := kindCluster.LoadImage(ctx, dockerImage); err != nil {
				log.Printf("Failed to load image into kind: %s", err)
				return ctx, err
			}

			// Deploy components
			log.Println("Deploying controller-manager resources...")
			if p := utils.RunCommand(`bash -c "kustomize build config/default | kubectl apply --server-side -f -"`); p.Err() != nil {
				log.Printf("Failed to deploy resource configurations: %s: %s", p.Err(), p.Out())
				return ctx, p.Err()
			}

			// wait for controller-manager to be ready
			log.Println("Waiting for controller-manager deployment to be available...")
			client := cfg.Client()
			if err := wait.For(
				conditions.New(client.Resources()).DeploymentAvailable("cronjob-controller-manager", "cronjob-system"),
				wait.WithTimeout(3*time.Minute),
				wait.WithInterval(10*time.Second),
			); err != nil {
				log.Printf("Timed out while waiting for cert-manager deployment: %s", err)
				return ctx, err
			}

			if err := os.Chdir(origWd); err != nil {
				log.Printf("Unable to set working directory: %s", err)
				return ctx, err
			}

			return ctx, nil
		},
	)

	// Use the Environment.Finish method to define clean up steps
	testEnv.Finish(
		func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
			log.Println("Finishing tests, cleaning cluster ...")
			utils.RunCommand(fmt.Sprintf("kubectl delete -f %s", promUrl))
			utils.RunCommand(fmt.Sprintf("kubectl delete -f %s", certMgrUrl))
			utils.RunCommand(`bash -c "kustomize build config/default | kubectl delete -f -"`)
			return ctx, nil
		},
		envfuncs.DeleteNamespace(namespace),
		envfuncs.DestroyCluster(kindClusterName),
	)

	// Use Environment.Run to launch the test
	os.Exit(testEnv.Run(m))
}
