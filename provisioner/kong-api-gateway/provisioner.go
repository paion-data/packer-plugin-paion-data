// Copyright (c) Paion Data
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package kongApiGateway

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2/hcldec"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	util "github.com/paion-data/packer-plugin-paion-data/provisioner"
	jwt_util "github.com/paion-data/packer-plugin-paion-data/provisioner"
)

type Config struct {
	SslCertSource    string `mapstructure:"sslCertSource" required:"false"`
	SslCertKeySource string `mapstructure:"sslCertKeySource" required:"false"`

	KongApiGatewayDomain string `mapstructure:"kongApiGatewayDomain" required:"false"`
	HomeDir              string `mapstructure:"homeDir" required:"false"`

	JwksUrl string `mapstructure:"jwksUrl" required:"false"`
	JwtIss  string `mapstructure:"jwtIss" required:"false"`

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

	skip, err := util.SkipConfigSSL(p.config.SslCertSource, p.config.SslCertKeySource, p.config.KongApiGatewayDomain)
	if err != nil {
		return err
	}

	if !skip {
		nginxConfig := strings.Replace(getNginxConfigTemplate(), "kong.domain.com", p.config.KongApiGatewayDomain, -1)
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
			err = upload(source, destination, p, ui, communicator)
			if err != nil {
				return err
			}
		}

		// create a file to store the public key
		file, err := os.Create("publicKey.txt")
		if err != nil {
			return err
		}
		defer file.Close()

		// get the public key
		publicKey, err := jwt_util.GetJWKSPublicKeyPEM(p.config.JwksUrl)
		if err != nil {
			return err
		}

		// write the public key to the file
		_, err = file.WriteString(publicKey)
		if err != nil {
			return err
		}

		currentDir, err := os.Getwd()
		if err != nil {
			return err
		}

		source := filepath.Join(currentDir, "publicKey.txt")
		destination := filepath.Join(p.config.HomeDir, "publicKey.txt")

		err = upload(source, destination, p, ui, communicator)
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

func upload(source string, destination string, p *Provisioner, ui packersdk.Ui, communicator packersdk.Communicator) error {
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

	return nil
}

func getCommands(homeDir string) []string {
	cmd := []string{
		"sudo apt update && sudo apt upgrade -y",
		"sudo apt install software-properties-common -y",

		"curl -fsSL https://get.docker.com -o get-docker.sh",
		"sh get-docker.sh",

		"sudo apt install -y nginx",
	}

	cmd = append(cmd, fmt.Sprintf("cd %s", homeDir))
	cmd = append(cmd, "git clone https://github.com/paion-data/docker-kong.git")

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
