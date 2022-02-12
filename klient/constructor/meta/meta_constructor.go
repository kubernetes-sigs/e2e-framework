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

// Package meta contains constructor type to build values of type *coreV1.ObjectMeta
package meta

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	EmptyName = ""
	EmptyLabels = map[string]string{}
    EmptyObjectMeta = &metaV1.ObjectMeta{}
)

// ObjectMetaConstructor constructor for type coreV1.ObjectMeta
type ObjectMetaConstructor struct {
	obj *metaV1.ObjectMeta
}

// Object is the constructor function for ObjectMetaConstructor
func Object(name string) *ObjectMetaConstructor {
	return &ObjectMetaConstructor{obj: &metaV1.ObjectMeta{Name: name}}
}

// Namespace setter for namespace value
func (c *ObjectMetaConstructor) Namespace(ns string) *ObjectMetaConstructor {
	c.obj.Namespace = ns
	return c
}

// Labels setter for labels
func (c *ObjectMetaConstructor) Labels(labels map[string]string) *ObjectMetaConstructor {
	c.obj.Labels = labels
	return c
}

// ClusterName setter for cluster name value
func (c *ObjectMetaConstructor) ClusterName(name string) *ObjectMetaConstructor {
	c.obj.ClusterName = name
	return c
}

// Build is the finalizer that returns *metav1.ObjectMeta
func (c *ObjectMetaConstructor) Build() (*metaV1.ObjectMeta, error) {
	return c.obj, nil
}
