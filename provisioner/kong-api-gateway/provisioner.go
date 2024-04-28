// Copyright (c) Paion Data
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package kongApiGateway

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

	KongApiGatewayDomain string `mapstructure:"kongApiGatewayDomain" required:"false"`
	HomeDir              string `mapstructure:"homeDir" required:"false"`

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

	skip, err := util.SkipConfigSSL(p.config.SslCertSource, p.config.SslCertKeySource, p.config.KongApiGatewayDomain)
	if err != nil {
		return err
	}

	if !skip {
		nginxConfig := strings.Replace(getNginxConfigTemplate(), "kong.domain.com", p.config.KongApiGatewayDomain, -1)
		nginxConfigMap, err := util.ConfigNginxSSL(ui, communicator, util.NginxConfig{
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

		"curl -fsSL https://get.docker.com -o get-docker.sh",
		"sh get-docker.sh",

		"git clone https://github.com/paion-data/docker-kong.git",

		"sudo apt install -y nginx",
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
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name kong.domain.com;

    location / {
        proxy_pass http://localhost:8000;
    }

    listen [::]:443 ssl ipv6only=on;
    listen 443 ssl;
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
}
server {
    if ($host = kong.domain.com) {
        return 301 https://$host$request_uri;
    }

    listen 80 ;
    listen [::]:80 ;
    server_name kong.domain.com;
    return 404;
}

server {
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name kong.domain.com;

    location / {
        proxy_pass http://localhost:8001;
    }

    listen [::]:8444 ssl ipv6only=on;
    listen 8444 ssl;
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
}
server {
    root /var/www/html;

    index index.html index.htm index.nginx-debian.html;
    server_name kong.domain.com;

    location / {
        proxy_pass http://localhost:8002;
    }

    listen [::]:8445 ssl ipv6only=on;
    listen 8445 ssl;
    ssl_certificate /etc/ssl/certs/server.crt;
    ssl_certificate_key /etc/ssl/private/server.key;
}
    `
}
