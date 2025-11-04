# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Local Folder Sync**: Complete support for syncing local folders with cloud storage
  - Sync configuration manager (`internal/syncconfig`) for managing local-to-remote mappings
  - Support for multiple sync pairs (local folder ↔ remote location)
  - Three sync directions: upload (local→remote), download (remote→local), bidirectional
  - Enable/disable individual sync pairs
  - Sync all enabled pairs with a single command
  - Dry-run mode for previewing changes
- **Rclone Extensions**: New methods for local folder operations
  - `SyncLocalToRemote()`: Upload local folder to cloud
  - `SyncRemoteToLocal()`: Download from cloud to local folder
  - `ListLocalFiles()`: Preview local files
  - `GetLocalDirSize()`: Get directory size information
- **Backup Manager API**: High-level methods for sync pair management
  - `AddSyncPair()`: Add new local-to-remote sync configuration
  - `RemoveSyncPair()`: Remove sync pair by name
  - `ListSyncPairs()`: List all configured sync pairs
  - `ToggleSyncPair()`: Enable/disable sync pairs
  - `SyncPair()`: Execute sync for specific pair
  - `SyncAllEnabled()`: Sync all enabled pairs
- **Shell Scripts**: Templates for local folder syncing
  - `sync_local_folder.sh`: Sync individual sync pair
  - `sync_all_folders.sh`: Sync all enabled pairs with summary
- **UI Components**: Sync pairs management view
  - Interactive TUI for adding/removing sync pairs
  - Visual status indicators (enabled/disabled)
  - Step-by-step configuration wizard
- **Documentation**:
  - Comprehensive local folder sync guide (`docs/LOCAL_FOLDER_SYNC.md`)
  - Usage examples (`examples/sync_local_folder_example.go`)
  - Updated README with new features
- **Tests**: Complete test coverage for sync configuration
  - Unit tests for all sync config operations
  - Validation tests for sync pairs and paths
  - Test coverage for enable/disable/toggle operations

### Changed
- Updated backup manager to include sync configuration manager
- Enhanced rclone manager with local folder support
- Updated project structure documentation

### Fixed
- Test name conflict between syncconfig and rclone tests

## [0.1.0-dev] - 2025-11-03

### Added
- Initial project structure
- Core backup functionality (cloud-to-cloud)
- LaunchAgent automation for macOS
- Log viewer and monitoring
- Installation wizard
- Configuration wizard
- Script generation
- Lockfile management

---

## Migration Guide

### Upgrading to Local Folder Sync

If you're upgrading from a version without local folder sync:

1. **No breaking changes**: Existing cloud-to-cloud backup functionality remains unchanged
2. **New configuration file**: Sync pairs are stored in `~/.config/cloud-sync/sync-config.json`
3. **Optional feature**: Local folder sync is completely optional and doesn't affect existing backups
4. **New dependencies**: Requires `jq` for shell script JSON parsing (install with `brew install jq`)

### Example Migration

```go
// Old: Cloud-to-cloud backup only
manager, _ := backup.NewManager(config)

// New: Same cloud-to-cloud backup + local folder sync
manager, _ := backup.NewManager(config)
manager.AddSyncPair("docs", "/Users/me/Documents", "remote", "bucket/docs", "upload")
```

---

## Commit History

### November 3, 2025

- `dc905cc` - fix(test): rename TestConfigExists to avoid conflict with rclone tests
- `f0c64df` - docs: add example code for local folder sync usage
- `a483e25` - docs: update README with local folder sync features
- `05e64b6` - docs: add comprehensive guide for local folder sync feature
- `2164262` - feat(scripts): add templates for syncing local folders with remotes
- `6d33707` - feat(ui): add sync pairs management view for configuring local folders
- `fc591a1` - feat(backup): add sync pair management and local folder sync support
- `a88bd84` - feat(rclone): add local-to-remote and remote-to-local sync methods
- `193c4da` - test(syncconfig): add comprehensive unit tests for sync configuration
- `6d3e089` - feat(syncconfig): add sync configuration manager for local folder mappings
