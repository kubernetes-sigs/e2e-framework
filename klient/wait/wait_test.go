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
	"log"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

func TestPodRunning(t *testing.T) {
	var err error
	pod := createPod("p1", t)
	err = For(conditions.New(getResourceManager()).WithVerboseLog().PodRunning(pod), WithImmediate())
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
			log.Println("ran into an error trying to delete the resource", err)
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
	stopChan := make(chan struct{})
	err = For(conditions.New(getResourceManager()).ResourceScaled(deployment, func(object k8s.Object) int32 {
		return object.(*appsv1.Deployment).Status.ReadyReplicas
	}, 2), WithStopChannel(stopChan))
	if err != nil {
		t.Error("failed waiting for resource to be scaled", err)
	}
	log.Println("Dione")
}
