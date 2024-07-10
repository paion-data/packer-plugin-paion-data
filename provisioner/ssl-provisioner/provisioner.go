// Copyright (c) Jiaqi Liu
// SPDX-License-Identifier: MPL-2.0

package ssl

import (
	"context"
	"encoding/base64"
	"fmt"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"github.com/paion-data/packer-plugin-paion-data/provisioner/file-provisioner"
	"github.com/paion-data/packer-plugin-paion-data/provisioner/shell"
	"path/filepath"
)

const defaultHomeDir string = "/home/ubuntu"
const nginxConfigDst string = "/etc/nginx/sites-enabled/default"
const nginxConfigFilename string = "nginx-ssl.conf"
const sslCertFilename string = "ssl.crt"
const sslCertKeyFilename string = "ssl.key"

const SslCertDst string = "/etc/ssl/certs/server.crt"
const SslCertKeyDst string = "/etc/ssl/private/server.key"

func Provision(
	ctx context.Context,
	interCtx interpolate.Context,
	ui packersdk.Ui,
	communicator packersdk.Communicator,
	homeDir string,
	sslCertBase64 string,
	sslCertKeyBase64 string,
	nginxConfig string,
) error {
	sslCert, err := DecodeBase64(sslCertBase64)
	sslCertSource, err := WriteToFile(sslCert)
	sslCertDestination := fmt.Sprintf(filepath.Join(homeDir, sslCertFilename))
	err = file.Provision(interCtx, ui, communicator, sslCertSource, sslCertDestination)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", sslCertSource, sslCertDestination, err)
	}

	sslCertKey, err := DecodeBase64(sslCertKeyBase64)
	sslCertKeySource, err := WriteToFile(sslCertKey)
	sslCertKeyDestination := fmt.Sprintf(filepath.Join(homeDir, sslCertKeyFilename))
	err = file.Provision(interCtx, ui, communicator, sslCertKeySource, sslCertKeyDestination)
	if err != nil {
		return fmt.Errorf("error uploading '%s' to '%s': %s", sslCertKeySource, sslCertKeyDestination, err)
	}

	if nginxConfig != "" {
		nginxSource, err := WriteToFile(nginxConfig)
		nginxDst := fmt.Sprintf(filepath.Join(homeDir, nginxConfigFilename))
		err = file.Provision(interCtx, ui, communicator, nginxSource, nginxDst)
		if err != nil {
			return fmt.Errorf("error uploading '%s' to '%s': %s", nginxSource, nginxDst, err)
		}
	}

	return shell.Provision(ctx, ui, communicator, getSslSetupCommands(homeDir))
}

// GetHomeDir Returns the home directory in Packer image builder. If a directory is specified, it is returned as it;
// otherwise the default Ubuntu home "/home/ubuntu" is returned
//
// configValue: A directory that can be either empty or a valid directory to be used unchanged.
//
// Returns:
// The actual home directory of the running Pakcer image builder
func GetHomeDir(configValue string) string {
	if configValue == "" {
		return defaultHomeDir
	}

	return configValue
}

// WriteToFile Flushes a specified string into a temporary file and returns the path of that file.
//
// content: The provided file content
//
// Returns:
// A path to the generated temporary file
func WriteToFile(content string) (string, error) {
	file, err := tmp.File("ssl-provisioner")
	if err != nil {
		return "", err
	}
	defer file.Close()
	if _, err := file.WriteString(content); err != nil {
		return "", err
	}

	return file.Name(), nil
}

// DecodeBase64 Decodes a base64-encoded string and returns the string representation of it
func DecodeBase64(encoded string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("error interpolating destination: %s", err)
	}
	return string(data), nil
}

// Return all commmnds for installing Nginx and loading SSL & Nginx config files to the proper location in remote
// machine
func getSslSetupCommands(homeDir string) []string {
	return []string{
		"sudo apt update && sudo apt upgrade -y",

		"sudo apt install -y nginx",
		fmt.Sprintf("sudo mv %s/%s %s", homeDir, nginxConfigFilename, nginxConfigDst),
		fmt.Sprintf("sudo mv %s/%s %s", homeDir, sslCertFilename, SslCertDst),
		fmt.Sprintf("sudo mv %s/%s %s", homeDir, sslCertKeyFilename, SslCertKeyDst),
	}
}
