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

package conditions

import (
	"context"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apimachinerywait "k8s.io/apimachinery/pkg/util/wait"
	"log"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
)

type Condition struct {
	resources *resources.Resources
}

func New(resources *resources.Resources) *Condition {
	return &Condition{resources: resources}
}

func (c *Condition) ResourceScaled(obj k8s.Object, scaleFetcher func(object k8s.Object) int32, replica int32) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking if the resource %s/%s has been scaled to %d", obj.GetNamespace(), obj.GetName(), replica)
		if err := c.resources.Get(context.TODO(), obj.GetName(), obj.GetNamespace(), obj); err != nil {
			return false, nil
		}
		return scaleFetcher(obj) == replica, nil
	}
}

func (c *Condition) ResourceDeleted(obj k8s.Object) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Resource deletion of %s/%s", obj.GetNamespace(), obj.GetName())
		if err := c.resources.Get(context.Background(), obj.GetName(), obj.GetNamespace(), obj); err != nil {
			if errors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}
}

func (c *Condition) JobConditionMatch(job k8s.Object, conditionType batchv1.JobConditionType, conditionState v1.ConditionStatus) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Job Condition %s/%s on %s/%s", conditionType, conditionState, job.GetNamespace(), job.GetName())
		if err := c.resources.Get(context.TODO(), job.GetName(), job.GetNamespace(), job); err != nil {
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

func (c *Condition) PodConditionMatch(pod k8s.Object, conditionType v1.PodConditionType, conditionState v1.ConditionStatus) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Pod Condition %s/%s on %s/%s", conditionType, conditionState, pod.GetNamespace(), pod.GetName())
		if err := c.resources.Get(context.TODO(), pod.GetName(), pod.GetNamespace(), pod); err != nil {
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

func (c *Condition) PodPhaseMatch(pod k8s.Object, phase v1.PodPhase) apimachinerywait.ConditionFunc {
	return func() (done bool, err error) {
		log.Printf("Checking for Pod %v Condition of %s/%s", phase, pod.GetNamespace(), pod.GetName())
		if err := c.resources.Get(context.Background(), pod.GetName(), pod.GetNamespace(), pod); err != nil {
			return false, err
		}
		return pod.(*v1.Pod).Status.Phase == phase, nil
	}
}

func (c *Condition) PodReady(pod k8s.Object) apimachinerywait.ConditionFunc {
	return c.PodConditionMatch(pod, v1.PodReady, v1.ConditionTrue)
}

func (c *Condition) ContainersReady(pod k8s.Object) apimachinerywait.ConditionFunc {
	return c.PodConditionMatch(pod, v1.ContainersReady, v1.ConditionTrue)
}

func (c *Condition) PodRunning(pod k8s.Object) apimachinerywait.ConditionFunc {
	return c.PodPhaseMatch(pod, v1.PodRunning)
}

func (c *Condition) JobCompleted(job k8s.Object) apimachinerywait.ConditionFunc {
	return c.JobConditionMatch(job, batchv1.JobComplete, v1.ConditionTrue)
}

func (c *Condition) JobFailed(job k8s.Object) apimachinerywait.ConditionFunc {
	return c.JobConditionMatch(job, batchv1.JobFailed, v1.ConditionTrue)
}
