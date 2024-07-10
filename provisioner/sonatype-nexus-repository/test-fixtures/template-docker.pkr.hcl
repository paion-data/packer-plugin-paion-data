# Copyright (c) Jiaqi
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    docker = {
      version = ">= 0.0.7"
      source  = "github.com/hashicorp/docker"
    }
  }
}

source "docker" "paion-data" {
  image  = "jack20191124/packer-plugin-hashicorp-aws-acc-test-base:latest"
  discard = true
}

build {
  sources = [
    "source.docker.paion-data"
  ]

  provisioner "paion-data-sonatype-nexus-repository-provisioner" {
    homeDir                       = "/"
    sslCertBase64                 = "YXNkZnNnaHRkeWhyZXJ3ZGZydGV3ZHNmZ3RoeTY0cmV3ZGZyZWd0cmV3d2ZyZw=="
    sslCertKeyBase64              = "MzI0NXRnZjk4dmJoIGNsO2VbNDM1MHRdzszNDM1b2l0cmo="
    sonatypeNexusRepositoryDomain = "nexus.mycompany.com"
  }
}
