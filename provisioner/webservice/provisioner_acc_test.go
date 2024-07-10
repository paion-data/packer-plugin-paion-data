// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webservice

import (
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/template-docker.pkr.hcl
var testProvisionerHCL2Docker string

func TestAccWebserviceProvisioner(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "my-webservice.war")
	if err != nil {
		return
	}

	testCaseDocker := &acctest.PluginTestCase{
		Name: "webservice_provisioner_docker_test",
		Setup: func() error {
			return nil
		},
		Teardown: func() error {
			return nil
		},
		Template: strings.Replace(testProvisionerHCL2Docker, "my-webservice.war", tempFile.Name(), -1),
		Type:     "hashicorp-aws-webservice-provisioner",
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
				t.Fatalf("%s\n Acceptance tests for %s failed. Please search for '%s' in log file at %s", logsString, "webservice provisioner", errorString, logfile)
			}

			provisionerOutputLog := "docker.hashicorp-aws: Exported Docker file:"
			if matched, _ := regexp.MatchString(provisionerOutputLog+".*", logsString); !matched {
				t.Fatalf("%s\n logs doesn't contain expected output %q", logsString, provisionerOutputLog)
			}

			return nil
		},
	}
	acctest.TestPlugin(t, testCaseDocker)
}
