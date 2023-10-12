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

package cilium

import (
	"bytes"
	"context"
	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"

	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNetworkPolicies(t *testing.T) {
	containerName := "nginx"
	podName := ""
	feature := features.New("FQDN whitelisting").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			// Setup cluster network policies to only allow whitelisted traffic
			r, err := resources.New(config.Client().RESTConfig())
			if err != nil {
				t.Fatal(err)
			}
			err = decoder.ApplyWithManifestDir(ctx, r, "./templates", "*", []resources.CreateOption{})
			if err != nil {
				t.Fatal(err)
			}
			// Create deployment
			deploymentName := "test-deployment"

			deployment := newDeployment(config.Namespace(), deploymentName, 1, containerName)
			client, err := config.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err = client.Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			err = wait.For(conditions.New(client.Resources()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, corev1.ConditionTrue), wait.WithTimeout(time.Minute*5))
			if err != nil {
				t.Fatal(err)
			}

			pods := &corev1.PodList{}
			err = client.Resources(config.Namespace()).List(context.TODO(), pods)
			if err != nil || pods.Items == nil {
				t.Error("error while getting pods", err)
			}
			podName = pods.Items[0].Name

			return ctx
		}).
		Assess("Nginx pod can call github api", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			client, err := c.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			var stdout, stderr bytes.Buffer

			// Curl Github API
			command := []string{"curl", "-I", "https://api.github.com"}
			if err := client.Resources().ExecInPod(context.TODO(), c.Namespace(), podName, containerName, command, &stdout, &stderr); err != nil {
				t.Log(stderr.String())
				t.Fatal(err)
			}

			httpStatus := strings.Split(stdout.String(), "\n")[0]
			if !strings.Contains(httpStatus, "200") {
				t.Fatal("Couldn't connect to api.github.com")
			}

			return ctx
		}).
		Assess("Nginx pod can call github api", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			client, err := c.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer

			// Curl Wikipedia
			command := []string{"curl", "-I", "-m", "1", "https://www.wikipedia.org"}
			if err := client.Resources().ExecInPod(context.TODO(), c.Namespace(), podName, containerName, command, &stdout, &stderr); err == nil {
				t.Log(stderr.String())
				t.Fatal(err)
			}

			httpStatus := strings.Split(stdout.String(), "\n")[0]
			if strings.Contains(httpStatus, "200") {
				t.Fatal("It should not connect to wikipedia")
			}

			return ctx
		}).Feature()

	_ = testEnv.Test(t, feature)

}

func newDeployment(namespace string, name string, replicas int32, containerName string) *appsv1.Deployment {
	labels := map[string]string{"app": "nginx"}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: containerName, Image: "nginx"}}},
			},
		},
	}
}
