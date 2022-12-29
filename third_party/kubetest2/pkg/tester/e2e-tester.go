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

package tester

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/vladimirvivien/gexe"
	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/testers"
)

var GitTag string

type Tester struct {
	TestFlags               string `desc:"Space-separated flags applied to 'go test' command"`
	Namespace               string `desc:"A namespace value to use for testing (optional)"`
	Kubeconfig              string `desc:"Path to a cluster kubeconfig file (optional)"`
	Feature                 string `desc:"Regular expression to select feature(s) to test"`
	Assess                  string `desc:"Regular expression to select assessment(s) to run"`
	Labels                  string `desc:"Comma-separated key=value to filter features by labels"`
	SkipLabels              string `desc:"Regular expression to skip label(s) to run"`
	SkipFeatures            string `desc:"Regular expression to skip feature(s) to run"`
	SkipAssessment          string `desc:"Regular expression to skip assessment(s) to run"`
	Parallel                bool   `desc:"Run test features in parallel"`
	DryRun                  bool   `desc:"Run Test suite in dry-run mode. This will list the tests to be executed without actually running them"`
	FailFast                bool   `desc:"Fail immediately and stop running untested code"`
	DisableGracefulTeardown bool   `desc:"Ignore panic recovery while running tests. This will prevent test finish steps from getting executed on panic"`
	Packages                string `desc:"Space-separated packages to test"`
}

const usage = `Usage: kubetest2-tester-e2e-framework [e2e-framework-flags]
When used with kubetest2: kubetest2 [<deployer>] --test=e2e-framework -- [e2e-framework-flags]
e2e-framework flags:
`

func (t *Tester) Execute(args []string) error {
	fs, err := gpflag.Parse(t)
	if err != nil {
		return fmt.Errorf("failed to initialize e2e-framework tester: %s", err)
	}

	fs.AddGoFlagSet(flag.CommandLine)

	help := fs.BoolP("help", "h", false, "")
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %s", err)
	}

	fs.Usage = func() {
		fmt.Print(usage)
	}

	if *help {
		fs.SetOutput(os.Stdout)
		fs.Usage()
		fs.PrintDefaults()
		return nil
	}

	if err := testers.WriteVersionToMetadata(GitTag); err != nil {
		return err
	}
	return t.Test()
}

func (t *Tester) Test() error {
	testCmd := t.buildCmd()
	klog.Info("Running: ", testCmd)
	p := gexe.NewProc(testCmd)
	p.SetStdout(os.Stdout)
	p.SetStderr(os.Stderr)
	return p.Run().Err()
}

func (t *Tester) buildCmd() string {
	var testCmd strings.Builder
	testCmd.WriteString(fmt.Sprintf("go test %s %s -args", t.TestFlags, t.Packages))
	if t.Namespace != "" {
		testCmd.WriteString(" --namespace=" + t.Namespace)
	}
	if t.Kubeconfig != "" {
		testCmd.WriteString(" --kubeconfig=" + t.Kubeconfig)
	}
	if t.Feature != "" {
		testCmd.WriteString(" --kubeconfig=" + t.Feature)
	}
	if t.SkipFeatures != "" {
		testCmd.WriteString(" --skip-features=" + t.SkipFeatures)
	}
	if t.Assess != "" {
		testCmd.WriteString(" --assess=" + t.Assess)
	}
	if t.SkipAssessment != "" {
		testCmd.WriteString(" --skip-assessment=" + t.SkipAssessment)
	}
	if t.Labels != "" {
		testCmd.WriteString(" --labels=" + t.Labels)
	}
	if t.SkipLabels != "" {
		testCmd.WriteString(" --skip-labels=" + t.SkipLabels)
	}
	if t.Parallel {
		testCmd.WriteString(" --parallel")
	}
	if t.DryRun {
		testCmd.WriteString(" --dry-run")
	}
	if t.FailFast {
		testCmd.WriteString(" --fail-fast")
	}
	if t.DisableGracefulTeardown {
		testCmd.WriteString(" --disable-graceful-shutdown")
	}

	return testCmd.String()
}

func Main() {
	t := &Tester{}
	if err := t.Execute(os.Args); err != nil {
		klog.Fatalf("failed to run e2e-framework tester: %v", err)
	}
}
