// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"github.com/paion-data/packer-plugin-paion-data/provisioner/react"
	artifactory "github.com/paion-data/packer-plugin-paion-data/provisioner/sonatype-nexus-repository"
	"github.com/paion-data/packer-plugin-paion-data/provisioner/webservice"
	"os"

	mailserver "github.com/paion-data/packer-plugin-paion-data/provisioner/docker-mailserver"
	gateway "github.com/paion-data/packer-plugin-paion-data/provisioner/kong-api-gateway"
	pluginVersion "github.com/paion-data/packer-plugin-paion-data/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterProvisioner("docker-mailserver-provisioner", new(mailserver.Provisioner))
	pps.RegisterProvisioner("kong-api-gateway-provisioner", new(gateway.Provisioner))
	pps.RegisterProvisioner("sonatype-nexus-repository-provisioner", new(artifactory.Provisioner))
	pps.RegisterProvisioner("webservice-provisioner", new(webservice.Provisioner))
	pps.RegisterProvisioner("react-provisioner", new(react.Provisioner))
	pps.SetVersion(pluginVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
