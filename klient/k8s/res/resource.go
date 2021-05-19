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

package res

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	cr "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/e2e-framework/klient/k8s"
)

type Resources struct {
	// config is the rest.Config to talk to an apiserver
	config *rest.Config

	// scheme will be used to map go structs to GroupVersionKinds
	scheme *runtime.Scheme

	// client is a wrapper for controller runtime client
	client cr.Client
}

// Res instantiates the controller runtime client
// object. User can get panic for belopw scenarios.
// 1. if user does not provide k8s config
// 2. if controller runtime client instantiation fails.
func Res(cfg *rest.Config) *Resources {
	if cfg == nil {
		// TODO: logging
		fmt.Println("must provide rest.Config")
		panic(errors.New("must provide rest.Config"))
	}

	cl, err := cr.New(cfg, cr.Options{Scheme: scheme.Scheme})
	if err != nil {
		// TODO: log error
		panic(err)
	}

	res := &Resources{
		config: cfg,
		scheme: scheme.Scheme,
		client: cl,
	}

	return res
}

func (r *Resources) Get(ctx context.Context, name, namespace string, obj k8s.Object) error {
	return r.client.Get(ctx, cr.ObjectKey{Namespace: namespace, Name: name}, obj)
}

type CreateOption func(*metav1.CreateOptions)

func (r *Resources) Create(ctx context.Context, obj k8s.Object, opts ...CreateOption) error {
	createOptions := &metav1.CreateOptions{}
	for _, fn := range opts {
		fn(createOptions)
	}

	o := &cr.CreateOptions{Raw: createOptions}

	return r.client.Create(ctx, obj, o)
}

type UpdateOption func(*metav1.UpdateOptions)

func (r *Resources) Update(ctx context.Context, obj k8s.Object, opts ...UpdateOption) error {
	updateOptions := &metav1.UpdateOptions{}
	for _, fn := range opts {
		fn(updateOptions)
	}

	o := &cr.UpdateOptions{Raw: updateOptions}
	return r.client.Update(ctx, obj, o)
}

type DeleteOption func(*metav1.DeleteOptions)

func (r *Resources) Delete(ctx context.Context, obj k8s.Object, opts ...DeleteOption) error {
	deleteOptions := &metav1.DeleteOptions{}
	for _, fn := range opts {
		fn(deleteOptions)
	}

	o := &cr.DeleteOptions{Raw: deleteOptions}
	return r.client.Delete(ctx, obj, o)
}

type ListOption func(*metav1.ListOptions)

func (r *Resources) List(ctx context.Context, objs k8s.ObjectList, opts ...ListOption) error {
	listOptions := &metav1.ListOptions{}

	for _, fn := range opts {
		fn(listOptions)
	}

	o := &cr.ListOptions{Raw: listOptions}
	return r.client.List(ctx, objs, o)
}
