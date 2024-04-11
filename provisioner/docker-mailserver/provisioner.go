// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package dockerMailserver

import (
	"context"
	"fmt"
	"github.com/QubitPi/packer-plugin-hashicorp-aws/provisioner"
	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"path/filepath"
	"strings"
)

type Config struct {
	BaseDomain string `mapstructure:"baseDomain" required:"true"`

	HomeDir string `mapstructure:"homeDir" required:"false"`

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
	p.config.HomeDir = getHomeDir(p.config.HomeDir)
	composeFile := strings.Replace(getDockerComposeFileTemplate(), "mail.domain.com", "mail."+p.config.BaseDomain, -1)
	file, err := tmp.File("docker-compose-file")
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.WriteString(composeFile); err != nil {
		return err
	}
	composeFile = ""
	composeFileDst := fmt.Sprintf(filepath.Join(p.config.HomeDir, "compose.yaml"))
	err = p.ProvisionUpload(ui, communicator, file.Name(), composeFileDst)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", file.Name(), composeFileDst, err)
	}

	for _, command := range getCommands(p.config.HomeDir, "mail."+p.config.BaseDomain) {
		err := (&packersdk.RemoteCmd{Command: command}).RunWithUi(ctx, communicator, ui)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Provisioner) ProvisionUpload(ui packersdk.Ui, communicator packersdk.Communicator, source string, destination string) error {
	src, err := interpolate.Render(source, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating source: %s", err)
	}

	dst, err := interpolate.Render(destination, &p.config.ctx)
	if err != nil {
		return fmt.Errorf("error interpolating destination: %s", err)
	}

	return provisioner.ProvisionUpload(ui, communicator, src, dst)
}

func getHomeDir(configValue string) string {
	if configValue == "" {
		return "/home/ubuntu"
	}

	return configValue
}

func getDockerComposeFileTemplate() string {
	return `
services:
  mailserver:
    image: ghcr.io/docker-mailserver/docker-mailserver:latest
    container_name: mailserver
    hostname: mail.domain.com
    env_file: mailserver.env
    ports:
      - "25:25"
      - "143:143"
      - "465:465"
      - "587:587"
      - "993:993"
    volumes:
      - ./docker-data/dms/mail-data/:/var/mail/
      - ./docker-data/dms/mail-state/:/var/mail-state/
      - ./docker-data/dms/mail-logs/:/var/log/mail/
      - ./docker-data/dms/config/:/tmp/docker-mailserver/
      - /etc/localtime:/etc/localtime:ro
      - ./docker-data/certbot/certs/:/etc/letsencrypt
    restart: always
    stop_grace_period: 1m
    healthcheck:
      test: "ss --listening --tcp | grep -P 'LISTEN.+:smtp' || exit 1"
      timeout: 3s
      retries: 0
    environment:
      - SSL_TYPE=letsencrypt
    `
}

func getCommands(homeDir string, domain string) []string {
	return []string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",

		"curl -fsSL https://get.docker.com -o get-docker.sh",
		"sh get-docker.sh",

		"wget \"https://raw.githubusercontent.com/docker-mailserver/docker-mailserver/master/mailserver.env\"",
	}
}
