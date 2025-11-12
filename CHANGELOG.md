# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release
- SSH remote command execution
- SFTP file operations (upload, download, list, mkdir, remove)
- Cross-platform password management using system keyring
- Command safety checks to prevent dangerous operations
- MCP (Model Context Protocol) support
- Sudo password auto-fill from keyring
- Environment variable support
- Comprehensive unit tests

### Changed

- Code refactored into modular structure
- Improved error handling

### Fixed

- N/A

## [1.0.0] - 2025-01-12

### Added

- Initial public release
- Core SSH/SFTP functionality
- Password management
- MCP support
- Safety checks
- Multi-platform support (Linux, macOS, Windows)

[Unreleased]: https://github.com/talkincode/sshmcp/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/talkincode/sshmcp/releases/tag/v1.0.0
