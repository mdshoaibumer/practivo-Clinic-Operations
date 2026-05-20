# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2026-05-21

### Added
- About tab in Settings with app version display
- In-app update checker (checks GitHub Releases, one-click download & install)
- Toast notifications for update status (success/failure)
- Loading spinners with animated icons during update operations
- 12 unit tests for About tab component

### Improved
- Accessibility: ARIA attributes (aria-busy, aria-live, role, aria-label) on update UI
- Better error messages for network failures during update check

### Fixed
- Phone validation for "91" prefix formats (091, 0091, spaces, dashes)
- BackupHandler nil pointer dereference when cloud not configured

## [1.0.0] - 2026-05-20

### Added
- Patient management (CRUD, search, treatment history)
- Appointment scheduling (daily/weekly views, conflict detection)
- Invoice & billing (GST support, payment tracking, outstanding balances)
- Dashboard with revenue stats and daily/monthly reports
- Treatment catalog management
- Local backup & restore
- Session-based authentication with login rate limiting
- Clinic setup wizard
- Logo upload for invoices
