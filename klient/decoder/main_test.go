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

package decoder

import (
	"context"
	"os"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/e2e-framework/klient/internal/testutil"
)

var (
	tc        *testutil.TestCluster
	clientset kubernetes.Interface
	ctx       = context.TODO()
	cfg       *rest.Config
)

func TestMain(m *testing.M) {
	tc = testutil.SetupTestCluster("")
	clientset = tc.Clientset
	cfg = tc.RESTConfig
	code := m.Run()
	teardown()
	os.Exit(code)
}

func teardown() {
	tc.DestroyTestCluster()
}
