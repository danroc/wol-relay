# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.7] - 2025-07-06

### Fixed

- Call `network.String()` to fix log messages

## [0.1.6] - 2025-07-06

### Changed

- Don't call unnecessary `String()`

## [0.1.5] - 2025-07-06

### Changed

- Use `WithError` when logging errors

## [0.1.4] - 2025-07-06

### Added

- Add `build-docker` target
- Add link flags to build command
- Add unit tests for `main.go`
- Add Makefile
- Add comment before code block

### Changed

- Enhance log messages

### Fixed

- Fix lint errors

## [0.1.3] - 2025-06-19

### Added

- Add a license
- Add features, dev and standalone sections to README ([#9](https://github.com/danroc/wol-repeater/pull/9))
- Add unit tests to `wol` package ([#6](https://github.com/danroc/wol-repeater/pull/6))
- Add a CHANGELOG

### Changed

- Update CHANGELOG
- Reduce indentation by handling error first
- Update golang docker tag to v1.24.4 ([#10](https://github.com/danroc/wol-repeater/pull/10))
- Update standalone command comments
- Remove dummy example section from README
- Update changelog ([#8](https://github.com/danroc/wol-repeater/pull/8))
- Remove unreachable conditions ([#7](https://github.com/danroc/wol-repeater/pull/7))
- Update golang docker tag to v1.24.3 ([#5](https://github.com/danroc/wol-repeater/pull/5))

## [0.1.2] - 2025-05-06

### Added

- Add example to README.md
- Add Renovate configuration
- Add renovate.json ([#1](https://github.com/danroc/wol-relay/pull/1))

### Changed

- Update module golang.org/x/sys to v0.33.0 ([#4](https://github.com/danroc/wol-relay/pull/4))
- Improve comments to clarify WOL packet sending logic
- Update module golang.org/x/sys to v0.32.0 ([#3](https://github.com/danroc/wol-relay/pull/3))

### Fixed

- Check if remote is in a monitored network

## [0.1.1] - 2025-05-05

### Changed

- Format network in log message

## [0.1.0] - 2025-05-05

### Added

- Add Dockerfile
- Implement WOL relaying
- Add action to publish releases

### Changed

- Format log message
- Remove invalid newlines
- Don't export helper functions
- Fix typo
- Add `BuildPacket` function
- Add comment about the number of arguments
- Change buffer size to 1024
- Return errors in `wol.ParsePacket`
- Use logrus as logger
- Remove `main` binary
- Initial commit

[Unreleased]: https://github.com/danroc/wol-repeater/compare/v0.1.7...HEAD
[0.1.7]: https://github.com/danroc/wol-repeater/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/danroc/wol-repeater/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/danroc/wol-repeater/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/danroc/wol-repeater/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/danroc/wol-repeater/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/danroc/wol-repeater/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/danroc/wol-repeater/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/danroc/wol-repeater/releases/tag/v0.1.0
