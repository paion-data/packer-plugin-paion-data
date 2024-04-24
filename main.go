// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	dockerMailServerProv "github.com/paion-data/packer-plugin-paion-data/provisioner/docker-mailserver"
	kongApiGatewayProv "github.com/paion-data/packer-plugin-paion-data/provisioner/kong-api-gateway"
	reactAppProv "github.com/paion-data/packer-plugin-paion-data/provisioner/react-app"
	pluginVersion "github.com/paion-data/packer-plugin-paion-data/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterProvisioner("docker-mailserver-provisioner", new(dockerMailServerProv.Provisioner))
	pps.RegisterProvisioner("kong-api-gateway-provisioner", new(kongApiGatewayProv.Provisioner))
	pps.RegisterProvisioner("react-app-provisioner", new(reactAppProv.Provisioner))
	pps.SetVersion(pluginVersion.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
