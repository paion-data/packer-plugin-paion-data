// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package webservice

import (
	"context"
	"fmt"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/paion-data/packer-plugin-paion-data/provisioner/file-provisioner"
	"github.com/paion-data/packer-plugin-paion-data/provisioner/shell"
	"github.com/paion-data/packer-plugin-paion-data/provisioner/ssl-provisioner"
	"path/filepath"
)

type Config struct {
	WarSource string `mapstructure:"warSource" required:"true"`
	HomeDir   string `mapstructure:"homeDir" required:"false"`

	ctx interpolate.Context
}

type Provisioner struct {
	config Config
}

func (p *Provisioner) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *Provisioner) Prepare(raws ...interface{}) error {
	err := config.Decode(&p.config, nil, raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, communicator packersdk.Communicator, generatedData map[string]interface{}) error {
	p.config.HomeDir = ssl.GetHomeDir(p.config.HomeDir)

	warFileDst := fmt.Sprintf(filepath.Join(p.config.HomeDir, "ROOT.war"))

	err := file.Provision(p.config.ctx, ui, communicator, p.config.WarSource, warFileDst)
	if err != nil {
		return err
	}

	return shell.Provision(ctx, ui, communicator, getCommands(p.config.HomeDir))
}

func getCommands(homeDir string) []string {
	return append(
		getCommandsUpdatingUbuntu(),
		append(getCommandsInstallingJDK17(), getCommandsInstallingJetty(homeDir)...)...,
	)
}

func getCommandsUpdatingUbuntu() []string {
	return []string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",
	}
}

// Install JDK 17 - https://www.rosehosting.com/blog/how-to-install-java-17-lts-on-ubuntu-20-04/
func getCommandsInstallingJDK17() []string {
	return []string{
		"sudo apt update -y",
		"sudo apt install openjdk-17-jdk -y",
		"export JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64",
	}
}

// Install and configure Jetty (for JDK 17) container
func getCommandsInstallingJetty(homeDir string) []string {
	return []string{
		"export JETTY_VERSION=11.0.15",
		"wget https://repo1.maven.org/maven2/org/eclipse/jetty/jetty-home/$JETTY_VERSION/jetty-home-$JETTY_VERSION.tar.gz",
		"tar -xzvf jetty-home-$JETTY_VERSION.tar.gz",
		"rm jetty-home-$JETTY_VERSION.tar.gz",
		fmt.Sprintf("export JETTY_HOME=%s/jetty-home-$JETTY_VERSION", homeDir),
		"mkdir jetty-base",
		"cd jetty-base",
		"java -jar $JETTY_HOME/start.jar --add-module=annotations,server,http,deploy,servlet,webapp,resources,jsp",
		fmt.Sprintf("mv %s/ROOT.war webapps/ROOT.war", homeDir),
		"cd ../",
	}
}
