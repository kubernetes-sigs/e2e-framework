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

package environment

import (
	"context"
	"testing"
)

type Operation func (context.Context, Environment)(context.Context, error)

type Environment interface {
	// AddPrereq adds an environment operation to be executed
	// prior to the environment being ready
	AddPrereq(op Operation) Environment

	// WithContext returns a new Environment with specified context
	WithContext(ctx context.Context) Environment

	// Context returns context used for this env
	Context() context.Context

	// Test will execute a feature test using the given Context and T.
	Test(t *testing.T, f feature.Feature)

	// Namespace returns the namespace of this environment.
	Namespace() string

	//// RequiredLevel returns the feature level required for this environment.
	//RequiredLevel() Level
	//
	//// AssertionKind returns the assertion kind required for this environment.
	//AssertionKind() Kind

	//// AddObjectReference registers an API object reference to the environment,
	//// so that it can be cleaned up when execution of the environment is done.
	//AddObjectReferences(ref ...corev1.ObjectReference)

	//// References returns the list of known object references registered for
	//// the environment.
	//ObjectReferences() []corev1.ObjectReference

	// Finish signals the end of an environment run which will be used to delete
	// namespaces and object references registered in the environment.
	AddFinalizer(ops Operation) Environment
}