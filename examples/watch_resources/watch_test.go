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

package watch_resources

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/k8s/watcher"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestWatchForResources(t *testing.T) {
	watchFeature := features.New("test watcher").WithLabel("env", "dev").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			cl, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			dep := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "watch-dep", Namespace: cfg.Namespace()},
			}

			// watch for the deployment and triger action based on the event recieved.
			cl.Resources().Watch(&appsv1.DeploymentList{}, resources.WithFieldSelector(labels.FormatLabels(map[string]string{"metadata.name": dep.Name}))).
				WithAddFunc(onAdd).WithDeleteFunc(onDelete).Start(ctx)

			return ctx
		}).
		Assess("create watch deployment", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// create a deployment
			deployment := newDeployment(cfg.Namespace(), "watch-dep", 1)
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			return context.WithValue(ctx, "test-dep", deployment)
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			depl := ctx.Value("test-dep").(*appsv1.Deployment)
			if err := client.Resources().Delete(ctx, depl); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testenv.Test(t, watchFeature)

}

// TestWatchForResourcesWithStop() demonstartes how to start and stop the watcher
var w *watcher.EventHandlerFuncs

func TestWatchForResourcesWithStop(t *testing.T) {
	watchFeature := features.New("test watcher with stop").WithLabel("env", "prod").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			cl, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			dep := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "watchnstop-dep", Namespace: cfg.Namespace()},
			}

			// watch for the deployment and triger action based on the event recieved.
			w = cl.Resources().Watch(&appsv1.DeploymentList{}, resources.WithFieldSelector(labels.FormatLabels(map[string]string{"metadata.name": dep.Name}))).
				WithAddFunc(onAdd).WithDeleteFunc(onDelete)

			err = w.Start(ctx)
			if err != nil {
				t.Error(err)
			}

			return ctx
		}).
		Assess("create watch deployment", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// create a deployment
			deployment := newDeployment(cfg.Namespace(), "watchnstop-dep", 1)
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}
			return context.WithValue(ctx, "stop-dep", deployment)
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			depl := ctx.Value("stop-dep").(*appsv1.Deployment)
			if err := client.Resources().Delete(ctx, depl); err != nil {
				t.Fatal(err)
			}

			w.Stop()

			return ctx
		}).Feature()

	testenv.Test(t, watchFeature)

}

func TestWatchForResources1(t *testing.T) {
	addWait := make(chan struct{})
	delWait := make(chan struct{})
	onAddfunc := func(obj interface{}) {
		dep := obj.(*appsv1.Deployment)
		depName := dep.GetName()
		if depName == "demo-app" {
			addWait <- struct{}{}
		}
	}

	onDelfunc := func(obj interface{}) {
		dep := obj.(*appsv1.Deployment)
		depName := dep.GetName()
		if depName == "demo-app" {
			delWait <- struct{}{}
		}
	}

	watchFeature := features.New("test watcher with callback methods").WithLabel("env", "prod").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			cl, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}

			dep := appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "demo-app", Namespace: cfg.Namespace()},
			}

			// watch for the deployment and triger action based on the event recieved.
			w = cl.Resources().Watch(&appsv1.DeploymentList{}, resources.WithFieldSelector(labels.FormatLabels(map[string]string{"metadata.name": dep.Name}))).
				WithAddFunc(onAddfunc).WithDeleteFunc(onDelfunc)

			err = w.Start(ctx)
			if err != nil {
				t.Error(err)
			}

			return ctx
		}).
		Assess("create watch deployment", func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			// create a deployment
			deployment := newDeployment(cfg.Namespace(), "demo-app", 1)
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			if err := client.Resources().Create(ctx, deployment); err != nil {
				t.Fatal(err)
			}

			// After creation, wait for signal
			select {
			case <-time.After(3 * time.Second):
				t.Error("Add callback not called")
			case <-addWait:
				klog.InfoS("recieved signal, closing add channel")
				close(addWait)
			}

			return context.WithValue(ctx, "demo-app", deployment)
		}).
		Teardown(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {
			client, err := cfg.NewClient()
			if err != nil {
				t.Fatal(err)
			}
			depl := ctx.Value("demo-app").(*appsv1.Deployment)
			if err := client.Resources().Delete(ctx, depl); err != nil {
				t.Fatal(err)
			}

			// After deletion, wait for signal
			select {
			case <-time.After(3 * time.Second):
				t.Error("Delete callback not called")
			case <-delWait:
				klog.InfoS("recieved signal, closing delete channel")
				close(delWait)
			}

			w.Stop()

			return ctx
		}).Feature()
	testenv.Test(t, watchFeature)
}

func newDeployment(namespace string, name string, replicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace, Labels: map[string]string{"app": "watch-for-resources"}},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "watch-for-resources"},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "watch-for-resources"}},
				Spec:       v1.PodSpec{Containers: []v1.Container{{Name: "nginx", Image: "nginx"}}},
			},
		},
	}
}

// onAdd is the function executed when the kubernetes watch notifies the
// presence of a new kubernetes deployment in the cluster
func onAdd(obj interface{}) {
	dep := obj.(*appsv1.Deployment)
	depName := dep.GetName()
	if depName == "watch-dep" || depName == "watchnstop-dep" {
		klog.InfoS("Deployment name matches with actual name!")
	}
}

// onDelete is the function executed when the kubernetes watch notifies
// delete event on deployment
func onDelete(obj interface{}) {
	dep := obj.(*appsv1.Deployment)
	depName := dep.GetName()
	if depName == "watch-dep" || depName == "watchnstop-dep" {
		klog.InfoS("Deployment deleted successfully!")
	}
}
