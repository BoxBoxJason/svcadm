# svcadm CHANGELOG

## [GoLang Migration] - 2024-09-16

### Breaking Changes
- Migrated the project to GoLang instead of zsh
- Changed the project structure to be more modular
- Removed all of the scripts and replaced them with GoLang code / binaries
- Removed the manual installation parameters specification, replaced it with a single yaml configuration file

### Added
- GitLab service
- Mattermost service
- VaultWarden service
- Users are specified in a file and are created automatically on the first run

## [Clamav] - 2024-09-03

### Added
- Clamav service

## [Minio] - 2024-08-30

### Added
- Minio service

## [Safeguard] - 2024-08-28

### Added
- Safeguard script to monitor directories and files for changes, checking for infections

### Changed
- Updated the autocomplete for each service

## [First Services] - 2024-08-26

### Added
- Initial services:
  - PostgreSQL
  - SonarQube
  - Nginx
- Documentation:
  - README.md
  - CHANGELOG.md
