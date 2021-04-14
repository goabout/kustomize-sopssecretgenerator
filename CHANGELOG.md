# Changelog

## Version 1.4.0

* Build and release amd64 binaries.

## Version 1.3.3

* Update sops dependency to 3.7.1 to support [age](https://age-encryption.org/)-encrypted secrets.
* Migrate CI to GitHub Actions.

## Version 1.3.2

* Support files encrypted using sops versions 3.6.1 and 3.5. No longer supports
  sops 3.6.0, which is backward (and now forward) incompatible.

## Version 1.3.1

* Support files encrypted using sops 3.6.0.

## Version 1.3.0

* Support sops-encrypted INI files.

## Version 1.2.2

* Improve messages for sops errors.
* Document integration with [ArgoCD](https://github.com/argoproj/argo-cd). ([Leland Sindt](https://github.com/LelandSindt))
* Link to alternative plugins.

## Version 1.2.1

* Fix sops dependency.
* Use [goreleaser](https://goreleaser.com) for releases.

## Version 1.2.0

* Improved error handling. ([Timon Wong](https://github.com/timonwong))


## Version 1.1.0

* Renamed project to kustomize-sopssecretgenerator and binary to SopsSecretGenerator.
* Added tests.


## Version 1.0.0

* Initial release.
