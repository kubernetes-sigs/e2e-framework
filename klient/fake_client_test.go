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

package klient_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

func TestNewFromCRClient_WithFakeClient(t *testing.T) {
	// Create a fake controller-runtime client
	fakeClient := fake.NewClientBuilder().
		WithObjects(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
			Data: map[string]string{
				"key": "value",
			},
		}).
		Build()

	// Create a klient.Client from the fake client
	klClient := klient.NewFromCRClient(fakeClient)

	// Test that RESTConfig returns nil
	if klClient.RESTConfig() != nil {
		t.Error("expected RESTConfig() to return nil for Client created via NewFromCRClient")
	}

	// Test that Resources() works and can perform CRUD operations
	res := klClient.Resources("default")

	// Test Get
	cm := &corev1.ConfigMap{}
	err := res.Get(context.TODO(), "test-cm", "default", cm)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if cm.Data["key"] != "value" {
		t.Errorf("expected data key=value, got %v", cm.Data)
	}

	// Test Create
	newCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-cm",
			Namespace: "default",
		},
		Data: map[string]string{
			"newkey": "newvalue",
		},
	}
	err = res.Create(context.TODO(), newCM)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Test Update
	newCM.Data["newkey"] = "updated"
	err = res.Update(context.TODO(), newCM)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	updatedCM := &corev1.ConfigMap{}
	err = res.Get(context.TODO(), "new-cm", "default", updatedCM)
	if err != nil {
		t.Fatalf("Get updated ConfigMap failed: %v", err)
	}
	if updatedCM.Data["newkey"] != "updated" {
		t.Errorf("expected newkey=updated, got %v", updatedCM.Data)
	}

	// Test Delete
	err = res.Delete(context.TODO(), updatedCM)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	deletedCM := &corev1.ConfigMap{}
	err = res.Get(context.TODO(), "new-cm", "default", deletedCM)
	if err == nil {
		t.Error("expected error after delete, but got nil")
	}
}

func TestNewFromCRClient_ResourcesWithoutNamespace(t *testing.T) {
	fakeClient := fake.NewClientBuilder().
		WithObjects(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "kube-system",
			},
		}).
		Build()

	klClient := klient.NewFromCRClient(fakeClient)

	// Test Resources() without namespace argument
	res := klClient.Resources()

	cm := &corev1.ConfigMap{}
	err := res.Get(context.TODO(), "test-cm", "kube-system", cm)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if cm.Name != "test-cm" {
		t.Errorf("expected ConfigMap name test-cm, got %s", cm.Name)
	}
}

func TestNewFromCRClient_MultipleNamespaces(t *testing.T) {
	fakeClient := fake.NewClientBuilder().
		WithObjects(
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cm1",
					Namespace: "ns1",
				},
			},
			&corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "cm2",
					Namespace: "ns2",
				},
			},
		).
		Build()

	klClient := klient.NewFromCRClient(fakeClient)

	// Test with first namespace
	res1 := klClient.Resources("ns1")
	cm1 := &corev1.ConfigMap{}
	err := res1.Get(context.TODO(), "cm1", "ns1", cm1)
	if err != nil {
		t.Fatalf("Get from ns1 failed: %v", err)
	}

	// Test with second namespace
	res2 := klClient.Resources("ns2")
	cm2 := &corev1.ConfigMap{}
	err = res2.Get(context.TODO(), "cm2", "ns2", cm2)
	if err != nil {
		t.Fatalf("Get from ns2 failed: %v", err)
	}
}

func TestNewFromCRClient_GetConfigReturnsNil(t *testing.T) {
	fakeClient := fake.NewClientBuilder().Build()
	klClient := klient.NewFromCRClient(fakeClient)

	if klClient.Resources().GetConfig() != nil {
		t.Error("expected GetConfig() to return nil for Resources created via NewFromCRClient")
	}
}

func TestNewFromCRClient_ExecInPodReturnsError(t *testing.T) {
	fakeClient := fake.NewClientBuilder().Build()
	klClient := klient.NewFromCRClient(fakeClient)
	res := klClient.Resources("default")

	var stdout, stderr bytes.Buffer
	err := res.ExecInPod(context.TODO(), "default", "test-pod", "container", []string{"echo", "test"}, &stdout, &stderr)
	if err == nil {
		t.Fatal("expected ExecInPod to return an error, got nil")
	}
	if !strings.Contains(err.Error(), "REST config") {
		t.Errorf("expected error about REST config, got: %v", err)
	}
}

func TestNewFromCRClient_GetControllerRuntimeClient(t *testing.T) {
	fakeClient := fake.NewClientBuilder().
		WithObjects(&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
		}).
		Build()

	klClient := klient.NewFromCRClient(fakeClient)
	crClient := klClient.Resources().GetControllerRuntimeClient()
	if crClient == nil {
		t.Fatal("expected GetControllerRuntimeClient() to return non-nil")
	}

	cm := &corev1.ConfigMap{}
	err := crClient.Get(context.TODO(), cr.ObjectKey{Name: "test-cm", Namespace: "default"}, cm)
	if err != nil {
		t.Fatalf("Get via returned cr.Client failed: %v", err)
	}
	if cm.Name != "test-cm" {
		t.Errorf("expected name test-cm, got %s", cm.Name)
	}
}

func TestNewFromCRClient_WatchStartReturnsError(t *testing.T) {
	fakeClient := fake.NewClientBuilder().Build()
	res := resources.NewFromClient(fakeClient)

	handler := res.Watch(&corev1.ConfigMapList{})
	err := handler.Start(context.TODO())
	if err == nil {
		t.Fatal("expected Watch.Start to return an error, got nil")
	}
	if !strings.Contains(err.Error(), "REST config") {
		t.Errorf("expected error about REST config, got: %v", err)
	}
}
