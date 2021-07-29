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
	"log"
	"time"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/support/kind"
)

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

var (
	// ContextKeyKubeConfig key for setting kube config value in context
	ContextKeyKubeConfig = ContextKey("kubeconfig")
	// ContextKeyClusterName key for setting clustername value in context
	ContextKeyClusterName = ContextKey("clustername")
)

// GetStringValueFromContext gets the caller value from the context.
// Use this function to retrieve value of a context key
// or can print out value of key by doing,
// fmt.Println("Key is:", ContextKeyKubeConfig.String())
func GetStringValueFromContext(ctx context.Context, key ContextKey) (string, bool) {
	value, ok := ctx.Value(key).(string)
	return value, ok
}

func CreateCluster(clusterName string) env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		k := kind.NewCluster(clusterName)
		kubecfg, err := k.Create()
		if err != nil {
			return ctx, err
		}

		// stall to wait for kind pods initialization
		waitTime := time.Second * 10
		time.Sleep(waitTime)
		return context.WithValue(context.WithValue(ctx, ContextKeyKubeConfig, kubecfg), ContextKeyClusterName, clusterName), nil
	}
}

func DestroyCluster(clusterName string) {
	// delete kind cluster
	k := kind.NewCluster(clusterName)

	err := k.Destroy()
	if err != nil {
		log.Println("error while deleting the cluster", err)
		return
	}
}
