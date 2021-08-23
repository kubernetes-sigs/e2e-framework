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

package conf

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/util/homedir"
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	home := homedir.HomeDir()

	kubeconfigdir := filepath.Join(home, "test", ".kube")
	kubeconfigpath := filepath.Join(kubeconfigdir, "config")

	// check if file exists
	_, err := os.Stat(kubeconfigpath)
	// create file if not exists
	if os.IsNotExist(err) {
		err = os.MkdirAll(kubeconfigdir, 0777)
		if err != nil {
			log.Println("failed to create .kube dir", err)
			return
		}

		// generate kube config data
		data := genKubeconfig("test-context")

		err = createFile(kubeconfigpath, data)
		if err != nil {
			log.Println("failed to create config file", err)
			return
		}
	}

	log.Println("file created successfully", kubeconfigpath)

	flag.StringVar(&kubeconfig, "kubeconfig", "", "Paths to a kubeconfig. Only required if out-of-cluster.")

	// set --kubeconfig flag
	err = flag.Set("kubeconfig", kubeconfigpath)
	if err != nil {
		log.Println("unexpected error while setting flag value", err)
		return
	}

	flag.Parse()
}

func createFile(path, data string) error {
	return ioutil.WriteFile(path, []byte(data), 0644)
}

// genKubeconfig used to genearte kube config file
// we can provide multiple contexts as well
func genKubeconfig(contexts ...string) string {
	var sb strings.Builder
	sb.WriteString(`---
apiVersion: v1
kind: Config
clusters:
`)
	for _, ctx := range contexts {
		sb.WriteString(`- cluster:
    server: ` + ctx + `
  name: ` + ctx + `
`)
	}
	sb.WriteString("contexts:\n")
	for _, ctx := range contexts {
		sb.WriteString(`- context:
    cluster: ` + ctx + `
    user: ` + ctx + `
  name: ` + ctx + `
`)
	}

	sb.WriteString("users:\n")
	for _, ctx := range contexts {
		sb.WriteString(`- name: ` + ctx + `
`)
	}
	sb.WriteString("preferences: {}\n")
	if len(contexts) > 0 {
		sb.WriteString("current-context: " + contexts[0] + "\n")
	}

	return sb.String()
}

func teardown() {
	home := homedir.HomeDir()
	err := os.RemoveAll(filepath.Join(home, "test"))
	if err != nil {
		log.Println("failed to delete .kube dir", err)
	}
}
