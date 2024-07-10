  Include a short description about the provisioner. This is a good place
  to call out what the provisioner does, and any additional text that might
  be helpful to a user. See https://www.packer.io/docs/provisioner/null
-->

The `sonatype-nexus-repository` provisioner is used to install Sonatype Nexus Repository package in AWS AMI image


<!-- Provisioner Configuration Fields -->

**Required**

- `sonatypeNexusRepositoryDomain` (string) - the SSL-enabled domain that will serve the deployed HTTP Nexus instance.
- `sslCertBase64` (string) - is a __base64 encoded__ string of the content of
  [SSL certificate file](https://immutable-infrastructure.com/docs/setup#optional-setup-ssl) for the SSL-enabled domain, for
  example `nexus.mycompany.com` given the `sonatypeNexusRepositoryDomain` is `nexus.mycompany.com`.
- `sslCertKeyBase64` (string) - is a __base64 encoded__ string of the content of
  [SSL certificate key file](https://immutable-infrastructure.com/docs/setup#optional-setup-ssl) for the SSL-enabled domain, for
  example `nexus.mycompany.com` given the `sonatypeNexusRepositoryDomain` is `nexus.mycompany.com`.

<!--
  Optional Configuration Fields

  Configuration options that are not required or have reasonable defaults
  should be listed under the optionals section. Defaults values should be
  noted in the description of the field
-->

**Optional**

- `homeDir` (string) - The `$Home` directory in AMI image; default to `/home/ubuntu`

<!--
  A basic example on the usage of the provisioner. Multiple examples
  can be provided to highlight various configurations.

-->

### Example Usage

```hcl
packer {
  required_plugins {
    amazon = {
      version = ">= 0.0.2"
      source  = "github.com/hashicorp/amazon"
    }
  }
}

source "amazon-ebs" "paion-data" {
  ami_name              = "packer-plugin-paion-data-acc-test-ami"
  force_deregister      = "true"
  force_delete_snapshot = "true"

  instance_type = "t2.micro"
  launch_block_device_mappings {
    device_name           = "/dev/sda1"
    volume_size           = 8
    volume_type           = "gp2"
    delete_on_termination = true
  }
  region = "us-west-1"
  source_ami_filter {
    filters = {
      name                = "ubuntu/images/*ubuntu-*-22.04-amd64-server-*"
      root-device-type    = "ebs"
      virtualization-type = "hvm"
    }
    most_recent = true
    owners      = ["099720109477"]
  }
  ssh_username = "ubuntu"
}

build {
  sources = [
    "source.amazon-ebs.paion-data"
  ]

  provisioner "paion-data-sonatype-nexus-repository-provisioner" {
    homeDir                       = "/home/ubuntu"
    sslCertBase64                 = "YXNkZnNnaHRkeWhyZXJ3ZGZydGV3ZHNmZ3RoeTY0cmV3ZGZyZWd0cmV3d2ZyZw=="
    sslCertKeyBase64              = "MzI0NXRnZjk4dmJoIGNsO2VbNDM1MHRdzszNDM1b2l0cmo="
    sonatypeNexusRepositoryDomain = "nexus.mycompany.com"
  }
}
```
