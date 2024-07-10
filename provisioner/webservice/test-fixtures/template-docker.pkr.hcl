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

  provisioner "paion-data-webservice-provisioner" {
    homeDir   = "/"
    warSource = "my-webservice.war"
  }
}
