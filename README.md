# JOCMirrorCtl

**`jocmirrorctl`** is a Go-based CLI tool that automates the mirroring of **Helm charts** and their **container images** into **Google Artifact Registry (GAR)**.  
It also manages provenance, SBOM generation, and signing — all integrated for use in **GitHub Actions**.

---

## Features

- Mirror container images referenced in Helm charts into GAR.
- Mirror and modify Helm charts so they point to mirrored GAR images.
- Append a `-joc` suffix to the chart version (e.g. `1.2.3-joc`).
- Preserve provenance and original digests.
- Generate and attach SBOMs to mirrored artifacts.
- Sign charts and images using [cosign](https://github.com/sigstore/cosign).
- Run in **dry-run** or **verbose** mode for safe testing.
- Designed for **GitHub Actions automation**.

---

## Project Structure

```bash
.
├── cmd/
│   └── jocmirrorctl/           # CLI entry point
├── pkg/
│   ├── charts/                 # Helm chart logic
│   ├── images/                 # Container image mirroring
│   ├── sbom/                   # SBOM generation
│   ├── sign/                   # Signing logic
│   ├── config/                 # Config (viper)
│   ├── logging/                # zerolog setup
│   ├── utils/                  # Helper functions
│   └── cmdutils/               # CLI helper functions
└── infra/                      # Terraform/GCP setup
```

---

## Installation

You can build `jocmirrorctl` locally:

```bash
git clone https://github.com/<your-org>/jocmirrorctl.git
cd jocmirrorctl
go build -o jocmirrorctl ./cmd/jocmirrorctl
```

Or run it directly with Go:

```bash
go run ./cmd/jocmirrorctl
```

---

## Configuration

`jocmirrorctl` uses **Viper** for configuration management.
By default, it loads `config.yaml` (or another path specified via `--config`).

### Example `config.yaml`

```yaml
gcp:
  project_id: my-gcp-project
  region: europe-west1
  gar_repo: europe-west1-docker.pkg.dev/my-gcp-project/helm-mirror

options:
  dry_run: false
  verbose: true
  suffix: "-joc"
  notify_tag_mutations: true
  sbom_format: cyclonedx
  provenance_store: ./provenance
```

### Config fields

| Key                            | Type   | Description                                  |
| ------------------------------ | ------ | -------------------------------------------- |
| `gcp.project_id`               | string | GCP project ID                               |
| `gcp.region`                   | string | GCP region                                   |
| `gcp.gar_repo`                 | string | Full GAR repo name                           |
| `options.dry_run`              | bool   | Logs actions without executing               |
| `options.verbose`              | bool   | Enables debug logging                        |
| `options.suffix`               | string | Version suffix for mirrored charts           |
| `options.notify_tag_mutations` | bool   | Warn if upstream tags change digest          |
| `options.sbom_format`          | string | SBOM format (`cyclonedx`, `spdx`)            |
| `options.provenance_store`     | string | Directory to store provenance JSON files     |

---

## Usage

### General command pattern

```bash
jocmirrorctl [command] [flags]
```

### Global Flags

| Flag        | Description                                      |
| ----------- | ------------------------------------------------ |
| `--config`  | Path to configuration file                       |
| `--charts`  | Path to YAML file with list of Helm charts       |
| `--images`  | Path to YAML file with list of container images  |
| `--dry-run` | Simulate actions without pushing                 |
| `--verbose` | Enable debug logs                                |

---

### Subcommands

#### Mirror all (charts + images)

```bash
jocmirrorctl mirror all \
  --config config.yaml \
  --charts charts.yaml \
  --images images.yaml
```

Mirrors images and charts, modifies them to use GAR image references, appends the `-joc` suffix, and pushes to GAR.

#### Mirror only container images

```bash
jocmirrorctl mirror images \
  --config config.yaml \
  --images images.yaml
```

Mirrors images specified in the config file (or discovered from Helm templates).

#### Mirror only Helm charts

```bash
jocmirrorctl mirror charts \
  --config config.yaml \
  --charts charts.yaml
```

#### Input files format

`charts.yaml`
```yaml
charts:
  - name: nginx
    source: https://charts.bitnami.com/bitnami/nginx
    version: 15.2.3
  - name: redis
    source: https://charts.bitnami.com/bitnami/redis
    version: 18.1.4
```

`images.yaml`
```yaml
images:
  - name: nginx
    source: docker.io/bitnami/nginx:1.25.3
  - name: redis
    source: docker.io/bitnami/redis:7.2.0
```

Mirrors Helm charts, updates image references, and pushes to GAR.

#### 📦 Generate SBOMs and provenance

```bash
jocmirrorctl mirror sbom --config config.yaml
```

Generates SBOMs for mirrored artifacts and attaches them as referrers.

#### 🔐 Sign artifacts

```bash
jocmirrorctl mirror sign --key cosign.key --config config.yaml
```

Signs mirrored Helm charts and images with Cosign.

#### ✅ Verify mirrored artifacts

```bash
jocmirrorctl verify --config config.yaml
```

Verifies that charts pull images from GAR and that all signatures are valid.

---

## 🔧 Example GitHub Actions Integration

```yaml
name: Mirror Helm Charts to GAR

on:
  workflow_dispatch:

jobs:
  mirror:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: 1.23

      - name: Setup tools
        run: |
          sudo apt-get install -y helm oras
          curl -sSL https://github.com/sigstore/cosign/releases/latest/download/cosign-linux-amd64 -o /usr/local/bin/cosign
          chmod +x /usr/local/bin/cosign
          gcloud auth configure-docker europe-west1-docker.pkg.dev

      - name: Run mirror
        run: |
          go run ./cmd/jocmirrorctl mirror all --config config.yaml --verbose
```

---

## 🧠 Notes

* The tool uses **Helm SDK** and **ORAS Go library** — not shelling out to CLIs.
* When a library is used instead of a CLI, equivalent command comments will be included in the code (for clarity).
* Designed to work **idempotently** — running it twice should only mirror updated artifacts.
* Dry-run mode makes it safe to test workflows before real pushes.

---

## 📋 Roadmap

* [ ] CLI skeleton with Cobra, Viper, Zerolog
* [ ] Image mirroring via ORAS library
* [ ] Helm chart modification logic
* [ ] SBOM and provenance attachment
* [ ] Cosign signing and verification
* [ ] GitHub Actions integration
* [ ] Vulnerability scanning (Trivy)

---

> “Make complex things simple, not simpler than they should be.” — A principle this project follows.

