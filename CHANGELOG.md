# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Uncategorized

- chore: remove unreachable conditions ([#7](https://github.com/danroc/wol-relay/pull/7))
- test: add unit tests to `wol` package ([#6](https://github.com/danroc/wol-relay/pull/6))
- chore: add a CHANGELOG
- chore(deps): update golang docker tag to v1.24.3 ([#5](https://github.com/danroc/wol-relay/pull/5))

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

[Unreleased]: https://github.com/danroc/wol-relay/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/danroc/wol-relay/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/danroc/wol-relay/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/danroc/wol-relay/releases/tag/v0.1.0
