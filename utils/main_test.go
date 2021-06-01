package utils_test

import (
	"testing"

	"wiley.com/do-k8s-cluster-health-check/utils"
)

func TestLoadConfig(t *testing.T) {
	if err := utils.LoadConfig("../tests/cmd/", "config.yml"); err != nil {
		t.Error(err)
	}
}
