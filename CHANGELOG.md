# Changelog

All notable fork changes are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Blacksmith quality checks execute declared Podman Makefile gates on an exact same-repository pull request head; a separate job publishes the candidate status after verifying the head is still current.

### Security

- Inherited Rancher release, FOSSA, Renovate, and test jobs run only in `rancher/machine`, preventing fork execution of upstream publication and Vault workflows.
