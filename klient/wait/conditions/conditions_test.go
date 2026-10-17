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

package conditions_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachinerywait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
)

func TestPodConditionsRespectContext(t *testing.T) {
	for _, tc := range []struct {
		name      string
		status    int
		body      string
		condition func(*conditions.Condition, *corev1.Pod) apimachinerywait.ConditionWithContextFunc
	}{
		{
			name:   "ResourceDeleted",
			status: http.StatusNotFound,
			body:   `{"kind":"Status","apiVersion":"v1","status":"Failure","reason":"NotFound","code":404}`,
			condition: func(c *conditions.Condition, pod *corev1.Pod) apimachinerywait.ConditionWithContextFunc {
				return c.ResourceDeleted(pod)
			},
		},
		{
			name:   "PodPhaseMatch",
			status: http.StatusOK,
			body:   `{"kind":"Pod","apiVersion":"v1","metadata":{"name":"test","namespace":"default"},"status":{"phase":"Running"}}`,
			condition: func(c *conditions.Condition, pod *corev1.Pod) apimachinerywait.ConditionWithContextFunc {
				return c.PodPhaseMatch(pod, corev1.PodRunning)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case "/api":
					_, _ = fmt.Fprint(w, `{"kind":"APIVersions","versions":["v1"]}`)
				case "/apis":
					_, _ = fmt.Fprint(w, `{"kind":"APIGroupList","groups":[]}`)
				case "/api/v1":
					_, _ = fmt.Fprint(w, `{"kind":"APIResourceList","groupVersion":"v1","resources":[{"name":"pods","namespaced":true,"kind":"Pod","verbs":["get"]}]}`)
				case "/api/v1/namespaces/default/pods/test":
					w.WriteHeader(tc.status)
					_, _ = fmt.Fprint(w, tc.body)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()

			res, err := resources.New(&rest.Config{Host: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "test", Namespace: "default"}}
			condition := tc.condition(conditions.New(res), pod)

			done, err := condition(context.Background())
			if err != nil || !done {
				t.Fatalf("active context: got (%v, %v), want (true, nil)", done, err)
			}

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			done, err = condition(ctx)
			if done || !errors.Is(err, context.Canceled) {
				t.Fatalf("canceled context: got (%v, %v), want (false, context.Canceled)", done, err)
			}
		})
	}
}
