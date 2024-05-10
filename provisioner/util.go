package provisioner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
	"os/user"
)

type NginxConfig struct {
	SslCertSource    string
	SslCertKeySource string

	HomeDir     string
	NginxConfig string
}

// ConfigNginxSSL creates a map of source and destination files for SSL configuration
func ConfigNginxSSL(config NginxConfig) (map[string]string, error) {

	sslCertDestination := filepath.Join(config.HomeDir, "ssl.crt")
	sslCertKeyDestination := filepath.Join(config.HomeDir, "ssl.key")

	file, err := tmp.File("nginx-config-file")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if _, err := file.WriteString(config.NginxConfig); err != nil {
		return nil, err
	}
	nginxDst := filepath.Join(config.HomeDir, "nginx-ssl.conf")

	configMap := map[string]string{
		config.SslCertSource:    sslCertDestination,
		config.SslCertKeySource: sslCertKeyDestination,
		file.Name():             nginxDst,
	}

	return configMap, nil
}

// ProvisionUpload uploads a file from the source to the destination
func ProvisionUpload(ui packersdk.Ui, communicator packersdk.Communicator, source string, destination string) error {
	ui.Say(fmt.Sprintf("Uploading %s => %s", source, destination))

	info, err := os.Stat(source)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("source should be a file; '%s', however, is a directory", source)
	}

	f, err := os.Open(source)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	filedst := destination
	if strings.HasSuffix(destination, "/") {
		filedst = destination + filepath.Base(source)
	}

	pf := ui.TrackProgress(filepath.Base(source), 0, info.Size(), f)
	defer pf.Close()

	// Upload the file
	if err = communicator.Upload(filedst, pf, &fi); err != nil {
		if strings.Contains(err.Error(), "Error restoring file") {
			ui.Error(fmt.Sprintf("Upload failed: %s; this can occur when "+
				"your file destination is a folder without a trailing "+
				"slash.", err))
		}
		ui.Error(fmt.Sprintf("Upload failed: %s", err))
		return err
	}

	return nil
}

func SkipConfigSSL(sslCertSource string, sslCertKeySource string, domain string) (bool, error) {
	if sslCertSource != "" && sslCertKeySource != "" && domain != "" {
		return false, nil
	}
	if sslCertSource == "" && sslCertKeySource == "" && domain == "" {
		return true, nil
	}
	return false, fmt.Errorf("sslCertSource, sslCertKeySource and domian must be set together")
}

// GetHomeDir create a directory if it does not exist
func GetHomeDir(homeDir string) (string, error) {
	if homeDir == "" {
		u, err := user.Current()
		if err != nil {
			return "", err
		}
		homeDir = u.HomeDir
	}

	stat, err := os.Stat(homeDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory '%s' does not exist", homeDir)
		} else {
			return homeDir, err
		}
	} else if !stat.IsDir() {
		return "", fmt.Errorf("'%s' is not a directory", homeDir)
	}

	return homeDir, nil
}
