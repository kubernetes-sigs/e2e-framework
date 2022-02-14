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

package meta

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	SelectorNone = LabelSelectorConstructor{sel: metaV1.LabelSelector{}}
)

type LabelSelectorConstructor struct{
	sel metaV1.LabelSelector
}

// MatchLabels initializer function for type LabelSelectorConstructor
func MatchLabels(labels map[string]string) LabelSelectorConstructor {
	return LabelSelectorConstructor{sel: metaV1.LabelSelector{MatchLabels: labels}}
}

// MatchExpressions initializer function for type LabelSelectorConstructor
func MatchExpressions(expressions...metaV1.LabelSelectorRequirement) LabelSelectorConstructor {
	return LabelSelectorConstructor{sel: metaV1.LabelSelector{MatchExpressions: expressions}}
}

// MatchLabels setter for map[string]string labels
func (c LabelSelectorConstructor) MatchLabels(labels map[string]string) LabelSelectorConstructor {
	c.sel.MatchLabels = labels
	return c
}

// MatchExpressions setter for metaV1.LabelSelectorRequirement
func (c LabelSelectorConstructor) MatchExpressions(expressions...metaV1.LabelSelectorRequirement) LabelSelectorConstructor {
	c.sel.MatchExpressions = expressions
	return c
}

// Build is the finalizer method that returns the built *metaV1.LabelSelector value
func (c LabelSelectorConstructor) Build() metaV1.LabelSelector {
	return c.sel
}
