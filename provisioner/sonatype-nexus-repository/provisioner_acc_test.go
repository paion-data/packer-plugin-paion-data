// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package artifactory

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/template-docker.pkr.hcl
var testProvisionerHCL2Docker string

func TestAccSonatypeNexusRepositoryProvisioner(t *testing.T) {
	testCaseDocker := &acctest.PluginTestCase{
		Name: "sonatype_nexus_repository_provisioner_docker_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: testProvisionerHCL2Docker,
		Type:     "paion-data-sonatype-nexus-repository-provisioner",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}

			logs, err := os.Open(logfile)
			if err != nil {
				return fmt.Errorf("Unable find %s", logfile)
			}
			defer logs.Close()

			logsBytes, err := ioutil.ReadAll(logs)
			if err != nil {
				return fmt.Errorf("Unable to read %s", logfile)
			}
			logsString := string(logsBytes)

			errorString := "error(s) occurred"
			if matched, _ := regexp.MatchString(".*"+errorString+".*", logsString); matched {
				t.Fatalf("Acceptance tests for %s failed. Please search for '%s' in log file at %s", "sonatype-nexus-repository provisioner", errorString, logfile)
			}

			provisionerOutputLog := "docker.paion-data: Exported Docker file:"
			if matched, _ := regexp.MatchString(provisionerOutputLog+".*", logsString); !matched {
				t.Fatalf("logs doesn't contain expected output %q", logsString)
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCaseDocker)
}
