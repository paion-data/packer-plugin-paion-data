<!--
  Include a short overview about the plugin.

  This document is a great location for creating a table of contents for each
  of the components the plugin may provide. This document should load automatically
  when navigating to the docs directory for a plugin.

-->

### Installation

To install this plugin, copy and paste this code into Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    paion-data = {
      version = ">= 0.0.1"
      source = "github.com/paion-data/paion-data"
    }
  }
}
```

Alternatively, we can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/paion-data/paion-data
```

### Components

#### Provisioners

Provisioners are used to execute scripts on remote machine as part of AWS EC2 related resource creation and destruction.
They enable

1. programmatic configuration management which is not possible with HCL, and
2. code reuse

The business logics that satisfy the two criteria above are offered as the paion-data provisioners below:

- [React App](./provisioners/react.mdx)
- [Sonatype Nexus Repository](./provisioners/sonatype-nexus-repository.mdx)
- [Jersey-Jetty Webservice](./provisioners/webservice.mdx)
