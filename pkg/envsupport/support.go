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

package envsupport

import (
	"context"
	"fmt"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/support/kind"
)

func CreateCluster(clusterName string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		k := kind.NewKindCluster(clusterName)
		kubecfg, err := k.Create()
		if err != nil {
			return ctx, err
		}

		fmt.Println("kind cluster created")

		// stall to wait for kind pods initialization
		waitTime := time.Second * 10
		fmt.Println("waiting for kind pods to initialize...", waitTime)
		time.Sleep(waitTime)
		return context.WithValue(context.WithValue(ctx, 1, kubecfg), 2, clusterName), nil
	}
}

func DestroyCluster(clusterName string) {
	// delete kind cluster
	k := kind.NewKindCluster(clusterName)

	err := k.Destroy()
	if err != nil {
		fmt.Println("error while deleting the cluster", err)
		return
	}
}
