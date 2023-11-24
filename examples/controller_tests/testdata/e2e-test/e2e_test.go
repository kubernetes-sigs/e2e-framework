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
	"strings"
	"testing"
	"time"

	batchV1 "k8s.io/api/batch/v1"
	coreV1 "k8s.io/api/core/v1"
	apiextensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cronjobV1 "tutorial.kubebuilder.io/project/api/v1"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

var (
	cronjobName = "cronjob-controller"

	cronjob = &cronjobV1.CronJob{
		TypeMeta: metaV1.TypeMeta{
			APIVersion: "batch.tutorial.kubebuilder.io/v1",
			Kind:       "CronJob",
		},
		ObjectMeta: metaV1.ObjectMeta{
			Name:      cronjobName,
			Namespace: namespace,
		},
		Spec: cronjobV1.CronJobSpec{
			Schedule: "1 * * * *",
			JobTemplate: batchV1.JobTemplateSpec{
				Spec: batchV1.JobSpec{
					// For simplicity, we only fill out the required fields.
					Template: coreV1.PodTemplateSpec{
						Spec: coreV1.PodSpec{
							// For simplicity, we only fill out the required fields.
							Containers: []coreV1.Container{
								{
									Name:  "test-container",
									Image: "test-image",
								},
							},
							RestartPolicy: coreV1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}
)

func TestCron(t *testing.T) {
	podCreationSig := make(chan *coreV1.Pod)

	feature := features.New("Cronjob Controller")

	// Use feature.Setup to define pre-test configuration
	feature.Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		client := cfg.Client()

		// setup watcher for creation of job v1.Pods by cronjob controller
		if err := client.Resources(namespace).Watch(&coreV1.PodList{}).WithAddFunc(func(obj interface{}) {
			pod := obj.(*coreV1.Pod)
			if strings.HasPrefix(pod.Name, "cronjob-controller") {
				podCreationSig <- pod
			}
		}).Start(ctx); err != nil {
			t.Fatal(err)
		}
		return ctx
	})

	// Assessment to check for CRD in cluster
	feature.Assess("CRD installed", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		client := cfg.Client()
		apiextensionsV1.AddToScheme(client.Resources().GetScheme())
		name := "cronjobs.batch.tutorial.kubebuilder.io"
		var crd apiextensionsV1.CustomResourceDefinition
		if err := client.Resources().Get(ctx, name, "", &crd); err != nil {
			t.Fatalf("CRD not found: %s", err)
		}

		if crd.Spec.Group != "batch.tutorial.kubebuilder.io" {
			t.Fatalf("Cronjob CRD has unexpected group: %s", crd.Spec.Group)
		}
		return ctx
	})

	// Assessment for Cronjo creation
	feature.Assess("Cronjob creation", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		client := cfg.Client()
		cronjobV1.AddToScheme(client.Resources(namespace).GetScheme())

		if err := client.Resources().Create(ctx, cronjob); err != nil {
			t.Fatalf("Failed to create cronjob: %s", err)
		}
		// wait for resource to be created
		if err := wait.For(
			conditions.New(client.Resources()).ResourceMatch(cronjob, func(object k8s.Object) bool {
				return true
			}),
			wait.WithTimeout(3*time.Minute),
			wait.WithInterval(30*time.Second),
		); err != nil {
			t.Fatal(err)
		}

		return ctx
	})

	// Assessment to check for cronjob deployment
	feature.Assess("Cronjob deployed Ok", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		client := cfg.Client()
		var cronjobLive cronjobV1.CronJob
		if err := client.Resources().Get(ctx, cronjobName, namespace, &cronjobLive); err != nil {
			t.Fatal(err)
		}
		if cronjobLive.Spec.Schedule != cronjob.Spec.Schedule {
			t.Fatalf("Expecting cronjob schedule %s, got %s", cronjob.Spec.Schedule, cronjobLive.Spec.Schedule)
		}
		return ctx
	})

	// Assessment to check for the creation of the pod for the scheduled job
	feature.Assess("Pod job created", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		select {
		case <-time.After(30 * time.Second):
			t.Error("Timed out wating for job pod creation by cronjob contoller")
		case pod := <-podCreationSig:
			t.Log("Pod created by cronjob-controller")
			refname := pod.GetOwnerReferences()[0].Name
			if !strings.HasPrefix(refname, cronjobName) {
				t.Fatalf("Job pod has unexpected owner ref: %#v", refname)
			}
		}
		return ctx
	})

	// Use the feature.Teardown to clean up
	feature.Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
		close(podCreationSig)
		return ctx
	})

	// submit the feature to be tested
	testEnv.Test(t, feature.Feature())
}
