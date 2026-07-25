# Changelog

All notable changes to tmux-attach-browser are documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-07-25

### Added

- Add a real terminal demo and a quick-start README.
- Add Dependabot, Dependency Review, CodeQL, secret scanning, push protection,
  and a protected `main` ruleset.
- Add compiled-binary, installer, real-tmux, source-length, vulnerability,
  workflow, and shell checks.

### Changed

- Pin the GoReleaser and Dependency Review actions to immutable commits.
- Document and enforce the 450-line limit for hand-written source files.

### Fixed

- Ignore stale asynchronous session refresh results.
- Cancel in-flight tmux list and create operations when the UI exits.

## [0.1.0] - 2026-07-19

### Added

- Browse, filter, select, create, attach to, and switch between tmux sessions.
- Install prebuilt macOS, Linux, and Windows binaries with checksum validation.

[0.1.1]: https://github.com/hmmhmmhm/tmux-attach-browser/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/hmmhmmhm/tmux-attach-browser/releases/tag/v0.1.0
