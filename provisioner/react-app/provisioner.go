// Copyright (c) Paion Data
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package reactApp

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	util "github.com/paion-data/packer-plugin-paion-data/provisioner"
)

type Config struct {
	SslCertSource    string `mapstructure:"sslCertSource" required:"true"`
	SslCertKeySource string `mapstructure:"sslCertKeySource" required:"true"`

	ReactAppDomain string `mapstructure:"ReactAppDomain" required:"true"`
	HomeDir        string `mapstructure:"homeDir" required:"false"`

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

var skipConfigSSL bool

func (p *Provisioner) Provision(ctx context.Context, ui packersdk.Ui, communicator packersdk.Communicator, generatedData map[string]interface{}) error {
	var err error
	p.config.HomeDir, err = util.GetHomeDir(p.config.HomeDir)

	nginxConfig := strings.Replace(getNginxConfigTemplate(), "react.domain.com", p.config.ReactAppDomain, -1)
	nginxConfigMap, err := util.ConfigNginxSSL(util.NginxConfig{
		SslCertSource:    p.config.SslCertSource,
		SslCertKeySource: p.config.SslCertKeySource,
		HomeDir:          p.config.HomeDir,
		NginxConfig:      nginxConfig,
	})

	if err != nil {
		return err
	}

	for source, destination := range nginxConfigMap {
		src, err := interpolate.Render(source, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("error interpolating source: %s", err)
		}

		dst, err := interpolate.Render(destination, &p.config.ctx)
		if err != nil {
			return fmt.Errorf("error interpolating destination: %s", err)
		}

		err = util.ProvisionUpload(ui, communicator, src, dst)
		if err != nil {
			return err
		}
	}

	for _, command := range getCommands(p.config.HomeDir) {
		err := (&packersdk.RemoteCmd{Command: command}).RunWithUi(ctx, communicator, ui)
		if err != nil {
			return err
		}
	}

	return nil
}

func getCommands(homeDir string) []string {
	cmd := []string{
		"DEBIAN_FRONTEND=noninteractive sudo apt update && DEBIAN_FRONTEND=noninteractive sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",
		"sudo apt install -y nginx",

		"sudo apt install -y curl",
		"curl -fsSL https://deb.nodesource.com/setup_16.x | sudo -E bash -",
		"sudo apt install -y nodejs",
		"sudo apt install -y serve",
	}

	if !skipConfigSSL {
		cmd = append(cmd, fmt.Sprintf("sudo mv %s/nginx-ssl.conf /etc/nginx/sites-enabled/default", homeDir))
		cmd = append(cmd, fmt.Sprintf("sudo mv %s/ssl.crt /etc/ssl/certs/server.crt", homeDir))
		cmd = append(cmd, fmt.Sprintf("sudo mv %s/ssl.key /etc/ssl/private/server.key", homeDir))
	}

	return cmd
}

func getNginxConfigTemplate() string {
	return `
server {
    if ($host = react.domain.com) {
        return 301 https://$host$request_uri;
    }

    listen 80 ;
    listen [::]:80 ;
    server_name react.domain.com;
    return 404;
}

server {
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name react.domain.com;

    location / {
        proxy_pass http://localhost:3000;
    }

    listen [::]:443 ssl ipv6only=on;
    listen 443 ssl;
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
}
    `
}
