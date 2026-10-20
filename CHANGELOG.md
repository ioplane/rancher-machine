# Changelog

All notable changes to this fork are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Optional current-request Create recovery capability over the existing RPC protocol, limited to driver-proved pending SSH custom-script creates; existing legacy and unsupported hosts retain the original refusal.
- Pinned Podman development checks for race tests, vet, build, module integrity, changed-code lint/format and vulnerability scanning.

### Changed

- Development lint uses pinned golangci-lint v2.15.0; one blocking gofumpt pass covers both recovery files and the existing three-file formatting scope.

- Fork pull request titles and descriptions and new commit messages use English; commit titles retain declarative Conventional Commits.

- Development toolchain pinned to Go 1.27.2 while retaining the inherited Go 1.26.0 module floor; the diagnostic RPC fallback log uses a constant format string for current vet checks.
- Crash-report printf forwarding preserves each variadic diagnostic argument under current Go vet checks.

### Security

- `golang.org/x/net` v0.61.0 and its required companion modules resolve reachable HTTP/2 advisories GO-2026-6617, GO-2026-6612, GO-2026-6611, GO-2026-6610 and GO-2026-6603; the scanner remains blocking over the full module.
