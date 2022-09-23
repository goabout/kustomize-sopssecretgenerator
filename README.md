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

Since version 1.5.0, the plugin can be used as a [KRM Function](https://github.com/kubernetes-sigs/kustomize/blob/master/cmd/config/docs/api-conventions/functions-spec.md).

Credit goes to [Seth Pollack](https://github.com/sethpollack) for the [Kustomize Secret Generator Plugins KEP](https://github.com/kubernetes/enhancements/blob/master/keps/sig-cli/kustomize-secret-generator-plugins.md) and subsequent implementation that made this possible.


## Installation


SopsSecretGenerator is available as a binary, or as a Docker image.

### Binary

Download the `SopsSecretGenerator` binary for your platform from the [GitHub releases page](https://github.com/goabout/kustomize-sopssecretgenerator/releases) and make it executable.

For example, to install version 1.6.0 on Linux:
```bash
VERSION=1.6.0 PLATFORM=linux ARCH=amd64
curl -Lo SopsSecretGenerator "https://github.com/goabout/kustomize-sopssecretgenerator/releases/download/v${VERSION}/SopsSecretGenerator_${VERSION}_${PLATFORM}_${ARCH}"
chmod +x SopsSecretGenerator
```

You do not need to install the `sops` binary for the plugin to work. The plugin includes and calls sops internally.


### Docker image

See the [goabout/kustomize-sopssecretgenerator](https://hub.docker.com/repository/docker/goabout/kustomize-sopssecretgenerator) image at Docker Hub.


## Usage

Create some encrypted values using `sops`:
```bash
echo FOO=secret >secret-vars.env
sops -e -i secret-vars.env

echo secret >secret-file.txt
sops -e -i secret-file.txt
```


### Exec KRM Function

Although the generator can run in a Docker container, any real usage requires to access to local resources such as the filesystem or a PGP socket. This example calls the binary directly.

Add a generator to your kustomization:
```bash
cat <<. >kustomization.yaml
generators:
  - generator.yaml
.

cat <<. >generator.yaml
apiVersion: goabout.com/v1beta1
kind: SopsSecretGenerator
metadata:
  annotations:
   config.kubernetes.io/function: |
      exec:
        path: ./SopsSecretGenerator
  name: my-secret
envs:
  - secret-vars.env
files:
  - secret-file.txt
.
```

(Change the path to the `SopsSecretGenerator` binary to suit your installation. Kustomize will use the binary search path, `$PATH`, if you use a bare command.)

Run `kustomize build` with the `--enable-alpha-plugins` and `--enable-exec` flags:

```bash
kustomize build --enable-alpha-plugins --enable-exec
```
    
The output is a Kubernetes secret containing the decrypted data:
```yaml
apiVersion: v1
data:
  FOO: J3NlY3JldCc=
  secret-file.txt: c2VjcmV0Cg==
kind: Secret
metadata:
  name: my-secret-6d2fchb89d
```


### Legacy Plugin

First, install the plugin to `$XDG_CONFIG_HOME`: (By default, `$XDG_CONFIG_HOME` points to `$HOME/.config` on Linux and OS X, and `%LOCALAPPDATA%` on Windows.)
```bash
mkdir -p "${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/goabout.com/v1beta1/sopssecretgenerator"
mv SopsSecretGenerator "${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/goabout.com/v1beta1/sopssecretgenerator"
```

Add a generator to your kustomization:
```bash
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
```


### Generator Options

Like SecretGenerator, SopsSecretGenerator supports the [generatorOptions](https://kubernetes-sigs.github.io/kustomize/api-reference/kustomization/generatoroptions/) fields. Additionally, labels and annotations are copied over to the Secret. Data key-values ("envs") can be read from dotenv, INI, YAML and JSON files. If the data is a file and the Secret data key needs to be different from the filename, you can specify the key by adding `desiredKey=filename` instead of just the filename.

An example showing all supported SecretGenerator options:
```yaml
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
  - secret-vars.ini
  - secret-vars.yaml
  - secret-vars.json
files:
  - secret-file1.txt
  - secret-file2.txt=secret-file2.sops.txt
type: Opaque
```

In addition to the `envs` and `files` SecretGenerator facilities, SopsSecretGenerator also supports sops decryption of entire objects as part of a single invocation, acting as a pipeline that can pass through arbitrary encrypted data as individual objects to kustomize or other KRM pipelines. This allows decrypting of other forms of static secret data, and can function as a more secure version of the SecretGenerator `literals` facility and allow management of more exotic secrets and other arbitrary type you need to encrypt.

For example, first create an encrypted Secret with sops:
```bash
kubectl create secret --dry-run ... >mysecret.yaml
sops -e -i --encrypted-regex='^data$' mysecret.yaml
```

Then create a SopsSecretGenerator that will decrypt it:
```yaml
apiVersion: goabout.com/v1beta1
kind: SopsSecretGenerator
metadata:
  name: passthrough
objects:
  - mysecret.yaml
  - encryptedConfigMap.yaml
```

The result of running SopsSecretGenerator in this case would be the decrypted versions of `mysecret.yaml` and `encryptedConfigMap.yaml`.

The `objects` section can be used along side `envs` and `files` as well:
```yaml
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
  - secret-vars.ini
  - secret-vars.yaml
  - secret-vars.json
files:
  - secret-file1.txt
  - secret-file2.txt=secret-file2.sops.txt
objects:
  - otherSecret.yaml
  - encryptedConfigMap.yaml
type: Opaque
```

This will generate the Secret called `my-secret` and also append decrypted copies of `otherSecret.yaml` and `encryptedConfigMap.yaml`

## Using SopsSecretsGenerator with ArgoCD

SopsSecretGenerator can be added to ArgoCD by [patching](./docs/argocd.md) an initContainer into the ArgoCD provided `install.yaml`.


## Alternatives

There are a number of other plugins that can serve the same function:

* [viaduct-ai/kustomize-sops](https://github.com/viaduct-ai/kustomize-sops)
* [Agilicus/kustomize-sops](https://github.com/Agilicus/kustomize-sops)
* [barlik/kustomize-sops](https://github.com/barlik/kustomize-sops)
* [monopole/sopsencodedsecrets](https://github.com/monopole/sopsencodedsecrets)
* [omninonsense/kustomize-sopsgenerator](https://github.com/omninonsense/kustomize-sopsgenerator)
* [whatever-company/secretgen](https://github.com/whatever-company/secretgen)

Additionally, there are other ways to use sops-encrypted secrets in Kubernetes:

* [isindir/sops-secrets-operator](https://github.com/isindir/sops-secrets-operator)
* [craftypath/sops-operator](https://github.com/craftypath/sops-operator)
* [jkroepke/helm-secrets](https://github.com/jkroepke/helm-secrets)
* [dschniepp/sealit](https://github.com/dschniepp/sealit)

Most of these projects are in constant development. I invite you to check them out and pick the project that best fits your goals.


## Development

You will need [Go](https://golang.org) 1.17 or higher to develop and build the plugin.


### Test

Run all tests:

    make test

In order to create encrypted test data, you need to import the secret key from `testdata/keyring.gpg` into your GPG keyring once:

    cd testdata
    gpg --import keyring.gpg
    
You can then use `sops` to create encrypted files:

    sops -e -i newfile.txt


### Build

Create a binary for your system:

    make
    
The resulting executable will be named `SopsSecretGenerator`.


### Release

This project uses GitHub Actions and [goreleaser](https://goreleaser.com) to publish releases on GitHub.

First, don't forget to update the documentation for the new version you are going to release.

Then create a Git tag for the release:

    VERSION=X.X.X
    git tag -a v$VERSION -m "Version $VERSION"

And push it to GitHub:

    git push

The GitHub Actions workflow will build and release the binaries automatically.
