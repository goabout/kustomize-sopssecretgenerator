# kustomize-sopssecret-plugin

[![Build Status](https://travis-ci.org/goabout/kustomize-sopssecret-plugin.svg?branch=master)](https://travis-ci.org/goabout/kustomize-sopssecret-plugin)
[![Go Report Card](https://goreportcard.com/badge/github.com/goabout/kustomize-sopssecret-plugin)](https://goreportcard.com/report/github.com/goabout/kustomize-sopssecret-plugin)
[![Codecov](https://img.shields.io/codecov/c/github/goabout/kustomize-sopssecret-plugin)](https://codecov.io/gh/goabout/kustomize-sopssecret-plugin)
[![Dependencies](https://img.shields.io/librariesio/github/goabout/kustomize-sopssecret-plugin)](https://libraries.io/github/goabout/kustomize-sopssecret-plugin)
[![Latest Release](https://img.shields.io/github/v/release/goabout/kustomize-sopssecret-plugin?sort=semver)](https://github.com/goabout/kustomize-sopssecret-plugin/releases/latest)
[![License](https://img.shields.io/github/license/goabout/kustomize-sopssecret-plugin)](https://github.com/goabout/kustomize-sopssecret-plugin/blob/master/LICENSE)

An exec plugin for [kustomize](https://github.com/kubernetes-sigs/kustomize)
that generates Secrets from files encrypted with [sops](https://github.com/mozilla/sops).


## Installation

Download the `SopsSecret` binary for your platform from the
[GitHub releases page](https://github.com/goabout/kustomize-sopssecret-plugin/releases) and
move it to `$XDG_CONFIG_HOME/kustomize/plugin/sopssecret`. (By default,
`$XDG_CONFIG_HOME` points to `$HOME/.config` on Linux and OS X and `%LOCALAPPDATA%` on Windows.)

For example, to install version 1.0.0 on Linux:

    VERSION=1.0.0 PLATFORM=linux ARCH=amd64
    wget https://github.com/goabout/kustomize-sopssecret-plugin/releases/download/v${VERSION}/SopsSecret_${VERSION}_${PLATFORM}_${ARCH} -O SopsSecret
    chmod +x SopsSecret
    mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/sopssecret"
    mv SopsSecret "${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/sopssecret"


## Usage

Create some encrypted values using `sops`:

    echo FOO=secret >secret-vars.env
    sops -e -i secret-vars.env
    
    echo secret >secret-file.txt
    sops -e -i secret-file.txt

Add a generator resources to your kustomization:

    cat <<. >kustomization.yaml
    generators:
      - generator.yaml
    .

    cat <<. >generator.yaml
    apiVersion: goabout.com/v1beta1
    kind: SopsSecret
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
      annotations:
        kustomize.config.k8s.io/needs-hash: "true"
      name: my-secret

The `kustomize.config.k8s.io/needs-hash` annotation uses a feature from
[kustomize #1473](https://github.com/kubernetes-sigs/kustomize/pull/1473) to add the content
hash as a suffix to the Secret name, just as the builtin secretGenerator plugin does.
If/when that PR is merged, annotations generated when using the `behavior` and
`disableNameSuffixHash` options will work as expected.

An example showing all options:

    apiVersion: goabout.com/v1beta1
    kind: SopsSecret
    metadata:
      name: my-secret
      labels:
        app: my-app
      annotations:
        create-by: me
    behavior: merge
    disableNameSuffixHash: true
    envs:
      - secret-vars.env
      - secret-vars.yaml
      - secret-vars.json
    files:
      - secret-file1.txt
      - secret-file2.txt=secret-file2.sops.txt
    type: Oblique


## Build

You will need [Go](https://golang.org) 1.12 or higher to build the plugin.

Run all tests:

    make test

Create a binary for your system:

    make
    
The resulting executable will be named `SopsPlugin`.

Make a release for all supported platforms:

    make release
    
Binaries can be found in `releases`.
