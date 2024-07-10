// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package react

import (
	"bytes"
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
	"text/template"
)

// PORT Default port of React app
const PORT string = "3000"

// NODE_VERSION Default node version running the React app
const NODE_VERSION = "18"

type Config struct {
	DistSource       string `mapstructure:"distSource" required:"true"`
	SslCertBase64    string `mapstructure:"sslCertBase64" required:"true"`
	SslCertKeyBase64 string `mapstructure:"sslCertKeyBase64" required:"true"`
	AppDomain        string `mapstructure:"appDomain" required:"true"`
	NodeVersion      string `mapstructure:"nodeVersion" required:"false"`
	HomeDir          string `mapstructure:"homeDir" required:"false"`

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

	distFileDst := fmt.Sprintf(filepath.Join(p.config.HomeDir, "dist"))
	err := file.Provision(p.config.ctx, ui, communicator, p.config.DistSource, distFileDst)
	if err != nil {
		return err
	}

	if p.config.NodeVersion == "" {
		p.config.NodeVersion = NODE_VERSION
	}
	err = shell.Provision(ctx, ui, communicator, getCommands(p.config.NodeVersion))
	if err != nil {
		return err
	}

	return ssl.Provision(ctx, p.config.ctx, ui, communicator, p.config.HomeDir, p.config.SslCertBase64, p.config.SslCertKeyBase64, getNginxConfig(p.config.AppDomain))
}

func getNginxConfig(domain string) string {
	var sslConfigs = struct {
		Domain        string
		SslCertDst    string
		SslCertKeyDst string
		Port          string
	}{domain, ssl.SslCertDst, ssl.SslCertKeyDst, PORT}
	var buf bytes.Buffer
	t := template.Must(template.New("Nginx Config").Parse(`
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
    server_name {{.Domain}};

    location / {
        proxy_pass http://localhost:{{.Port}};
    }

    listen [::]:443 ssl ipv6only=on;
    listen 443 ssl;
    ssl_certificate {{.SslCertDst}};
    ssl_certificate_key {{.SslCertKeyDst}};
}
server {
    if ($host = {{.Domain}}) {
        return 301 https://$host$request_uri;
    }

    listen 80 ;
    listen [::]:80 ;
    server_name {{.Domain}};
    return 404;
}
	`))

	if err := t.Execute(&buf, sslConfigs); err != nil {
		panic(err)
	}

	return buf.String()
}

func getCommands(nodeVersion string) []string {
	return append(getCommandsUpdatingUbuntu(), getCommandsInstallingNode(nodeVersion)...)
}

func getCommandsUpdatingUbuntu() []string {
	return []string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",
	}
}

func getCommandsInstallingNode(nodeVersion string) []string {
	return []string{
		"sudo apt install -y curl",
		fmt.Sprintf("curl -fsSL https://deb.nodesource.com/setup_%s.x | sudo -E bash -", nodeVersion),
		"sudo apt install -y nodejs",
		"sudo npm install -g yarn",
		"sudo npm install -g serve",
	}
}
