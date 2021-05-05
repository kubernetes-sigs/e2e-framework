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

const(
	flagFeatureName = "feature"
	flagAssessName = "assess"
	flagLabelsName = "labels"
)
type Flags struct {
	feature string
	assess string
	labels LabelsMap
}

func (f *Flags) Feature() string {
	return f.feature
}

func (f *Flags) Assessment() string {
	return f.assess
}

func (f *Flags) Labels() LabelsMap {
	return f.labels
}

func Parse() (*Flags, error){
	return parseFlags(os.Args[0], os.Args[1:])
}

func parseFlags(cmdName string, flags []string) (*Flags, error){
	var feature string
	var assess string
	labels := make(LabelsMap)

	flagset := flag.NewFlagSet(cmdName, flag.ExitOnError)
	flagset.StringVar(&feature, flagFeatureName, "", "Regular expression that targets features to test")
	flagset.StringVar(&assess, flagAssessName, "", "Regular expression that targets assertive steps to run")
	flagset.Var(&labels, flagLabelsName, "Comma-separated key/value pairs to filter tests by labels")
	if err := flagset.Parse(flags); err != nil {
		return nil, err
	}

	return &Flags{feature: feature, assess: assess, labels:labels}, nil
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
		kv := strings.Split(label,"=")
		if len(kv) != 2 {
			return fmt.Errorf("label format error: %s", label)
		}
		m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	return nil
}