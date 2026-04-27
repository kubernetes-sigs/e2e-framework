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
	"context"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	klientcel "sigs.k8s.io/e2e-framework/cel"
	celdecoder "sigs.k8s.io/e2e-framework/cel/decoder"
	celfeature "sigs.k8s.io/e2e-framework/cel/feature"
	"sigs.k8s.io/e2e-framework/cel/policy"
	celwait "sigs.k8s.io/e2e-framework/cel/wait"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// Manifest used by the YAML-assertion step. Multi-document to exercise the
// streaming path that a typical operator author would apply to their
// rendered Helm or kustomize output.
const manifestYAML = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: cel-cfg
  namespace: cel-ns
data:
  greeting: hello
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: cel-sa
  namespace: cel-ns
`

// TestCELAssertions demonstrates the four use cases the cel package
// is designed around:
//
//  1. One-line assertion against a live object (feature.AssertObject).
//  2. Offline ValidatingAdmissionPolicy evaluation (feature.AssertPolicyOnObject).
//  3. wait.For backed by a CEL condition (celwait.Match).
//  4. CEL assertions against decoded YAML manifests (celdecoder.AssertYAMLAll).
func TestCELAssertions(t *testing.T) {
	ev, err := klientcel.NewEvaluator()
	if err != nil {
		t.Fatal(err)
	}

	// A small policy with the shape a ValidatingAdmissionPolicy would ship.
	replicasPolicy := policy.Policy{
		Name: "deployment-replicas",
		Validations: []policy.Validation{
			{Expression: "object.spec.replicas >= 1", Message: "replicas must be at least 1"},
			{Expression: "object.spec.replicas <= 100", Message: "replicas must not exceed 100"},
		},
	}

	f := features.New("deployment/cel-assertions").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			dep := newDeployment(cfg.Namespace(), "cel-demo", 2)
			if err := client.Resources().Create(ctx, dep); err != nil {
				t.Fatal(err)
			}
			// Use a CEL condition as our readiness gate — on each poll the
			// Deployment is refetched and the expression is re-evaluated.
			err = wait.For(
				celwait.Match(client.Resources(), ev, &appsv1.Deployment{}, "cel-demo", cfg.Namespace(),
					"object.status.readyReplicas == object.spec.replicas"),
				wait.WithTimeout(2*time.Minute),
			)
			if err != nil {
				t.Fatal(err)
			}
			return ctx
		}).
		Assess("live object passes single CEL assertion",
			celfeature.AssertObject(ev,
				"object.status.readyReplicas == object.spec.replicas",
				&appsv1.Deployment{}, "cel-demo"),
		).
		Assess("live object passes VAP-shaped policy offline",
			celfeature.AssertPolicyOnObject(ev, replicasPolicy,
				&appsv1.Deployment{}, "cel-demo"),
		).
		Assess("live object has at least one replica",
			celfeature.AssertObject(ev,
				"object.spec.replicas >= 1",
				&appsv1.Deployment{}, "cel-demo"),
		).
		Assess("every object in a YAML manifest has a namespace",
			func(ctx context.Context, t *testing.T, _ *envconf.Config) context.Context {
				if err := celdecoder.AssertYAMLAll(ctx, ev,
					`object.metadata.namespace != ""`,
					strings.NewReader(manifestYAML),
				); err != nil {
					t.Fatal(err)
				}
				return ctx
			},
		).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			_ = client.Resources().Delete(ctx, newDeployment(cfg.Namespace(), "cel-demo", 2))
			return ctx
		}).
		Feature()

	testenv.Test(t, f)
}

func newDeployment(namespace, name string, replicas int32) *appsv1.Deployment {
	labels := map[string]string{"app": "cel-example"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx"},
					},
				},
			},
		},
	}
}
