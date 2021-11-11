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

package wait

import (
	"context"
	"log"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apimachinerywait "k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

const (
	defaultPollTimeout  = 5 * time.Minute
	defaultPollInterval = 5 * time.Second
)

type Interface interface {
	For(cond apimachinerywait.ConditionFunc) error
	ForWithIntervalAndTimeout(interval time.Duration, timeout time.Duration, resource *resources.Resources, cond apimachinerywait.ConditionFunc) error
}

type waiter struct {
	resources *resources.Resources
}

func New(resources *resources.Resources) *waiter {
	return &waiter{resources: resources}
}

func checkIfPodIsRunning(pod *v1.Pod) bool {
	switch pod.Status.Phase {
	case v1.PodRunning:
		return true
	default:
		return false
	}
}

func (w *waiter) PodReadyCondition(pod *v1.Pod) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Pod Ready Condition of %s/%s", pod.GetNamespace(), pod.GetName())
		if err := w.resources.Get(context.Background(), pod.GetName(), pod.GetNamespace(), pod); err != nil {
			return false, err
		}
		return checkIfPodIsRunning(pod), nil
	}
}

func (w *waiter) PodReadyConditionBySelector(selector string) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Waiting for Pod Ready Condition using Label selector %s", selector)
		var pods v1.PodList
		if err := w.resources.List(context.Background(), &pods, resources.WithLabelSelector(selector)); err != nil {
			return false, err
		}
		allOk := true
		for _, pod := range pods.Items {
			if ok := checkIfPodIsRunning(&pod); !ok {
				allOk = false
				break
			}
		}
		return allOk, nil
	}
}

func (w *waiter) DeploymentScaled(deployment *appsv1.Deployment, replica int) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for the Scale of Deployment %s/%s to be %d", deployment.GetNamespace(), deployment.GetName(), replica)
		if err := w.resources.Get(context.Background(), deployment.GetName(), deployment.GetNamespace(), deployment); err != nil {
			return false, err
		}
		return *deployment.Spec.Replicas == int32(replica), nil
	}
}

func (w *waiter) For(cond apimachinerywait.ConditionFunc) error {
	return apimachinerywait.PollImmediate(defaultPollInterval, defaultPollTimeout, cond)
}

func (w *waiter) ForWithIntervalAndTimeout(interval time.Duration, timeout time.Duration, resource *resources.Resources, cond apimachinerywait.ConditionFunc) error {
	return apimachinerywait.PollImmediate(interval, timeout, cond)
}
