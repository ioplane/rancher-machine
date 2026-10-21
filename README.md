# Rancher Machine, a fork of [Docker Machine](https://github.com/docker/machine)

Machine lets you create Docker hosts on your computer, on cloud providers, and
inside your own data center. It creates servers, installs Docker on them, then
configures the Docker client to talk to them.

## Installation and documentation
Fork development and recovery boundaries: [Create recovery](docs/ru/create-recovery.md).

The original full Docker Machine documentation [is available here](https://gcbw.github.io/docker.github.io/machine/).

This project is intended to be embedded and executed by the full [Rancher](https://github.com/rancher/rancher) product
and the stand alone cli functionality will remain but the human use of it will not be the primary focus as we will expect
inputs provided by other things like Terraform or UIs.

Cli binaries can be found in our [Releases Pages](https://github.com/rancher/machine/releases)

## Fork contribution language

Pull request titles and descriptions and all new commit messages use English.
Commit titles follow Conventional Commits with declarative noun phrases.

## Issues

For historical context you can read the [Docker Machine Issues](https://github.com/docker/machine/issues)
but all new issues created for Rancher Machine will need to be created 
in [Rancher](https://github.com/rancher/rancher/issues) 

## Driver Plugins

In addition to the core driver plugins bundled alongside Rancher Machine, users
can make and distribute their own plugin for any virtualization technology or
cloud provider.  To browse the list of known Rancher Machine plugins, please [see
this document in our
docs repo](https://github.com/docker/docker.github.io/blob/master/machine/AVAILABLE_DRIVER_PLUGINS.md).

## Branching and Versioning

The `master` branch is used for ongoing development and always contains the latest release. 
Tags created from `master` are consumed by Rancher’s `main` branch.

In addition to `master`, this repository maintains multiple release branches. 
Each branch name corresponds to the Rancher release line that consumes tags cut from that branch.

Whenever a new Rancher release branch is created, a corresponding branch is also created in this repository from `master`.

After a release branch is created, it only receives bug fixes and security patches. 
New tags created from a release branch increment only the `.x` suffix of the version number. 
For example, at the time of writing, `v0.15.0-rancher142.2` is the latest tag in the `release/v2.14` branch. 
Subsequent tags from that branch would be `v0.15.0-rancher142.3`, `v0.15.0-rancher142.4`, and so on.

| Machine Branch | Rancher Releae line |
|----------------|---------------------|
| master         | main                |
| release/v2.14  | v2.14               |
| release/v2.13  | v2.13               |

## Releasing a New Version

- **Prerequisite:**

	- Ensure all release changes are already merged into the `master` branch before creating a tag. This process creates the tag from the latest `master` commit.

- **Prepare `master` and remote:**

	- Choose your release remote (replace `<remote>`):

		- `git fetch <remote> --tags`
		- `git checkout master`
		- `git pull --ff-only <remote> master`

- **Verify the latest tag (locally):**

	- `git describe --tags --abbrev=0`

- **Create tag (from latest `master`):**

	- `git tag v0.15.0-rancher<NEW-VERSION>`

- **Push the tag to your remote:**

	- Push single tag: `git push <remote> v0.15.0-rancher<NEW-VERSION>`

- **Verify the tag is on the remote:**

	- `git ls-remote --tags <remote> | grep v0.15.0-rancher<NEW-VERSION>`

## Fork quality checks

In `ioplane/rancher-machine`, `.github/workflows/fork-blacksmith.yml` runs the
candidate's `make dev-build lint-build` and `make check` on Blacksmith using
rootless Podman. The candidate must declare these targets; missing targets fail
the check. `pull_request_target` accepts only open pull requests whose head and
base repositories are this fork. Manual dispatch requires a pull request number.
The workflow resolves and checks out the exact head SHA with read-only permissions
and no persisted checkout credentials, secrets, environments, or shared caches.
A separate job checks that the pull request still has that head and writes the
`Blacksmith / fork quality` commit status without checking out candidate code.

The inherited Rancher test, release, FOSSA, and Renovate workflows are restricted
to `rancher/machine`; they do not publish releases or request Rancher credentials
from this fork. Fork CI activation requires this workflow on the default branch
and available Blacksmith runners. No production infrastructure is provisioned.
## Development checks

As of 2026-10-21, the development toolchain declarations select Go 1.27.3
for [GO-2026-6809](https://pkg.go.dev/vuln/GO-2026-6809), retaining the Go 1.26.0
module floor. `make check-build` logs the built binary's compiler and module
metadata with `go version -m`. This source change does not fix the SUSE vendor
release compiler in `package/Dockerfile`: that path remains affected and unfixed,
with release acceptance blocked until a supported vendor fix and an actual build
are verified, or a separate builder-channel migration is reviewed.

The pinned Podman lint image provides golangci-lint v2.15.0. `make check-format`
uses one blocking gofumpt pass for the recovery files and the existing three-file
formatting scope; `make check-lint` retains the configured changed-code filter.
