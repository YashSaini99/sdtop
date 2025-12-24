# Changelog

All notable changes to sdtop will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
### Changed
### Deprecated
### Removed
### Fixed
### Security

## [1.0.0] - 2024-12-24

### Added
- Initial MVP release
- Service list view with color-coded status indicators (● green = running, ✗ red = failed, ○ gray = stopped)
- Real-time log streaming from journald
- Log priority highlighting (errors in red, warnings in yellow)
- Process tree visualization (press 'p' key)
- Service control operations:
  - Start service (s key)
  - Stop service (t key)
  - Restart service (r key)
  - Enable on boot (e key)
  - Disable on boot (d key)
- Service filtering:
  - Show all services (1 key)
  - Show only running (2 key)
  - Show only stopped (3 key)
  - Show only failed (f key)
- Keyboard navigation (j/k or arrow keys)
- Context-aware help messages
- Empty state guidance for new users
- Auto-scrolling logs
- Process tree showing:
  - PIDs and parent-child relationships
  - Command lines
  - Process hierarchy

### Technical
- Built with Go 1.21+ and Bubble Tea framework
- Direct DBus integration with systemd
- Direct journald integration via go-systemd
- Process information from /proc filesystem
- Minimal dependencies for fast startup
- MVC architecture for maintainability

[Unreleased]: https://github.com/YashSaini99/sdtop/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/YashSaini99/sdtop/releases/tag/v1.0.0
