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

package nginx

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestFluxRepoWorkflow(t *testing.T) {
	feature := features.New("Install resources by flux").
		Assess("check if deployment was successful", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			deployment := &appsv1.Deployment{
				ObjectMeta: v1.ObjectMeta{
					Name:      "hello-app",
					Namespace: c.Namespace(),
				},
				Spec: appsv1.DeploymentSpec{},
			}

			err := wait.For(conditions.New(c.Client().Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionStatus(v1.ConditionTrue)), wait.WithTimeout(time.Minute*5))
			if err != nil {
				t.Fatal("Error deployment not found", err)
			}

			return ctx
		}).Feature()

	testEnv.Test(t, feature)
}
