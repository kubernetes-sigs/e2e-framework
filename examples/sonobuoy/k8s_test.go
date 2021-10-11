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

package sonobuoy

import (
	"context"
	"testing"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

// The following shows an example of a simple
// test function that reaches out to the API server.
func TestAPICall(t *testing.T) {
	feat := features.New("API Feature").
		WithLabel("type", "API").
		Assess("test message", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			var pods v1.PodList
			if err := c.Client().Resources("kube-system").List(ctx, &pods); err != nil {
				t.Error(err)
			}
			t.Logf("Got pods %v in namespace", len(pods.Items))
			if len(pods.Items) == 0 {
				t.Errorf("Expected >0 pods in kube-system but got %v", len(pods.Items))
			}
			return ctx
		}).Feature()

	// testenv is the one global that we rely on; it passes the context
	//and *envconf.Config to our feature.
	testenv.Test(t, feat)
}
