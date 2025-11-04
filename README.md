# Mirror Artifacts

A CLI tool for mirroring Helm charts and container images into private registries, particularly Google Artifact Registry.

## Table of Contents
- [Introduction](#introduction)
- [Repository Structure](#repository-structure)
- [Infrastructure](#infrastructure)
- [Mirrorctl CLI Tool](#mirrorctl-cli-tool)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Input File Format](#input-file-format)
- [Building the CLI Tool](#building-the-cli-tool)

## Introduction

Mirror Artifacts is a command-line tool designed to simplify the process of mirroring Helm charts and container images from public repositories to private registries. This is particularly useful for organizations that want to maintain internal copies of public artifacts for security, compliance, or connectivity reasons.

## Repository Structure

The repository is organized as follows:

```
mirror-artifacts/
├── .gitignore
├── .goreleaser.yaml
├── helm-charts.yaml
├── images.yaml
├── README.md
├── .github/
│   └── workflows/
├── infra/
│   └── gcp/
└── mirrorctl/
    ├── .gitignore
    ├── go.mod
    └── main.go
```

- **mirrorctl/**: Contains the Go source code for the CLI tool
- **infra/gcp/**: Infrastructure as code for Google Cloud Platform components
- **.goreleaser.yaml**: Configuration for building and releasing binaries
- **helm-charts.yaml, images.yaml**: Sample configuration files for defining artifacts to mirror

## Infrastructure

The infrastructure directory contains configurations for deploying the necessary components to Google Cloud Platform.

### What it does
The infrastructure provisions resources needed to host and manage mirrored artifacts, including:
- Google Artifact Registry instances

### How to run it
To deploy the infrastructure, navigate to the `infra/gcp/` directory and follow the specific deployment instructions provided there.

## Mirrorctl CLI Tool

The `mirrorctl` command-line tool provides the primary interface for mirroring artifacts.

### Goals
- Simplify the process of mirroring public Helm charts and container images
- Provide a secure way to manage artifacts in private registries
- Automate the mirroring process through configuration files
- Support multiple target registries and platforms

### Installation

#### Downloading Binaries
You can download pre-built binaries for your platform from the [releases page](https://github.com/jose-oc/mirror-artifacts/releases). Choose the appropriate binary based on your operating system and architecture:

- **Linux**: `mirrorctl_linux_amd64`, `mirrorctl_linux_arm64`
- **macOS**: `mirrorctl_darwin_amd64`, `mirrorctl_darwin_arm64`
- **Windows**: `mirrorctl_windows_amd64.exe`

#### macOS Security Note
Since the macOS binaries are not signed with the Apple tool, macOS may block them. To bypass this security measure, run the following command:

```shell
xattr -dr com.apple.quarantine /path/to/mirrorctl
```

## Configuration

The `mirrorctl` tool looks for a configuration file at `.mirrorctl.yaml` by default in your home directory or in the same directory where the binary is. 
You can specify a different configuration file path using the `--config` flag.

The configuration file contains settings for target registries, credentials, and other operational parameters.

## Usage

To use `mirrorctl`, run commands from your terminal:

```shell
mirrorctl [command] [flags]
```

### Available Parameters

#### Global Flags
- `--config`: Path to the configuration file (default is $HOME/.mirrorctl.yaml)
- `--dry-run`: Simulate actions without executing
- `--keep-temp-dir`: Keep temporary directories for inspection
- `--log-color`: Enables colored output in development mode (default true)
- `--log-file`: If set, writes logs to the specified file path instead of the console
- `--log-level`: Sets the minimum log level (e.g., debug, info, warn, error) (default "info")
- `--prod-mode`: Enables production-style JSON logging

#### Mirror Images Command
- `--images`: Path to YAML file with list of container images

Example:
```shell
mirrorctl mirror images --images images.yaml
```

#### Mirror Charts Command
- `--charts`: Path to YAML file with list of Helm charts

Example:
```shell
mirrorctl mirror charts --charts helm-charts.yaml
```

#### Generate SBOM from Charts Command

This command generates Software Bill of Materials (SBOM) for a list of Helm charts. 
The SBOM is a list of all container images used by the charts, even if certain conditions are required for the image to used.
The SBOM can be saved in a file in JSON or YAML format.

- `--charts`: Path to YAML file with a list of Helm charts
- `--output-file`: Path to the output file in JSON or YAML format

Example:
```shell
mirrorctl sbom list chart-images --charts=charts.yaml --output-file=charts-images-sbom.yaml
```

## Input File Format

The input files for `mirrorctl` use YAML format to define artifacts to be mirrored:

### Helm Charts Format
```yaml
charts:
  - name: grafana-agent-operator
    source: https://grafana.github.io/helm-charts
    version: 0.5.1
```

### Container Images Format
```yaml
  - name: hello-world
    source: docker.io/library/hello-world:latest
  - name: alpine
    source: docker.io/library/alpine:3.22.2
  - name: curl
    source: quay.io/curl/curl:8.16.0
```

## Building the CLI Tool

There are two ways to build the `mirrorctl` CLI tool:

### Option 1: Using Go Build

Navigate to the `mirrorctl` directory and run:

```shell
cd mirrorctl
go build -o mirrorctl .
```

This will create a binary named `mirrorctl` in the current directory.

### Option 2: Using Goreleaser

Goreleaser builds binaries for all platforms, signs checksums, and publishes to GitHub. 
This tool is typically used for official releases.

First, install Goreleaser. On macOS, you can use Homebrew:
```shell
brew install --cask goreleaser/tap/goreleaser
```
For other platforms, see the [Goreleaser documentation](https://goreleaser.com/install/).

You have to run `goreleaser` from the root directory of the repository.

#### Building

If you just want to build a binary for your local platform, use:
```shell
goreleaser build --single-target --clean
```

If you made changes in the code, you can run `goreleaser build --single-target --snapshot --clean` to build a binary for your local platform.

This approach only builds for your current architecture and operating system, making it faster for local development and testing.

The `--clean` flag cleans up the dist directory before building to ensure a clean build environment.

#### Releasing

It requires a couple of tools to be installed:
- `syft`
- `cosign`

Building binaries for all platforms (typically for releases):
```shell
goreleaser release --clean
```
