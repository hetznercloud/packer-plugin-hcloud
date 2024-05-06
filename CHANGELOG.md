# Changelog

## [1.4.0](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.3.0...v1.4.0) (2024-05-06)


### Features

* enable hcloud-go debug logging ([#167](https://github.com/hetznercloud/packer-plugin-hcloud/issues/167)) ([1ab544e](https://github.com/hetznercloud/packer-plugin-hcloud/commit/1ab544ebd1b4ad862f19ff2ec07c129feeef5434)), closes [#165](https://github.com/hetznercloud/packer-plugin-hcloud/issues/165)

## [1.3.0](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.2.1...v1.3.0) (2024-01-09)


### Features

* add labels options to server and ssh keys ([#128](https://github.com/hetznercloud/packer-plugin-hcloud/issues/128)) ([3f7dcae](https://github.com/hetznercloud/packer-plugin-hcloud/commit/3f7dcae20a07ad8367a06926aa2df1c415105e8b))
* use existing IP for server create ([#144](https://github.com/hetznercloud/packer-plugin-hcloud/issues/144)) ([1ebdfe7](https://github.com/hetznercloud/packer-plugin-hcloud/commit/1ebdfe74b395b2adaf015bc909a621aea2897a97))


### Bug Fixes

* do not pass nil error to error handler ([#145](https://github.com/hetznercloud/packer-plugin-hcloud/issues/145)) ([e742263](https://github.com/hetznercloud/packer-plugin-hcloud/commit/e74226397f420ecfd016c60e050ffc37fd3cddc8))
* improve logs messages and error handling ([#139](https://github.com/hetznercloud/packer-plugin-hcloud/issues/139)) ([2f2bcf1](https://github.com/hetznercloud/packer-plugin-hcloud/commit/2f2bcf1bca4aa46440e639907093095df67faf34))
* improve missing hcloud token error ([#138](https://github.com/hetznercloud/packer-plugin-hcloud/issues/138)) ([e47f476](https://github.com/hetznercloud/packer-plugin-hcloud/commit/e47f47679164d4ba700de64c4b44400604f152e2))

## [1.2.1](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.2.0...v1.2.1) (2023-11-08)


### Bug Fixes

* update integrations metadata identifier ([#119](https://github.com/hetznercloud/packer-plugin-hcloud/issues/119)) ([402c146](https://github.com/hetznercloud/packer-plugin-hcloud/commit/402c1464217f01f16d1262b8a7c5e525593e71e9)), closes [#95](https://github.com/hetznercloud/packer-plugin-hcloud/issues/95)

## [1.2.0](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.2.0-rc1...v1.2.0) (2023-11-08)


### Features

* transfer the packer plugin to the `hetznercloud` organization ([f3d1ed1](https://github.com/hetznercloud/packer-plugin-hcloud/commit/f3d1ed19f1596dc81561222f88f8edfddeec02a6))

## [1.2.0-rc1](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.2.0-rc0...v1.2.0-rc1) (2023-11-07)


### Build System

* rework build/release pipeline ([#103](https://github.com/hetznercloud/packer-plugin-hcloud/issues/103)) ([21c739b](https://github.com/hetznercloud/packer-plugin-hcloud/commit/21c739bb284377a13402cf47d1d720f0be00147b))

## [1.2.0-rc0](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.1.1...v1.2.0-rc0) (2023-11-02)


### Features

* receive repo transfer from hashicorp ([#95](https://github.com/hetznercloud/packer-plugin-hcloud/issues/95)) ([008082d](https://github.com/hetznercloud/packer-plugin-hcloud/commit/008082da523385a0ccc0956f594be33c3034eaf5))


### Build System

* setup release pipeline after transfer ([#100](https://github.com/hetznercloud/packer-plugin-hcloud/issues/100)) ([08c9695](https://github.com/hetznercloud/packer-plugin-hcloud/commit/08c96954b8fc451c363479d50854a78bb7052109))

## [1.1.1](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.1.0...v1.1.1) (2023-11-01)

### New Features

- Add networks in config to allow attaching server to private net by @magec in [#45](https://github.com/hetznercloud/packer-plugin-hcloud/pull/45)

### Other Changes

- makefile: remove old docs targets by @lbajolet-hashicorp in [#88](https://github.com/hetznercloud/packer-plugin-hcloud/pull/88)
- [COMPLIANCE] Add Copyright and License Headers by @hashicorp-copywrite in [#90](https://github.com/hetznercloud/packer-plugin-hcloud/pull/90)

### New Contributors

- @magec made their first contribution in [#45](https://github.com/hetznercloud/packer-plugin-hcloud/pull/45)

## [1.1.0](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.0.5...v1.1.0) (2023-09-25)

### New Features

- Change server type after create by @MarkusFreitag in [#62](https://github.com/hetznercloud/packer-plugin-hcloud/pull/62)
- feat: add support for ARM APIs by @apricote in [#75](https://github.com/hetznercloud/packer-plugin-hcloud/pull/75)
- chore(deps): migrate to hcloud-go v2 by @apricote in [#83](https://github.com/hetznercloud/packer-plugin-hcloud/pull/83)

### Bug fixes

- Update hcloud-go client to use Plugin version information by @apricote in [#70](https://github.com/hetznercloud/packer-plugin-hcloud/pull/70)

### Documentation improvements

- docs: fix references to internal documentation by @lbajolet-hashicorp in [#57](https://github.com/hetznercloud/packer-plugin-hcloud/pull/57)
- Update docs to be HCL2 by @nebloc in [#55](https://github.com/hetznercloud/packer-plugin-hcloud/pull/55)
- docs: add example for plugin by @apricote in [#74](https://github.com/hetznercloud/packer-plugin-hcloud/pull/74)
- doc: remove freebsd64 from rescue system list by @fsrv-xyz in [#73](https://github.com/hetznercloud/packer-plugin-hcloud/pull/73)

### Other Changes

- Update Plugin binary releases to match Packer by @nywilken in [#50](https://github.com/hetznercloud/packer-plugin-hcloud/pull/50)
- go.mod: run go mod tidy on go.mod/go.sum by @lbajolet-hashicorp in [#51](https://github.com/hetznercloud/packer-plugin-hcloud/pull/51)
- [COMPLIANCE] Update MPL 2.0 LICENSE by @hashicorp-copywrite in [#53](https://github.com/hetznercloud/packer-plugin-hcloud/pull/53)
- go.mod: bump go version from 1.17 to 1.18 by @lbajolet-hashicorp in [#60](https://github.com/hetznercloud/packer-plugin-hcloud/pull/60)
- Fix issues reported by Go checks by @nywilken in [#47](https://github.com/hetznercloud/packer-plugin-hcloud/pull/47)
- .gitignore: ignore .docs by @lbajolet-hashicorp in [#61](https://github.com/hetznercloud/packer-plugin-hcloud/pull/61)
- Bump github.com/hashicorp/packer-plugin-sdk from 0.3.1 to 0.4.0 by @dependabot in [#69](https://github.com/hetznercloud/packer-plugin-hcloud/pull/69)
- [COMPLIANCE] Add Copyright and License Headers by @hashicorp-copywrite in [#66](https://github.com/hetznercloud/packer-plugin-hcloud/pull/66)
- .gitignore: ignore crash.log by @lbajolet-hashicorp in [#76](https://github.com/hetznercloud/packer-plugin-hcloud/pull/76)
- cleanup github workflows by @lbajolet-hashicorp in [#78](https://github.com/hetznercloud/packer-plugin-hcloud/pull/78)
- bump go 1.18 to 1.19 by @lbajolet-hashicorp in [#81](https://github.com/hetznercloud/packer-plugin-hcloud/pull/81)
- Bump github.com/hashicorp/packer-plugin-sdk from 0.4.0 to 0.5.1 by @dependabot in [#82](https://github.com/hetznercloud/packer-plugin-hcloud/pull/82)
- Migration plugin docs to integration framework by @nywilken in [#84](https://github.com/hetznercloud/packer-plugin-hcloud/pull/84)
- version: prepare v1.1.0 release by @lbajolet-hashicorp in [#89](https://github.com/hetznercloud/packer-plugin-hcloud/pull/89)

### New Contributors

- @hashicorp-copywrite made their first contribution in [#53](https://github.com/hetznercloud/packer-plugin-hcloud/pull/53)
- @nebloc made their first contribution in [#55](https://github.com/hetznercloud/packer-plugin-hcloud/pull/55)
- @MarkusFreitag made their first contribution in [#62](https://github.com/hetznercloud/packer-plugin-hcloud/pull/62)
- @apricote made their first contribution in [#70](https://github.com/hetznercloud/packer-plugin-hcloud/pull/70)
- @fsrv-xyz made their first contribution in [#73](https://github.com/hetznercloud/packer-plugin-hcloud/pull/73)

## [1.0.5](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.0.4...v1.0.5) (2022-08-04)

### New Features

- Use SDK communicator to generate SSH key pair by @Feder1co5oave in [#39](https://github.com/hetznercloud/packer-plugin-hcloud/pull/39)

### Other Changes

- goreleaser: add missing target goos/goarch by @lbajolet-hashicorp in [#40](https://github.com/hetznercloud/packer-plugin-hcloud/pull/40)
- Bump github.com/hashicorp/packer-plugin-sdk from 0.2.13 to 0.3.1 by @dependabot in [#44](https://github.com/hetznercloud/packer-plugin-hcloud/pull/44)

### New Contributors

- @lbajolet-hashicorp made their first contribution in [#40](https://github.com/hetznercloud/packer-plugin-hcloud/pull/40)
- @Feder1co5oave made their first contribution in [#39](https://github.com/hetznercloud/packer-plugin-hcloud/pull/39)

## [1.0.4](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.0.3...v1.0.4) (2022-05-25)

This release contains the latest [golang.org/x/crypto/ssh](http://golang.org/x/crypto/ssh) module which implements client authentication support for signature algorithms based on SHA-2 for use with existing RSA keys. Previously, a client would fail to authenticate with RSA keys to servers that reject signature algorithms based on SHA-1.

### Bug fixes

- Bump packer-plugin-sdk to address legacy SSH key algorithms in SSH communicator

## [1.0.3](https://github.com/hetznercloud/packer-plugin-hcloud/compare/v1.0.2...v1.0.3) (2022-05-06)

### Other Changes

- goreleaser: auto-generate changelog file by @azr in [#29](https://github.com/hetznercloud/packer-plugin-hcloud/pull/29)
- Update release signing configuration by @nywilken in [#32](https://github.com/hetznercloud/packer-plugin-hcloud/pull/32)

## 1.0.2 (November 12, 2021)

- Add pre validate step check for snapshot names. Invoking packer build -force will bypass the checks, thus multiple snapshots with the same name will allowed. This is equivalent to the previous behavior of this plugin. [GH-17] [GH-18]
- Upgrade packer-plugin-sdk to v0.2.9. [GH-23]

## 1.0.1 (September 1, 2021)

- Add SSH key support for the Freebsd64 Rescue System. [GH-9]
- Bump to Go 1.17

## 1.0.0 (June 14, 2021)

- upgrade packer-plugin-sdk to v0.2.3. [GH-8]

## 0.0.1 (April 22, 2021)

- initial extraction of hcloud plugin from Packer core.
