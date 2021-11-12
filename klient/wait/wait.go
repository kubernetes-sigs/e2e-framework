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
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	apimachinerywait "k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

const (
	defaultPollTimeout  = 5 * time.Minute
	defaultPollInterval = 5 * time.Second
)

type Interface interface {
	For(cond apimachinerywait.ConditionFunc) error
	ForWithIntervalAndTimeout(interval time.Duration, timeout time.Duration, cond apimachinerywait.ConditionFunc) error
}

type waiter struct {
	resources *resources.Resources
}

func New(resources *resources.Resources) *waiter {
	return &waiter{resources: resources}
}

func (w *waiter) ResourceScaled(obj k8s.Object, scaleFetcher func(object k8s.Object) int32, replica int32) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking if the resource %s/%s has been scaled to %d", obj.GetNamespace(), obj.GetName(), replica)
		if err := w.resources.Get(context.TODO(), obj.GetName(), obj.GetNamespace(), obj); err != nil {
			return false, nil
		}
		return scaleFetcher(obj) == replica, nil
	}
}

func (w *waiter) ResourceDeleted(obj k8s.Object) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Resource deletion of %s/%s", obj.GetNamespace(), obj.GetName())
		if err := w.resources.Get(context.Background(), obj.GetName(), obj.GetNamespace(), obj); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}
}

func (w *waiter) JobConditionMatch(job k8s.Object, conditionType batchv1.JobConditionType, conditionState v1.ConditionStatus) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Job Condition %s/%s on %s/%s", conditionType, conditionState, job.GetNamespace(), job.GetName())
		if err := w.resources.Get(context.TODO(), job.GetName(), job.GetNamespace(), job); err != nil {
			return false, err
		}
		for _, cond := range job.(*batchv1.Job).Status.Conditions {
			if cond.Type == conditionType && cond.Status == conditionState {
				done = true
			}
		}
		return
	}
}

func (w *waiter) PodConditionMatch(pod k8s.Object, conditionType v1.PodConditionType, conditionState v1.ConditionStatus) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Pod Condition %s/%s on %s/%s", conditionType, conditionState, pod.GetNamespace(), pod.GetName())
		if err := w.resources.Get(context.TODO(), pod.GetName(), pod.GetNamespace(), pod); err != nil {
			return false, err
		}
		for _, cond := range pod.(*v1.Pod).Status.Conditions {
			if cond.Type == conditionType && cond.Status == conditionState {
				done = true
			}
		}
		return
	}
}

func (w *waiter) PodPhaseMatch(pod k8s.Object, phase v1.PodPhase) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Pod %v Condition of %s/%s", phase, pod.GetNamespace(), pod.GetName())
		if err := w.resources.Get(context.Background(), pod.GetName(), pod.GetNamespace(), pod); err != nil {
			return false, err
		}
		return pod.(*v1.Pod).Status.Phase == phase, nil
	}
}

func (w *waiter) PodReady(pod k8s.Object) apimachinerywait.ConditionFunc {
	return w.PodConditionMatch(pod, v1.PodReady, v1.ConditionTrue)
}

func (w *waiter) ContainersReady(pod k8s.Object) apimachinerywait.ConditionFunc {
	return w.PodConditionMatch(pod, v1.ContainersReady, v1.ConditionTrue)
}

func (w *waiter) PodRunning(pod k8s.Object) apimachinerywait.ConditionFunc {
	return w.PodPhaseMatch(pod, v1.PodRunning)
}

func (w *waiter) JobCompleted(job k8s.Object) apimachinerywait.ConditionFunc {
	return w.JobConditionMatch(job, batchv1.JobComplete, v1.ConditionTrue)
}

func (w *waiter) JobFailed(job k8s.Object) apimachinerywait.ConditionFunc {
	return w.JobConditionMatch(job, batchv1.JobFailed, v1.ConditionTrue)
}

func (w *waiter) For(cond apimachinerywait.ConditionFunc) error {
	return apimachinerywait.PollImmediate(defaultPollInterval, defaultPollTimeout, cond)
}

func (w *waiter) ForWithIntervalAndTimeout(interval time.Duration, timeout time.Duration, cond apimachinerywait.ConditionFunc) error {
	return apimachinerywait.PollImmediate(interval, timeout, cond)
}
