# kustomize-sopssecretgenerator

[![Build Status](https://travis-ci.org/Meiqia/kustomize-sopssecretgenerator.svg?branch=master)](https://travis-ci.org/Meiqia/kustomize-sopssecretgenerator)
[![Go Report Card](https://goreportcard.com/badge/github.com/Meiqia/kustomize-sopssecretgenerator)](https://goreportcard.com/report/github.com/Meiqia/kustomize-sopssecretgenerator)
[![Codecov](https://img.shields.io/codecov/c/github/Meiqia/kustomize-sopssecretgenerator)](https://codecov.io/gh/Meiqia/kustomize-sopssecretgenerator)
[![Latest Release](https://img.shields.io/github/v/release/Meiqia/kustomize-sopssecretgenerator?sort=semver)](https://github.com/Meiqia/kustomize-sopssecretgenerator/releases/latest)
[![License](https://img.shields.io/github/license/Meiqia/kustomize-sopssecretgenerator)](https://github.com/Meiqia/kustomize-sopssecretgenerator/blob/master/LICENSE)

An generator plugin for [kustomize](https://github.com/kubernetes-sigs/kustomize)
that generates Secrets from files encrypted with [sops](https://github.com/mozilla/sops).


## Installation

Download the `SopsSecretGenerator` binary for your platform from the
[GitHub releases page](https://github.com/Meiqia/kustomize-sopssecretgenerator/releases) and
move it to `$XDG_CONFIG_HOME/kustomize/plugin/kustomize.meiqia.com/v1beta1/sopssecretgenerator`. (By default,
`$XDG_CONFIG_HOME` points to `$HOME/.config` on Linux and OS X and `%LOCALAPPDATA%` on Windows.)

For example, to install version 1.2.0 on Linux:

    VERSION=1.2.0 PLATFORM=linux ARCH=amd64
    curl -Lo SopsSecretGenerator https://github.com/Meiqia/kustomize-sopssecretgenerator/releases/download/v${VERSION}/SopsSecretGenerator_${VERSION}_${PLATFORM}_${ARCH}
    chmod +x SopsSecretGenerator
    mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/kustomize.meiqia.com/v1beta1/sopssecretgenerator"
    mv SopsSecretGenerator "${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/kustomize.meiqia.com/v1beta1/sopssecretgenerator"


## Usage

Create some encrypted values using `sops`:

    echo FOO=secret >secret-vars.env
    sops -e -i secret-vars.env
    
    echo secret >secret-file.txt
    sops -e -i secret-file.txt

Add a generator to your kustomization:

    cat <<. >kustomization.yaml
    generators:
      - generator.yaml
    .

    cat <<. >generator.yaml
    apiVersion: kustomize.meiqia.com/v1beta1
    kind: SopsSecretGenerator
    metadata:
      name: my-secret
    envs:
      - secret-vars.env
    files:
      - secret-file.txt
    .
      
Run `kustomize build` with the `--enable_alpha_plugins` flag:

    kustomize build --enable_alpha_plugins
    
The output is a Kubernetes secret containing the decrypted data:

    apiVersion: v1
    data:
      FOO: c2VjcmV0
      secret-file.txt: c2VjcmV0Cg==
    kind: Secret
    metadata:
      name: my-secret-g8m5mh84c2

An example showing all options:

    apiVersion: kustomize.meiqia.com/v1beta1
    kind: SopsSecretGenerator
    metadata:
      name: my-secret
      labels:
        app: my-app
      annotations:
        create-by: me
    behavior: create
    disableNameSuffixHash: true
    envs:
      - secret-vars.env
      - secret-vars.yaml
      - secret-vars.json
    files:
      - secret-file1.txt
      - secret-file2.txt=secret-file2.sops.txt
    type: Oblique


## Development

You will need [Go](https://golang.org) 1.12 or higher to develop and build the plugin.


### Test

Run all tests:

    make test

In order to create encrypted test data, you need to import the secret key from `testdata/keyring.gpg` into
your GPG keyring once:

    gpg --import testdata/keyring.gpg
    
You can then use [sops](https://github.com/mozilla/sops) to create encrypted files:

    sops -e -i newfile.txt


### Build

Create a binary for your system:

    make
    
The resulting executable will be named `SopsSecretGenerator`.


### Release

First create a Git tag for the release:

    git tag -a v$VERSION

Then make releases for all supported platforms:

    make release
    
Binaries can be found in `releases`.
