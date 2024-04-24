// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config,NginxConfig

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
	SslCertSource    string `mapstructure:"sslCertSource" required:"false"`
	SslCertKeySource string `mapstructure:"sslCertKeySource" required:"false"`

	ReactAppDomain string `mapstructure:"ReactAppDomain" required:"false"`
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
	p.config.HomeDir = util.GetHomeDir(p.config.HomeDir)

	nginxConfig := strings.Replace(getNginxConfigTemplate(), "react.domain.com", p.config.ReactAppDomain, -1)

	var err error
	skipConfigSSL, err = util.ConfigNginxSSL(ui, communicator, p.config.ctx, util.NginxConfig{
		SslCertSource:    p.config.SslCertSource,
		SslCertKeySource: p.config.SslCertKeySource,
		Domain:           p.config.ReactAppDomain,
		HomeDir:          p.config.HomeDir,
		NginxConfig:      nginxConfig,
	})

	if err != nil {
		return err
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
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",
		"sudo apt install -y nginx",

		"git clone https://github.com/paion-data/dental-llm-web-app",
		"cd dental-llm-web-app",
		"yarn",
		"yarn build",
	}
	cmd = append(cmd, fmt.Sprintf("sudo mv dist/ %s/", homeDir))

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
    listen 80 default_server;
    listen [::]:80 default_server;

    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;

    server_name _;

    location / {
        try_files $uri $uri/ =404;
    }
}

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
