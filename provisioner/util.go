package provisioner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer-plugin-sdk/tmp"
)

type NginxConfig struct {
	SslCertSource    string
	SslCertKeySource string

	Domain      string
	HomeDir     string
	NginxConfig string
}

func ConfigNginxSSL(ui packersdk.Ui, communicator packersdk.Communicator, ctx interpolate.Context, config NginxConfig) (bool, error) {
	skip, err := skipConfigSSL(config.SslCertSource, config.SslCertKeySource, config.Domain)
	if err != nil {
		return skip, err
	}

	if skip {
		return skip, nil
	}
	sslCertDestination := filepath.Join(config.HomeDir, "ssl.crt")
	err = ProvisionUpload(ui, communicator, config.SslCertSource, sslCertDestination, ctx)
	if err != nil {
		return skip, fmt.Errorf("error uploading '%s' to '%s': %s", config.SslCertSource, sslCertDestination, err)
	}

	sslCertKeyDestination := filepath.Join(config.HomeDir, "ssl.key")
	err = ProvisionUpload(ui, communicator, config.SslCertKeySource, sslCertKeyDestination, ctx)
	if err != nil {
		return skip, fmt.Errorf("error uploading '%s' to '%s': %s", config.SslCertKeySource, sslCertKeyDestination, err)
	}

	file, err := tmp.File("nginx-config-file")
	if err != nil {
		return skip, err
	}
	defer file.Close()
	if _, err := file.WriteString(config.NginxConfig); err != nil {
		return skip, err
	}

	nginxDst := filepath.Join(config.HomeDir, "nginx-ssl.conf")
	err = ProvisionUpload(ui, communicator, file.Name(), nginxDst, ctx)
	if err != nil {
		return skip, fmt.Errorf("error uploading '%s' to '%s': %s", file.Name(), nginxDst, err)
	}

	return skip, nil
}

// ProvisionUpload uploads a file from the source to the destination
func ProvisionUpload(ui packersdk.Ui, communicator packersdk.Communicator, source string, destination string, ctx interpolate.Context) error {

	src, err := interpolate.Render(source, &ctx)
	if err != nil {
		return fmt.Errorf("error interpolating source: %s", err)
	}

	dst, err := interpolate.Render(destination, &ctx)
	if err != nil {
		return fmt.Errorf("error interpolating destination: %s", err)
	}

	ui.Say(fmt.Sprintf("Uploading %s => %s", src, dst))

	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("source should be a file; '%s', however, is a directory", src)
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	filedst := dst
	if strings.HasSuffix(dst, "/") {
		filedst = dst + filepath.Base(src)
	}

	pf := ui.TrackProgress(filepath.Base(src), 0, info.Size(), f)
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

func skipConfigSSL(sslCertSource string, sslCertKeySource string, domain string) (bool, error) {
	if sslCertSource != "" && sslCertKeySource != "" && domain != "" {
		return false, nil
	}
	if sslCertSource == "" && sslCertKeySource == "" && domain == "" {
		return true, nil
	}
	return false, fmt.Errorf("sslCertSource, sslCertKeySource and domian must be set together")
}

func GetHomeDir(configValue string) string {
	if configValue == "" {
		return "/root"
	}

	return configValue
}
