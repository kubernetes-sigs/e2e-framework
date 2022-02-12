package meta

import (
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LabelSelectorConstructor struct{
	sel *metaV1.LabelSelector
}

// MatchLabels initializer function for type LabelSelectorConstructor
func MatchLabels(labels map[string]string) *LabelSelectorConstructor {
	return &LabelSelectorConstructor{sel: &metaV1.LabelSelector{MatchLabels: labels}}
}

// MatchExpressions initializer function for type LabelSelectorConstructor
func MatchExpressions(expressions...metaV1.LabelSelectorRequirement) *LabelSelectorConstructor {
	return &LabelSelectorConstructor{sel: &metaV1.LabelSelector{MatchExpressions: expressions}}
}

// MatchLabels setter for map[string]string labels
func (c *LabelSelectorConstructor) MatchLabels(labels map[string]string) *LabelSelectorConstructor {
	c.sel.MatchLabels = labels
	return c
}

// MatchExpressions setter for metaV1.LabelSelectorRequirement
func (c *LabelSelectorConstructor) MatchExpressions(expressions...metaV1.LabelSelectorRequirement) *LabelSelectorConstructor {
	c.sel.MatchExpressions = expressions
	return c
}

// Build is the finalizer method that returns the built *metaV1.LabelSelector value
func (c *LabelSelectorConstructor) Build() (*metaV1.LabelSelector, error) {
	return c.sel, nil
}
