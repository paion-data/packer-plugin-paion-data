Packer Plugin paion-data
========================

![Go Badge][Go Badge]
[![HashiCorp Packer Badge][HashiCorp Packer Badge]][HashiCorp Packer URL]
![HashiCorp Packer SDK Badge][HashiCorp Packer SDK Badge]
[![GitHub Workflow Status][GitHub Workflow Status badge]][GitHub Workflow Status URL]
![GitHub Last Commit]
[![Apache License][Apache License Badge]][Apache License URL]

The paion-data multi-component plugin can be used with HashiCorp [Packer][HashiCorp Packer] to create images
supported by [immutable-infrastructure.com]. For the full list of available features for this plugin see [docs](./docs).

Installation
------------

### Using pre-built releases

#### Using the `packer init` command

Starting from version 1.7, Packer supports a new `packer init` command allowing automatic installation of Packer
plugins. Read the [Packer documentation][HashiCorp Packer init] for more information.

To install this plugin, copy and paste this code into Packer configuration. Then, run
[`packer init`][HashiCorp Packer init].

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

#### Manual installation

We can find pre-built binary releases of the plugin
[here](https://github.com/paion-data/packer-plugin-paion-data/releases). Once we have downloaded the latest archive
corresponding to our target OS, uncompress it to retrieve the plugin binary file corresponding to our platform. To
install the plugin, please follow the Packer documentation on
[installing a plugin][HashiCorp Packer installing a plugin].

### From Sources

If one prefer to build the plugin from sources, clone the GitHub repository locally and run the command `make build`
from the root directory. Upon successful compilation, a `packer-plugin-paion-data` plugin binary file can be found in
the root directory. To install the compiled plugin, please follow the official Packer documentation on
[installing a plugin][HashiCorp Packer installing a plugin].

### Configuration

For more information on how to configure the plugin, please read the documentation located in the [`docs/`](docs)
directory.

Contributing
------------

See [CONTRIBUTING.md](.github/CONTRIBUTING.md) for best practices and instructions on contributing to paion-data
Plugin.

Developing paion-data Plugin
----------------------------

### The Go Workspace

Go expects a single workspace for third-party Go tools installed via `go install`. By default, this workspace is located
in `$HOME/go` with source code for these tools stored in `$HOME/go/src` and the compiled binaries in `$HOME/go/bin`. Set
`$GOPATH` environment variable to this path first:

```shell
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

### Building from Source

1. Clone this GitHub repository locally:

   ```shell
   git clone git@github.com:paion-data/packer-plugin-paion-data.git
   cd packer-plugin-paion-data
   ```

2. Build the plugin from the root directory:

   ```shell 
   go build -ldflags="-X github.com/paion-data/packer-plugin-paion-data/version.VersionPrerelease=dev" -o packer-plugin-paion-data
   ```

3. After We successfully compile, the `packer-plugin-paion-data` plugin binary file is in the root directory.

4. To install the compiled plugin, run the following command

   ```shell
   packer plugins install --path packer-plugin-paion-data github.com/paion-data/paion-data
   ```

> [!TIP]
>
> If executing the `packer plugins install` reports an error, please make sure the version of `packer` command is the
> latest. To upgrade to the latest version, please refer to
> [Packer's documentation](https://developer.hashicorp.com/packer/tutorials/docker-get-started/get-started-install-cli)

### Running Acceptance Tests

Make sure to install the plugin locally using the steps in [Build from source](#building-from-source).

Once everything needed is set up, run:

```shell
PACKER_ACC=1 go test -count 1 -v ./... -timeout=120m
```

This will run the acceptance tests for all plugins in this set.

> [!CAUTION]
> 
> Please make sure the acceptance tests are running against the local version by deleting all previously installed
> versions under `$HOME/. config/packer/plugins` directory. Otherwise, the tests will pick up the old released version
> if they were installed before. Deleting `github.com/paion-data/paion-data`, for example, would be
> 
> ```console
> rm -rf ~/.config/packer/plugins/github.com/paion-data/paion-data
> ```

## Registering Plugin as Packer Integration

Partner and community plugins can be hard to find if a user doesn't know what
they are looking for. To assist with plugin discovery Packer offers an integration
portal at https://developer.hashicorp.com/packer/integrations to list known integrations
that work with the latest release of Packer.

Registering a plugin as an integration requires [metadata configuration](./metadata.hcl) within the plugin
repository and approval by the Packer team. To initiate the process of registering your
plugin as a Packer integration refer to the [Developing Plugins](https://developer.hashicorp.com/packer/docs/plugins/creation#registering-plugins) page.

License
-------

The use and distribution terms for [packer-plugin-paion-data] are covered by the [Apache License, Version 2.0].

[Apache License Badge]: https://img.shields.io/badge/Apache%202.0-F25910.svg?style=for-the-badge&logo=Apache&logoColor=white
[Apache License URL]: https://www.apache.org/licenses/LICENSE-2.0
[Apache License, Version 2.0]: http://www.apache.org/licenses/LICENSE-2.0.html

[GitHub Last Commit]: https://img.shields.io/github/last-commit/paion-data/packer-plugin-paion-data/master?logo=github&style=for-the-badge
[GitHub Workflow Status badge]: https://img.shields.io/github/actions/workflow/status/paion-data/packer-plugin-paion-data/ci-cd.yml?branch=master&logo=github&style=for-the-badge
[GitHub Workflow Status URL]: https://github.com/paion-data/packer-plugin-paion-data/actions/workflows/ci-cd.yml
[Go Badge]: https://img.shields.io/badge/Go%20>=%201.20-00ADD8?style=for-the-badge&logo=go&logoColor=white

[immutable-infrastructure.com]: https://immutable-infrastructure.com
[HashiCorp Packer]: https://developer.hashicorp.com/packer/docs
[HashiCorp Packer init]: https://developer.hashicorp.com/packer/docs/commands/init
[HashiCorp Packer installing a plugin]: https://developer.hashicorp.com/packer/docs/plugins#installing-plugins
[HashiCorp Packer SDK Badge]: https://img.shields.io/badge/Packer%20Plugin%20SDK>=%20v0.5.2-000000?style=for-the-badge&logo=hashicorp&logoColor=white
[HashiCorp Packer SDK URL]: https://github.com/hashicorp/packer-plugin-sdk
[HashiCorp Packer Badge]: https://img.shields.io/badge/Packer%20>=%20v1.11.0-02A8EF?style=for-the-badge&logo=Packer&logoColor=white
[HashiCorp Packer URL]: https://developer.hashicorp.com/packer/docs
