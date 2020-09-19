# kustomize-sopssecretgenerator

[![Build Status](https://travis-ci.org/goabout/kustomize-sopssecretgenerator.svg?branch=master)](https://travis-ci.org/goabout/kustomize-sopssecretgenerator)
[![Go Report Card](https://goreportcard.com/badge/github.com/goabout/kustomize-sopssecretgenerator)](https://goreportcard.com/report/github.com/goabout/kustomize-sopssecretgenerator)
[![Codecov](https://img.shields.io/codecov/c/github/goabout/kustomize-sopssecretgenerator)](https://codecov.io/gh/goabout/kustomize-sopssecretgenerator)
[![Latest Release](https://img.shields.io/github/v/release/goabout/kustomize-sopssecretgenerator?sort=semver)](https://github.com/goabout/kustomize-sopssecretgenerator/releases/latest)
[![License](https://img.shields.io/github/license/goabout/kustomize-sopssecretgenerator)](https://github.com/goabout/kustomize-sopssecretgenerator/blob/master/LICENSE)

SecretGenerator ❤ sops


## Why use this?

[Kustomize](https://github.com/kubernetes-sigs/kustomize) is a great tool for implementing a [GitOps](https://www.weave.works/blog/gitops-operations-by-pull-request) workflow. When a repository describes the entire system state, it often contains secrets that need to be encrypted at rest. Mozilla's [sops](https://github.com/mozilla/sops) is a simple and flexible tool that is very suitable for that task.

This Kustomize plugin allows you to create Secrets transparently from sops-encrypted files during resource generation. It is explicitly modeled after the builtin [SecretGenerator](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/plugins/builtins.md#secretgenerator) plugin. Because it is an exec plugin, it is not tied to the specific compilation of Kustomize, [like Go plugins are](https://github.com/kubernetes-sigs/kustomize/blob/master/docs/plugins/goPluginCaveats.md).


### Alternatives

There are a number of other plugins that can serve the same function:

* [viaduct-ai/kustomize-sops](https://github.com/viaduct-ai/kustomize-sops)
* [Agilicus/kustomize-sops](https://github.com/Agilicus/kustomize-sops)
* [barlik/kustomize-sops](https://github.com/barlik/kustomize-sops)
* [monopole/sopsencodedsecrets](https://github.com/monopole/sopsencodedsecrets)
* [omninonsense/kustomize-sopsgenerator](https://github.com/omninonsense/kustomize-sopsgenerator)
* [isindir/sops-secrets-operator](https://github.com/isindir/sops-secrets-operator)

Most of these projects are in constant development. I invite you to check them out and pick the project that best fits your goals.

Credit goes to [Seth Pollack](https://github.com/sethpollack) for the [Kustomize Secret Generator Plugins KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/kustomize-secret-generator-plugins.md) and subsequent implementation that made this possible.


## Installation

**Note:** Exec plugins do not seem to work on Windows at the moment. See issues [goabout/kustomize-sopssecretgenerator#14](https://github.com/goabout/kustomize-sopssecretgenerator/issues/14) and [kubernetes-sigs/kustomize#2924](https://github.com/kubernetes-sigs/kustomize/issues/2924).

Download the `SopsSecretGenerator` binary for your platform from the
[GitHub releases page](https://github.com/goabout/kustomize-sopssecretgenerator/releases) and
move it to `$XDG_CONFIG_HOME/kustomize/plugin/goabout.com/v1beta1/sopssecretgenerator`. (By default,
`$XDG_CONFIG_HOME` points to `$HOME/.config` on Linux and OS X, and `%LOCALAPPDATA%` on Windows.)

For example, to install version 1.3.2 on Linux:

    VERSION=1.3.2 PLATFORM=linux ARCH=amd64
    curl -Lo SopsSecretGenerator https://github.com/goabout/kustomize-sopssecretgenerator/releases/download/v${VERSION}/SopsSecretGenerator_${VERSION}_${PLATFORM}_${ARCH}
    chmod +x SopsSecretGenerator
    mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/goabout.com/v1beta1/sopssecretgenerator"
    mv SopsSecretGenerator "${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/goabout.com/v1beta1/sopssecretgenerator"

You do not need to install the `sops` binary for the plugin to work. The plugin includes and calls sops internally.


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
    apiVersion: goabout.com/v1beta1
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
      FOO: J3NlY3JldCc=
      secret-file.txt: c2VjcmV0Cg==
    kind: Secret
    metadata:
      name: my-secret-6d2fchb89d

Like SecretGenerator, SopsSecretGenerator supports the [generatorOptions](https://kubernetes-sigs.github.io/kustomize/api-reference/kustomization/generatoroptions/) fields. Data key-values ("envs") can be read from dotenv, YAML and JSON files. If the data is a file and the Secret data key needs to be different from the filename, you can use `key=file`.

An example showing all options:

    apiVersion: goabout.com/v1beta1
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


## Using SopsSecretsGenerator with ArgoCD

SopsSecretGenerator can be added to ArgoCD by [patching](./docs/argocd.md) an initContainer into the ArgoCD provided `install.yaml`.


## Development

You will need [Go](https://golang.org) 1.13 or higher to develop and build the plugin.


### Test

Run all tests:

    make test

In order to create encrypted test data, you need to import the secret key from `testdata/keyring.gpg` into
your GPG keyring once:

    cd testdata
    gpg --import keyring.gpg
    
You can then use `sops` to create encrypted files:

    sops -e -i newfile.txt


### Build

Create a binary for your system:

    make
    
The resulting executable will be named `SopsSecretGenerator`.


### Release

This project uses [goreleaser](https://goreleaser.com) to publish releases on GitHub.

First create a Git tag for the release:

    git tag -a v$VERSION

Then make releases for all supported platforms:

    make release

Binaries can be found in `dist`.

If everything looks good, set a GitHub personal token in the `GITHUB_TOKEN` environment variable
(or a file named `.github_token`) and publish the release to GitHub:
    
    export GITHUB_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    make publish-release
