package vcluster

import "testing"

func TestCluster_VclusterInstall(t *testing.T) {
	c := Cluster{}
	err := c.findOrInstallVcluster()
	if err != nil {
		t.Fatalf("Error installing vcluster: %v", err)
	}
}
