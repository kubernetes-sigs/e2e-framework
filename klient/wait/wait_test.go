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

package wait

import (
	"context"
	"testing"
	"time"

	log "k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

func TestPodRunning(t *testing.T) {
	var err error
	pod := createPod("p1", t)
	err = For(conditions.New(getResourceManager()).PodRunning(pod), WithImmediate())
	if err != nil {
		t.Error("failed to wait for pod to reach running condition", err)
	}
}

func TestPodPhaseMatch(t *testing.T) {
	var err error
	pod := createPod("p2", t)
	err = For(conditions.New(getResourceManager()).PodPhaseMatch(pod, v1.PodRunning))
	if err != nil {
		t.Error("failed to wait for pod to reach Running condition", err)
	}
}

func TestPodReady(t *testing.T) {
	var err error
	pod := createPod("p3", t)
	err = For(conditions.New(getResourceManager()).PodReady(pod), WithInterval(2*time.Second))
	if err != nil {
		t.Error("failed to wait for pod to reach Ready condition", err)
	}
}

func TestContainersReady(t *testing.T) {
	var err error
	pod := createPod("p4", t)
	err = For(conditions.New(getResourceManager()).ContainersReady(pod))
	if err != nil {
		t.Error("failed to wait for containers to reach Ready condition", err)
	}
}

func TestJobCompleted(t *testing.T) {
	var err error
	job := createJob("j1", "echo", "kubernetes", t)
	err = For(conditions.New(getResourceManager()).JobCompleted(job))
	if err != nil {
		t.Error("failed waiting for job to complete", err)
	}
}

func TestJobFailed(t *testing.T) {
	var err error
	job := createJob("j2", "exit", "1", t)
	err = For(conditions.New(getResourceManager()).JobFailed(job))
	if err != nil {
		t.Error("failed waiting for job to fail", err)
	}
}

func TestResourceDeleted(t *testing.T) {
	var err error
	pod := createPod("p5", t)
	err = For(conditions.New(getResourceManager()).ContainersReady(pod))
	if err != nil {
		t.Error("failed to wait for containers to reach Ready condition", err)
	}
	go func() {
		err := getResourceManager().Delete(context.TODO(), pod)
		if err != nil {
			log.ErrorS(err, "ran into an error trying to delete the resource")
		}
	}()
	err = For(conditions.New(getResourceManager()).ResourceDeleted(pod), WithInterval(2*time.Second), WithTimeout(7*time.Minute), WithImmediate())
	if err != nil {
		t.Error("failed waiting for pod resource to be deleted", err)
	}
}

func TestResourceScaled(t *testing.T) {
	var err error
	deployment := createDeployment("d1", 2, t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = For(conditions.New(getResourceManager()).ResourceScaled(deployment, func(object k8s.Object) int32 {
		return object.(*appsv1.Deployment).Status.ReadyReplicas
	}, 2), WithContext(ctx))
	if err != nil {
		t.Error("failed waiting for resource to be scaled", err)
	}
	log.Info("Done")
}

func TestDeploymentConditionMatch(t *testing.T) {
	var err error
	deployment := createDeployment("d2", 3, t)
	err = For(conditions.New(getResourceManager()).DeploymentConditionMatch(deployment, appsv1.DeploymentAvailable, v1.ConditionTrue))
	if err != nil {
		t.Error("failed waiting for deployment to become available", err)
	}
	log.Info("Done")
}

func TestResourceListN(t *testing.T) {
	var err error
	createDeployment("d3", 4, t)
	pods := &v1.PodList{}
	err = For(conditions.New(getResourceManager()).ResourceListN(pods, 4, resources.WithLabelSelector(labels.FormatLabels(map[string]string{"app": "d3"}))))
	if err != nil {
		t.Error("failed waiting for deployment pods to be created", err)
	}
	log.Info("Done")
}

func TestResourceListMatchN(t *testing.T) {
	var err error
	createDeployment("d4", 5, t)
	pods := &v1.PodList{}
	err = For(conditions.New(getResourceManager()).ResourceListMatchN(pods, 5, func(object k8s.Object) bool {
		for _, c := range object.(*v1.Pod).Spec.Containers {
			if c.Image == "nginx" {
				return true
			}
		}
		return false
	}, resources.WithLabelSelector(labels.FormatLabels(map[string]string{"app": "d4"}))))
	if err != nil {
		t.Error("failed waiting for deployment pods with nginx containers to be created", err)
	}
	log.Info("Done")
}

func TestResourcesMatch(t *testing.T) {
	var err error
	go func() {
		createPod("p6", t)
		createPod("p7", t)
		createPod("p8", t)
	}()
	pods := &v1.PodList{
		Items: []v1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "p6", Namespace: namespace}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p7", Namespace: namespace}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p8", Namespace: namespace}},
		},
	}
	err = For(conditions.New(getResourceManager()).ResourcesMatch(pods, func(object k8s.Object) bool {
		return object.(*v1.Pod).Status.Phase == v1.PodRunning
	}))
	if err != nil {
		t.Error("failed waiting for deployment pods to start running", err)
	}
	log.Info("Done")
}

func TestResourcesFound(t *testing.T) {
	var err error
	go func() {
		createPod("p9", t)
		createPod("p10", t)
		createPod("p11", t)
	}()
	pods := &v1.PodList{
		Items: []v1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "p9", Namespace: namespace}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p10", Namespace: namespace}},
			{ObjectMeta: metav1.ObjectMeta{Name: "p11", Namespace: namespace}},
		},
	}
	err = For(conditions.New(getResourceManager()).ResourcesFound(pods))
	if err != nil {
		t.Error("failed waiting for deployment pods to be created", err)
	}
	log.Info("Done")
}

func TestResourcesDeleted(t *testing.T) {
	var err error
	deployment := createDeployment("d5", 1, t)
	pods := &v1.PodList{}
	err = For(conditions.New(getResourceManager()).ResourceListN(pods, 1, resources.WithLabelSelector(labels.FormatLabels(map[string]string{"app": "d5"}))))
	if err != nil {
		t.Error("failed waiting for deployment pods to be created", err)
	}
	err = getResourceManager().Delete(context.Background(), deployment)
	if err != nil {
		t.Error("failed to delete deployment due to an error", err)
	}
	err = For(conditions.New(getResourceManager()).ResourcesDeleted(pods))
	if err != nil {
		t.Error("failed waiting for pods to be deleted", err)
	}
	log.Info("Done")
}

func TestResourceMatch(t *testing.T) {
	var err error
	deployment := createDeployment("d6", 2, t)
	err = For(conditions.New(getResourceManager()).ResourceMatch(deployment, func(object k8s.Object) bool {
		d, ok := object.(*appsv1.Deployment)
		if !ok {
			t.Fatalf("unexpected type %T in list, does not satisfy *appsv1.Deployment", object)
		}
		return d.Status.AvailableReplicas == 2 && d.Status.ReadyReplicas == 2
	}))
	if err != nil {
		t.Error("failed waiting for deployment replicas", err)
	}
	log.Info("Done")
}
