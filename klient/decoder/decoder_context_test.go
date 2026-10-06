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

package decoder_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sigs.k8s.io/e2e-framework/klient/decoder"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

func TestDecodeURLCancellation(t *testing.T) {
	for _, tc := range []struct {
		name      string
		flushBody bool
	}{
		{name: "waiting for headers"},
		{name: "reading the body", flushBody: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			started := make(chan struct{})
			decoded := make(chan struct{})
			release := make(chan struct{})
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tc.flushBody {
					if _, err := fmt.Fprint(w, "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cancel-test\n---\n"); err != nil {
						t.Errorf("write manifest: %v", err)
						return
					}
					if err := http.NewResponseController(w).Flush(); err != nil {
						t.Errorf("flush manifest: %v", err)
						return
					}
				}
				close(started)
				select {
				case <-r.Context().Done():
				case <-release:
				}
			}))
			defer server.Close()
			defer close(release)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			result := make(chan error, 1)
			go func() {
				handlerCalls := 0
				result <- decoder.DecodeURL(ctx, server.URL, func(context.Context, k8s.Object) error {
					handlerCalls++
					if !tc.flushBody || handlerCalls > 1 {
						t.Error("handler called unexpectedly")
					} else {
						close(decoded)
					}
					return nil
				})
			}()
			ready := started
			if tc.flushBody {
				ready = decoded
			}
			select {
			case <-ready:
			case <-time.After(5 * time.Second):
				t.Fatal("download did not reach the expected cancellation point")
			}
			cancel()
			select {
			case err := <-result:
				if !errors.Is(err, context.Canceled) {
					t.Errorf("expected context cancellation, got %v", err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("DecodeURL did not return after cancellation")
			}
		})
	}
}
