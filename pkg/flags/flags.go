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

package flags

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	flagNamespaceName = "namespace"
	flagKubecofigName = "kubeconfig"
	flagFeatureName   = "feature"
	flagAssessName    = "assess"
	flagLabelsName    = "labels"
)

// Supported flag definitions
var (
	featureFlag = flag.Flag{
		Name:  flagFeatureName,
		Usage: "Regular expression to select feature(s) to test",
	}
	assessFlag = flag.Flag{
		Name:  flagAssessName,
		Usage: "Regular expression to select assessment(s) to run",
	}
	labelsFlag = flag.Flag{
		Name:  flagLabelsName,
		Usage: "Comma-separated key=value to filter features by labels",
	}
	kubecfgFlag = flag.Flag{
		Name:  flagKubecofigName,
		Usage: "Path to a cluster kubeconfig file (optional)",
	}
	kubeNSFlag = flag.Flag{
		Name:  flagNamespaceName,
		Usage: "A namespace value to use for testing (optional)",
	}
)

// EnvFlags surfaces all resolved flag values for the testing framework
type EnvFlags struct {
	feature string
	assess  string
	labels  LabelsMap

	// optional kube flags
	kubeconfig string
	namespace  string
}

// Feature returns value for `-feature` flag
func (f *EnvFlags) Feature() string {
	return f.feature
}

// Assessment returns value for `-assess` flag
func (f *EnvFlags) Assessment() string {
	return f.assess
}

// Labels returns a map of parsed key/value from `-labels` flag
func (f *EnvFlags) Labels() LabelsMap {
	return f.labels
}

// Namespace returns an optional namespace flag value
func (f *EnvFlags) Namespace() string {
	return f.namespace
}

// Kubeconfig returns an optional path for kubeconfig file
func (f *EnvFlags) Kubeconfig() string {
	return f.kubeconfig
}

// Parse parses defined CLI args os.Args[1:]
func Parse() (*EnvFlags, error) {
	return ParseArgs(os.Args[1:])
}

// ParseArgs parses the specified args from global flag.CommandLine
// and returns a set of environment flag values.
func ParseArgs(args []string) (*EnvFlags, error) {
	var feature string
	var assess string
	labels := make(LabelsMap)
	var namespace string
	var kubeconfig string

	if flag.Lookup(featureFlag.Name) == nil {
		flag.StringVar(&feature, featureFlag.Name, featureFlag.DefValue, featureFlag.Usage)
	}

	if flag.Lookup(assessFlag.Name) == nil {
		flag.StringVar(&assess, assessFlag.Name, assessFlag.DefValue, assessFlag.Usage)
	}

	if flag.Lookup(kubecfgFlag.Name) == nil {
		flag.StringVar(&kubeconfig, kubecfgFlag.Name, kubecfgFlag.DefValue, kubecfgFlag.Usage)
	}

	if flag.Lookup(kubeNSFlag.Name) == nil {
		flag.StringVar(&namespace, kubeNSFlag.Name, kubeNSFlag.DefValue, kubeNSFlag.Usage)
	}

	if flag.Lookup(labelsFlag.Name) == nil {
		flag.Var(&labels, labelsFlag.Name, labelsFlag.Usage)
	}

	if err := flag.CommandLine.Parse(args); err != nil {
		return nil, fmt.Errorf("flags parsing: %w", err)
	}

	return &EnvFlags{feature: feature, assess: assess, labels: labels, namespace: namespace, kubeconfig: kubeconfig}, nil
}

type LabelsMap map[string]string

func (m LabelsMap) String() string {
	i := map[string]string(m)
	return fmt.Sprint(i)
}

func (m LabelsMap) Set(val string) error {
	// label: []string{"key=value",...}
	for _, label := range strings.Split(val, ",") {
		// split into k,v
		kv := strings.Split(label, "=")
		if len(kv) != 2 {
			return fmt.Errorf("label format error: %s", label)
		}
		m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	return nil
}
