# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.11] - 2025-11-11

### Changed

- Update dependencies
  - Update module golang.org/x/sys to v0.38.0 ([#22](https://github.com/danroc/wol-relay/pull/22))
  - Update golang docker tag to v1.25.4 ([#21](https://github.com/danroc/wol-relay/pull/21))
  - Update golang docker tag to v1.25.3 ([#20](https://github.com/danroc/wol-relay/pull/20))
  - Update module golang.org/x/sys to v0.37.0 ([#18](https://github.com/danroc/wol-relay/pull/18))
  - Update golang docker tag to v1.25.2 ([#19](https://github.com/danroc/wol-relay/pull/19))
  - Update golang docker tag to v1.25.1 ([#17](https://github.com/danroc/wol-relay/pull/17))
  - Update golang docker tag to v1.25.0 ([#16](https://github.com/danroc/wol-relay/pull/16))
  - Update actions/checkout action to v5 ([#15](https://github.com/danroc/wol-relay/pull/15))
  - Update module golang.org/x/sys to v0.35.0 ([#14](https://github.com/danroc/wol-relay/pull/14))
  - Update golang docker tag to v1.24.6 ([#13](https://github.com/danroc/wol-relay/pull/13))

## [0.1.10] - 2025-07-14

### Fixed

- Continue main loop on errors ([94b91e7](https://github.com/danroc/wol-repeater/commit/94b91e7))

## [0.1.9] - 2025-07-12

### Changed

- Improve log field names ([c96941f](https://github.com/danroc/wol-repeater/commit/c96941f))
- Update dependencies
  - Update golang docker tag to v1.24.5 ([#11](https://github.com/danroc/wol-repeater/pull/11))
  - Update module golang.org/x/sys to v0.34.0 ([#12](https://github.com/danroc/wol-repeater/pull/12))

## [0.1.8] - 2025-07-06

### Changed

- Rename `remote` to `source` in log messages ([15a8cbb](https://github.com/danroc/wol-repeater/commit/15a8cbb))

## [0.1.7] - 2025-07-06

### Fixed

- Call `network.String()` to fix log messages ([a362af2](https://github.com/danroc/wol-repeater/commit/a362af2))

## [0.1.6] - 2025-07-06

### Fixed

- Don't call unnecessary `String()` ([5dda3a3](https://github.com/danroc/wol-repeater/commit/5dda3a3))

## [0.1.5] - 2025-07-06

### Added

- Add error field to log messages ([dd73bde](https://github.com/danroc/wol-repeater/commit/dd73bde))

## [0.1.4] - 2025-07-06

### Added

- Add context to log messages ([490fc69](https://github.com/danroc/wol-repeater/commit/490fc69))
- Add Docker build target ([fba71e2](https://github.com/danroc/wol-repeater/commit/fba71e2))
- Add link flags to build command ([1b95620](https://github.com/danroc/wol-repeater/commit/1b95620))
- Add unit tests for `main.go` ([8052766](https://github.com/danroc/wol-repeater/commit/8052766))
- Add Makefile ([36f43d7](https://github.com/danroc/wol-repeater/commit/36f43d7))

### Fixed

- Fix lint errors ([1a3ed98](https://github.com/danroc/wol-repeater/commit/1a3ed98))

## [0.1.3] - 2025-06-19

### Added

- Add a license ([7c5db8a](https://github.com/danroc/wol-repeater/commit/7c5db8a))
- Add a CHANGELOG ([c236c51](https://github.com/danroc/wol-repeater/commit/c236c51))
- Add features, dev and standalone sections to README ([#9](https://github.com/danroc/wol-repeater/pull/9))
- Add unit tests to `wol` package ([#6](https://github.com/danroc/wol-repeater/pull/6))

### Changed

- Reduce indentation by handling error first ([467e3cb](https://github.com/danroc/wol-repeater/commit/467e3cb))
- Update dependencies
  - Update golang docker tag to v1.24.4 ([#10](https://github.com/danroc/wol-repeater/pull/10))
  - Update golang docker tag to v1.24.3 ([#5](https://github.com/danroc/wol-repeater/pull/5))
- Remove dummy example section from README ([3f1f1bd](https://github.com/danroc/wol-repeater/commit/3f1f1bd))

### Fixed

- Remove unreachable conditions ([#7](https://github.com/danroc/wol-repeater/pull/7))

## [0.1.2] - 2025-05-06

### Added

- Add usage example to README.md ([858f394](https://github.com/danroc/wol-repeater/commit/858f394))
- Add Renovate
  - Add Renovate configuration ([44822f6](https://github.com/danroc/wol-repeater/commit/44822f6))
  - Add renovate.json ([#1](https://github.com/danroc/wol-repeater/pull/1))

### Changed

- Update dependencies
  - Update module golang.org/x/sys to v0.33.0 ([#4](https://github.com/danroc/wol-repeater/pull/4))
  - Update module golang.org/x/sys to v0.32.0 ([#3](https://github.com/danroc/wol-repeater/pull/3))
- Improve comments to clarify WOL packet sending logic ([26703f4](https://github.com/danroc/wol-repeater/commit/26703f4))

### Fixed

- Check if remote is in a monitored network ([abad8d8](https://github.com/danroc/wol-repeater/commit/abad8d8))

## [0.1.1] - 2025-05-05

### Changed

- Format network in log message ([3ff6afb](https://github.com/danroc/wol-repeater/commit/3ff6afb))

## [0.1.0] - 2025-05-05

### Added

- Add Dockerfile ([fb66a97](https://github.com/danroc/wol-repeater/commit/fb66a97))
- Implement WOL relaying ([bdc9b73](https://github.com/danroc/wol-repeater/commit/bdc9b73))
- Add action to publish releases ([700181a](https://github.com/danroc/wol-repeater/commit/700181a))
- Initial commit ([a3e5665](https://github.com/danroc/wol-repeater/commit/a3e5665))

### Changed

- Format log message ([bf2248e](https://github.com/danroc/wol-repeater/commit/bf2248e))
- Remove invalid newlines ([604f805](https://github.com/danroc/wol-repeater/commit/604f805))
- Don't export helper functions ([14942f2](https://github.com/danroc/wol-repeater/commit/14942f2))
- Add `BuildPacket` function ([db3fa9c](https://github.com/danroc/wol-repeater/commit/db3fa9c))
- Change buffer size to 1024 ([9efb630](https://github.com/danroc/wol-repeater/commit/9efb630))
- Return errors in `wol.ParsePacket` ([75725c5](https://github.com/danroc/wol-repeater/commit/75725c5))
- Use logrus as logger ([8739a5b](https://github.com/danroc/wol-repeater/commit/8739a5b))
- Remove `main` binary ([a9dc4e4](https://github.com/danroc/wol-repeater/commit/a9dc4e4))

[Unreleased]: https://github.com/danroc/wol-relay/compare/v0.1.11...HEAD
[0.1.11]: https://github.com/danroc/wol-relay/compare/v0.1.10...v0.1.11
[0.1.10]: https://github.com/danroc/wol-relay/compare/v0.1.9...v0.1.10
[0.1.9]: https://github.com/danroc/wol-relay/compare/v0.1.8...v0.1.9
[0.1.8]: https://github.com/danroc/wol-relay/compare/v0.1.7...v0.1.8
[0.1.7]: https://github.com/danroc/wol-relay/compare/v0.1.6...v0.1.7
[0.1.6]: https://github.com/danroc/wol-relay/compare/v0.1.5...v0.1.6
[0.1.5]: https://github.com/danroc/wol-relay/compare/v0.1.4...v0.1.5
[0.1.4]: https://github.com/danroc/wol-relay/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/danroc/wol-relay/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/danroc/wol-relay/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/danroc/wol-relay/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/danroc/wol-relay/releases/tag/v0.1.0
