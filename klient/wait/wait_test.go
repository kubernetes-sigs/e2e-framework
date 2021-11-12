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
	v1 "k8s.io/api/core/v1"
	"testing"
)

func TestPodRunning(t *testing.T) {
	var err error
	pod := createPod("p1", waitHelper.resources, t)
	err = waitHelper.For(waitHelper.PodRunning(pod))
	if err != nil {
		t.Error("failed to wait for pod to reach running condition", err)
	}
}

func TestPodPhaseMatch(t *testing.T) {
	var err error
	pod := createPod("p2", waitHelper.resources, t)
	err = waitHelper.For(waitHelper.PodPhaseMatch(pod, v1.PodRunning))
	if err != nil {
		t.Error("failed to wait for pod to reach Running condition", err)
	}
}

func TestPodReady(t *testing.T) {
	var err error
	pod := createPod("p3", waitHelper.resources, t)
	err = waitHelper.For(waitHelper.PodReady(pod))
	if err != nil {
		t.Error("failed to wait for pod to reach Ready condition", err)
	}
}

func TestContainersReady(t *testing.T) {
	var err error
	pod := createPod("p4", waitHelper.resources, t)
	err = waitHelper.For(waitHelper.ContainersReady(pod))
	if err != nil {
		t.Error("failed to wait for containers to reach Ready condition", err)
	}
}

func TestJobCompleted(t *testing.T) {
	var err error
	job := createJob("j1", "echo", "kubernetes", waitHelper.resources, t)
	err = waitHelper.For(waitHelper.JobCompleted(job))
	if err != nil {
		t.Error("failed waiting for job to complete", err)
	}
}

func TestJobFailed(t *testing.T) {
	var err error
	job := createJob("j2", "exit", "1", waitHelper.resources, t)
	err = waitHelper.For(waitHelper.JobFailed(job))
	if err != nil {
		t.Error("failed waiting for job to fail", err)
	}
}

func TestResourceDeleted(t *testing.T) {
	var err error
	pod := createPod("p5", waitHelper.resources, t)
	err = waitHelper.For(waitHelper.ContainersReady(pod))
	if err != nil {
		t.Error("failed to wait for containers to reach Ready condition", err)
	}
	go func() {
		_ = waitHelper.resources.Delete(context.TODO(), pod)
	}()
	err = waitHelper.For(waitHelper.ResourceDeleted(pod))
	if err != nil {
		t.Error("failed waiting for pod resource to be deleted", err)
	}
}