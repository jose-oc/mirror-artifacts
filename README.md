# mirror-artifacts
Project to mirror helm charts and container images into a private registry (Google Artifact Registry)

## Run in macOS

The macOS binary is not notarized so macOS prevents you to use it with a message similar to:

```
```

but you can avoid this with 

```shell
xattr -dr com.apple.quarantine ./mirrorctl
```

## Build

### With go

```shell
go build .
```

### With goreleaser

```shell
brew install --cask goreleaser/tap/goreleaser
goreleaser release --snapshot --clean
```

If you want to generate SBOMs
```shell
brew install syft
```




go install github.com/sigstore/cosign/v3/cmd/cosign@latest

Verify the signature
```shell
cosign verify-blob --key cosign.pub --signature checksum.sig checksum
```




## Features
TODO: list the application features

List this repo features:

- New releases generate:
  - binaries for linux, mac and windows
  - SBOMs
  - 