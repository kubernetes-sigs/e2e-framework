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

package resources_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/vladimirvivien/gexe"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	log "k8s.io/klog/v2"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources/testdata/projectExample"
)

func TestCreate(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	// create a namespace
	err = res.Create(context.TODO(), namespace)
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
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	err = res.Create(context.TODO(), dep)
	if err != nil {
		t.Error("error while creating deployment", err)
	}

	actual, err := clientset.AppsV1().Deployments(namespace.Name).Get(context.TODO(), dep.Name, metav1.GetOptions{})
	if err != nil {
		t.Error("error while getting the deployment details", err)
	}

	if actual == dep {
		log.Info("deployment found", dep.Name)
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

func TestResNoConfig(t *testing.T) {
	_, err := resources.New(nil)
	if err == nil {
		t.Error("expected error while invoking Res without k8s config")
	}
}

func TestUpdate(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	depActual := getDeployment("update-test-dep-name")

	err = res.Create(context.TODO(), depActual)
	if err != nil {
		t.Error("error while creating deployment", err)
	}

	depUpdated := depActual
	depUpdated.ObjectMeta.Labels["test-key"] = "test-val"

	err = res.Update(context.TODO(), depUpdated)
	if err != nil {
		t.Error("error while updating deployment", err)
	}

	var depObj appsv1.Deployment
	err = res.Get(context.TODO(), depUpdated.Name, namespace.Name, &depObj)
	if err != nil {
		t.Error("error while getting the deployment", err)
	}

	val, ok := depObj.Labels["test-key"]
	if !ok {
		t.Error("deployment not updated")
	} else if val != "test-val" {
		t.Error("deployment label value mismatch, expected : ", "test-val", "obtained :", val)
	}
}

func TestUpdateStatus(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	depActual := getDeployment("update-status-test-dep-name")

	err = res.Create(context.TODO(), depActual)
	if err != nil {
		t.Error("error while creating deployment", err)
	}

	depUpdated := depActual
	depUpdated.Status.Conditions = append(depUpdated.Status.Conditions,
		appsv1.DeploymentCondition{
			Type:               "UpdateStatusTest",
			Status:             corev1.ConditionTrue,
			LastUpdateTime:     metav1.NewTime(time.Now()),
			LastTransitionTime: metav1.NewTime(time.Now()),
		},
	)

	err = res.UpdateStatus(context.TODO(), depUpdated)
	if err != nil {
		t.Error("error while updating deployment status", err)
	}

	var depObj appsv1.Deployment
	err = res.Get(context.TODO(), depUpdated.Name, namespace.Name, &depObj)
	if err != nil {
		t.Error("error while getting the deployment", err)
	}

	var cond *appsv1.DeploymentCondition
	for i := range depObj.Status.Conditions {
		if depObj.Status.Conditions[i].Type == "UpdateStatusTest" {
			cond = &depObj.Status.Conditions[i]
			break
		}
	}

	if cond == nil {
		t.Error("deployment status not updated")
	} else if cond.Status != corev1.ConditionTrue {
		t.Error("deployment status value mismatch, expected : ", corev1.ConditionTrue, "obtained :", cond.Status)
	}
}

func TestDelete(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	depActual := getDeployment("delete-test-dep-name")

	err = res.Create(context.TODO(), depActual)
	if err != nil {
		t.Error("error while creating deployment", err)
	}

	err = res.Delete(context.TODO(), depActual)
	if err != nil {
		t.Error("error while deleting deployment", err)
	}
}

func TestList(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	deps := &appsv1.DeploymentList{}
	err = res.List(context.TODO(), deps)
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

func TestPatch(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]interface{}{
				"ping": "pong",
			},
		},
	})
	if err != nil {
		t.Error("error while json marshalling", err)
	}

	err = res.Patch(context.Background(), dep, k8s.Patch{PatchType: types.StrategicMergePatchType, Data: mergePatch})
	if err != nil {
		t.Error("error while patching the deployment", err)
	}

	obj := &appsv1.Deployment{}
	err = res.Get(context.Background(), dep.Name, dep.Namespace, obj)
	if err != nil {
		t.Error("error while getting patched deployment", err)
	}

	if obj.Annotations["ping"] != "pong" {
		t.Error("resource patch not applied correctly.")
	}
}

func TestPatchStatus(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	mergePatch, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"type":               "UpdateStatusTest",
					"status":             "True",
					"lastUpdateTime":     metav1.NewTime(time.Now()),
					"lastTransitionTime": metav1.NewTime(time.Now()),
				},
			},
		},
	})
	if err != nil {
		t.Error("error while json marshalling", err)
	}

	err = res.PatchStatus(context.Background(), dep, k8s.Patch{PatchType: types.StrategicMergePatchType, Data: mergePatch})
	if err != nil {
		t.Error("error while patching the deployment", err)
	}

	obj := &appsv1.Deployment{}
	err = res.Get(context.Background(), dep.Name, dep.Namespace, obj)
	if err != nil {
		t.Error("error while getting patched deployment", err)
	}

	var cond *appsv1.DeploymentCondition
	for i := range obj.Status.Conditions {
		if obj.Status.Conditions[i].Type == "UpdateStatusTest" {
			cond = &obj.Status.Conditions[i]
			break
		}
	}

	if cond == nil {
		t.Error("deployment status not updated")
	} else if cond.Status != corev1.ConditionTrue {
		t.Error("deployment status value mismatch, expected : ", corev1.ConditionTrue, "obtained :", cond.Status)
	}
}

func TestListAllPods(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	pods := &corev1.PodList{}
	err = res.List(context.TODO(), pods)
	if err != nil {
		t.Error("error while getting the deployment", err)
	}

	if pods.Items == nil {
		t.Error("error while getting the list of deployments", err)
	}

	t.Logf("pod list contains %d pods", len(pods.Items))
}

func TestGetCRDs(t *testing.T) {
	res, err := resources.New(cfg)
	if err != nil {
		t.Fatalf("Error creating new resources object: %v", err)
	}

	// Register type for the API server.
	e := gexe.New()
	p := e.RunProc(`kubectl apply -f ./testdata/projectExample/resourcedefinition.yaml`)
	if p.Err() != nil {
		t.Fatalf("Failed to register CRD: %v %v", p.Err(), p.Result())
	}
	// Sometimes CRDs need just a bit of time before being ready to use.
	time.Sleep(5 * time.Second)

	// Create one
	p = e.RunProc(`kubectl apply -f ./testdata/projectExample/project.yaml`)
	if p.Err() != nil {
		t.Fatalf("Failed to create a CRD via yaml: %v %v", p.Err(), p.Result())
	}

	// See that we can't list it because we don't know the type.
	ps := &projectExample.ProjectList{}
	err = res.List(context.TODO(), ps)
	if err == nil {
		t.Error("Expected error while listing custom resources before adding it to scheme, but got none")
	}

	// Register type with klient.
	if err := projectExample.AddToScheme(res.GetScheme()); err != nil {
		t.Fatalf("Failed to add to resource scheme: %v", err)
	}

	// See that we can after registering it.
	err = res.List(context.TODO(), ps)
	if err != nil {
		t.Error("error while listing custom resources", err)
	}
}

func TestExecInPod(t *testing.T) {
	res, err := resources.New(cfg)
	containerName := "nginx"
	if err != nil {
		t.Fatalf("Error initiating runtime controller: %v", err)
	}
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "test-exec-ns"}}
	err = res.Create(context.TODO(), namespace)
	if err != nil {
		t.Fatalf("Error while creating namespace resource: %v", err)
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-exec", Namespace: namespace.Name},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: containerName, Image: "nginx"}}},
	}

	err = res.Create(context.TODO(), pod)
	if err != nil {
		t.Error("Error while creating pod resource", err)
	}

	pods := &corev1.PodList{}
	err = res.List(context.TODO(), pods)
	if err != nil {
		t.Error("error while getting pods", err)
	}
	if pods.Items == nil {
		t.Error("error while getting the list of pods", err)
	}
	var stdout, stderr bytes.Buffer

	addWait := make(chan struct{})
	onAddfunc := func(obj interface{}) {
		addWait <- struct{}{}
	}
	w := res.Watch(&corev1.PodList{}, resources.WithFieldSelector(labels.FormatLabels(
		map[string]string{
			"metadata.name":      pod.Name,
			"metadata.namespace": namespace.Name,
			"status.phase":       "Running",
		}))).
		WithAddFunc(onAddfunc)

	if err = w.Start(ctx); err != nil {
		t.Fatal(err)
	}

	select {
	case <-time.After(300 * time.Second):
		t.Error("Add callback not called")
	case <-addWait:
		close(addWait)
	}

	if err := res.ExecInPod(context.TODO(), namespace.Name, pod.Name, containerName, []string{"printenv"}, &stdout, &stderr); err != nil {
		t.Log(stderr.String())
		t.Fatal(err)
	}

	hostName := "HOSTNAME=" + pod.Name
	if !strings.Contains(stdout.String(), hostName) {
		t.Fatal("Couldn't find proper env")
	}
}
