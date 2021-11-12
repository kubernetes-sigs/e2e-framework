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
	"fmt"
	v1 "k8s.io/api/core/v1"
	"testing"
	"time"
)

func TestPodRunning(t *testing.T) {
	pod, err := createPod("p1", waitHelper.resources)
	if err != nil {
		t.Error("failed to create test resource pod", err)
	}
	err = waitHelper.For(waitHelper.PodRunning(pod))
	if err != nil {
		t.Error("failed to wait for pod to reach running condition", err)
	}
}

func TestPodPhaseMatch(t *testing.T) {
	pod, err := createPod("p2", waitHelper.resources)
	if err != nil {
		t.Error("failed to create test resource pod", err)
	}
	err = waitHelper.For(waitHelper.PodPhaseMatch(pod, v1.PodRunning))
	if err != nil {
		t.Error("failed to wait for pod to reach Running condition", err)
	}
}

func TestPodRunningBySelector(t *testing.T) {
	_, err := createPod("p3", waitHelper.resources)
	if err != nil {
		t.Error("failed to create test resource pod", err)
	}
	err = waitHelper.ForWithIntervalAndTimeout(2 * time.Second, 5 * time.Minute, waitHelper.PodRunningBySelector(fmt.Sprintf("app=p3")))
	if err != nil {
		t.Error("failed to wait for pod to reach Ready condition", err)
	}
}

