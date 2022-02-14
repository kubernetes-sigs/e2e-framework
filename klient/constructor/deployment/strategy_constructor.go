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

package deployment

import (
	appsV1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	//StrategyDefault is set to value "Recreate"
	StrategyDefault = StrategyConstructor{strat: appsV1.DeploymentStrategy{Type: appsV1.RecreateDeploymentStrategyType}}
	StrategyRecreate = StrategyDefault
)

// StrategyConstructor is type to build values of type appsV1.DeploymentStrategy
type StrategyConstructor struct {
	strat appsV1.DeploymentStrategy
}

// StrategyRollingUpdate is an initializer function that creates a value of type appsV1.DeploymentStrategy
func StrategyRollingUpdate(maxUnavailable string, maxSurge string) StrategyConstructor {
	unavailParsed := intstr.FromString(maxUnavailable)
	surgeParsed := intstr.FromString(maxSurge)
	return StrategyConstructor{
		strat: appsV1.DeploymentStrategy{
			Type: appsV1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsV1.RollingUpdateDeployment{
				MaxUnavailable: &unavailParsed,
				MaxSurge:       &surgeParsed,
			},
		},
	}
}

// Build is a finalizer method that constructs the value of type appsV1.DeploymentStrategy
func (c StrategyConstructor) Build() appsV1.DeploymentStrategy {
	return c.strat
}