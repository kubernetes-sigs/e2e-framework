/*
Copyright 2022 The Kubernetes Authors.

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

// Package deployment provides constructor types for appV1.Deployment
package deployment

import (
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/constructor/meta"
	"sigs.k8s.io/e2e-framework/klient/constructor/pod"
)

type Constructor struct {
	deployment appsV1.Deployment
}

// Deployment is the initializer function for type deployment.Constructor
func Deployment(
	deploymentMeta meta.ObjectMetaConstructor,
	replicas *int32,
	selector meta.LabelSelectorConstructor,
	strategy StrategyConstructor,
	template pod.TemplateSpecConstructor,
) Constructor {
	tempSpec := template.Build()
	depSel := selector.Build()
	return Constructor{
		deployment: appsV1.Deployment{
			ObjectMeta: deploymentMeta.Build(),
			Spec: appsV1.DeploymentSpec{
				Replicas: replicas,
				Selector: &depSel,
				Template: v1.PodTemplateSpec{
					ObjectMeta: tempSpec.ObjectMeta,
					Spec:       tempSpec.Spec,
				},
				Strategy: strategy.Build(),
			},
		},
	}
}

// Build is the finalizer method that constructs and return a value of type appsV1.Deployment
func (c Constructor) Build() appsV1.Deployment{
	return c.deployment
}

// Replicas is an initializer func that converts int -> *int32
func Replicas(r int) *int32 {
	rep := int32(r)
	return &rep
}
