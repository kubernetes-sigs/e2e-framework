/*
Copyright 2021 The Kubernetes Authors.

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

package res

import (
	"context"
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCreate(t *testing.T) {
	res := Res(cfg)
	if res == nil {
		t.Errorf("config is nill")
	}

	// create a namespace
	err := res.Create(context.TODO(), namespace)
	if err != nil {
		t.Error("error while creating namespace", err)
	}

	var nsObj corev1.Namespace
	err = res.Get(context.TODO(), namespace.Name, namespace.Name, &nsObj)
	if err != nil {
		t.Error("error while getting the namespace", err)
	}

	if nsObj.Name != namespace.Name {
		t.Error("namespace name mismatch, expected : ", namespace.Name, "obtained :", nsObj.Name)
	}
}

func TestRes(t *testing.T) {
	res := Res(cfg)
	if res == nil {
		t.Errorf("config is nill")
	}

	err := res.Create(context.TODO(), dep)
	if err != nil {
		t.Error("error while creating deployment", err)
	}

	actual, err := clientset.AppsV1().Deployments(namespace.Name).Get(context.TODO(), dep.Name, metav1.GetOptions{})
	if err != nil {
		t.Error("error while getting the deployment details", err)
	}

	if actual == dep {
		fmt.Println("deployment found", dep.Name)
	}

	var depObj appsv1.Deployment
	err = res.Get(context.TODO(), dep.Name, namespace.Name, &depObj)
	if err != nil {
		t.Error("error while getting the deployment", err)
	}

	if depObj.Name != dep.Name {
		t.Error("deployment name mismatch, expected : ", dep.Name, "obtained :", depObj.Name)
	}
}

func TestList(t *testing.T) {
	res := Res(cfg)
	if res == nil {
		t.Errorf("config is nill")
	}

	deps := &appsv1.DeploymentList{}
	err := res.List(context.TODO(), deps)
	if err != nil {
		t.Error("error while getting the deployment", err)
	}

	if deps.Items == nil {
		t.Error("error while getting the list of deployments", err)
	}

	hasDep := false
	for _, item := range deps.Items {
		if item.Name == dep.Name && item.Namespace == dep.Namespace {
			hasDep = true
			break
		}
	}

	if !hasDep {
		t.Error("there are no deployment exist", hasDep)
	}
}
