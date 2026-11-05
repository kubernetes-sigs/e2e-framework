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

package watcher

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

func TestHandleEventsReturnsWhenWatchIsStopped(t *testing.T) {
	fw := watch.NewFake()
	var added []string
	e := (&EventHandlerFuncs{}).WithAddFunc(func(obj interface{}) {
		added = append(added, obj.(*corev1.Pod).Name)
	})

	done := make(chan struct{})
	go func() {
		e.handleEvents(context.Background(), fw.ResultChan())
		close(done)
	}()

	fw.Add(&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod-1"}})
	fw.Stop()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handleEvents did not return after the watch was stopped")
	}
	if len(added) != 1 || added[0] != "pod-1" {
		t.Fatalf("expected the add func to be called once with pod-1, got %v", added)
	}
}

func TestHandleEventsReturnsWhenContextIsDone(t *testing.T) {
	fw := watch.NewFake()
	defer fw.Stop()
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		(&EventHandlerFuncs{}).handleEvents(ctx, fw.ResultChan())
		close(done)
	}()
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("handleEvents did not return after the context was cancelled")
	}
}
