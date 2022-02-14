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

package pod

import (
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/constructor/container"
)

type SpecConstructor struct {
	spec coreV1.PodSpec
}

// Spec initializer method for type PodSpecConstructor
func Spec(containerConstructors... container.Constructor) SpecConstructor {
	spec := SpecConstructor{spec: coreV1.PodSpec{}}
	for _, constructor := range containerConstructors {
		spec.spec.Containers = append(spec.spec.Containers, constructor.Build())
	}
	return spec
}

// AddVolume is a setter method to store coreV1.Volume definition
func (c SpecConstructor) AddVolume(vol coreV1.Volume) SpecConstructor {
	c.spec.Volumes = append(c.spec.Volumes, vol)
	return c
}

// Build is the finalizer method that returns a value of type coreV1.PodSpec
func (c SpecConstructor) Build() coreV1.PodSpec {
	return c.spec
}