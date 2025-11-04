# Local Folder Sync

This document describes how to configure and use local folder synchronization with cloud storage.

## Overview

Cloud-sync now supports syncing local folders with cloud storage remotes. You can configure multiple sync pairs, each representing a local folder that syncs with a specific remote location.

## Features

- **Multiple sync pairs**: Configure as many local-to-remote sync pairs as needed
- **Bidirectional sync**: Support for upload, download, or bidirectional synchronization
- **Enable/disable**: Toggle individual sync pairs on or off
- **Automated sync**: Sync all enabled pairs with a single command
- **Dry-run mode**: Preview changes before applying them

## Configuration

### Sync Pair Structure

Each sync pair consists of:

- **Name**: A unique identifier for the sync pair
- **Local Path**: The local folder to sync
- **Remote Name**: The rclone remote name (must be configured in rclone)
- **Remote Path**: The path on the remote (bucket/folder)
- **Direction**: One of:
  - `upload`: Local → Remote (one-way sync from local to cloud)
  - `download`: Remote → Local (one-way sync from cloud to local)
  - `bidirectional`: Both ways (sync in both directions)
- **Enabled**: Whether this sync pair is active

### Configuration File

Sync pairs are stored in `~/.config/cloud-sync/sync-config.json`:

```json
{
  "version": "1.0",
  "sync_pairs": [
    {
      "name": "documents",
      "local_path": "/Users/username/Documents",
      "remote_name": "backblaze",
      "remote_path": "my-bucket/documents",
      "direction": "upload",
      "enabled": true
    },
    {
      "name": "photos",
      "local_path": "/Users/username/Pictures",
      "remote_name": "s3",
      "remote_path": "photo-backup/pictures",
      "direction": "bidirectional",
      "enabled": true
    }
  ]
}
```

## Usage

### Using the Go API

```go
import (
    "github.com/andreisuslov/cloud-sync/pkg/backup"
    "github.com/andreisuslov/cloud-sync/internal/syncconfig"
)

// Create backup manager
config := &backup.Config{
    Username:     "myuser",
    HomeDir:      "/Users/myuser",
    SourceRemote: "backblaze",
    SourceBucket: "source-bucket",
    DestRemote:   "s3",
    DestBucket:   "dest-bucket",
    RclonePath:   "/usr/local/bin/rclone",
}

manager, err := backup.NewManager(config)
if err != nil {
    panic(err)
}

// Add a sync pair
err = manager.AddSyncPair(
    "documents",                    // name
    "/Users/myuser/Documents",      // local path
    "backblaze",                    // remote name
    "my-bucket/documents",          // remote path
    "upload",                       // direction
)

// List all sync pairs
pairs, err := manager.ListSyncPairs()

// Sync a specific pair
err = manager.SyncPair("documents", true, false) // progress=true, dryRun=false

// Sync all enabled pairs
err = manager.SyncAllEnabled(true, false)

// Toggle a sync pair
err = manager.ToggleSyncPair("documents")

// Remove a sync pair
err = manager.RemoveSyncPair("documents")
```

### Using Shell Scripts

#### Sync a specific folder

```bash
# Sync a specific sync pair
~/bin/sync_local_folder.sh documents

# Dry-run mode (preview changes)
~/bin/sync_local_folder.sh documents --dry-run
```

#### Sync all enabled folders

```bash
# Sync all enabled sync pairs
~/bin/sync_all_folders.sh

# Dry-run mode
~/bin/sync_all_folders.sh --dry-run
```

### Using the TUI (Terminal UI)

The TUI provides an interactive interface for managing sync pairs:

1. Launch the application
2. Navigate to "Sync Pairs Management"
3. Press `a` to add a new sync pair
4. Follow the prompts to configure:
   - Name
   - Local path
   - Remote name
   - Remote path
   - Direction
5. Press `t` to toggle a sync pair on/off
6. Press `d` to delete a sync pair

## Sync Directions Explained

### Upload (Local → Remote)

Files are synced from your local folder to the cloud. The remote becomes an exact copy of the local folder.

**Use case**: Backing up local files to the cloud

```bash
# Example: Backup documents to cloud
Local:  /Users/username/Documents
Remote: backblaze:my-bucket/documents
Direction: upload
```

### Download (Remote → Local)

Files are synced from the cloud to your local folder. The local folder becomes an exact copy of the remote.

**Use case**: Downloading files from cloud to local machine

```bash
# Example: Download photos from cloud
Local:  /Users/username/Pictures
Remote: s3:photo-bucket/pictures
Direction: download
```

### Bidirectional (Both Ways)

Files are synced in both directions. Changes in either location are propagated to the other.

**Use case**: Keeping folders in sync across multiple machines

**Note**: Bidirectional sync performs upload first, then download. For production use, consider using rclone's bisync feature for proper conflict resolution.

```bash
# Example: Keep work folder in sync
Local:  /Users/username/Work
Remote: gdrive:work-backup
Direction: bidirectional
```

## Best Practices

### 1. Start with Dry-Run

Always test with `--dry-run` first to preview changes:

```bash
~/bin/sync_local_folder.sh documents --dry-run
```

### 2. Use Descriptive Names

Choose clear, descriptive names for sync pairs:

- ✅ `documents-backup`
- ✅ `photos-archive`
- ❌ `sync1`
- ❌ `test`

### 3. Validate Paths

Ensure local paths exist before creating sync pairs:

```bash
ls -la /Users/username/Documents
```

### 4. Configure Remotes First

Set up rclone remotes before creating sync pairs:

```bash
rclone config
```

### 5. Monitor Logs

Check sync logs for errors:

```bash
tail -f ~/logs/sync_documents_*.log
```

### 6. Be Careful with Bidirectional

Bidirectional sync can lead to data loss if not used carefully. Consider:

- Using version control for important files
- Regular backups before syncing
- Testing with non-critical data first

### 7. Exclude Sensitive Files

Use rclone filters to exclude sensitive or temporary files:

```bash
# Add to rclone config or use --exclude flags
--exclude ".DS_Store"
--exclude "*.tmp"
--exclude ".git/"
```

## Automation

### Schedule with LaunchAgent (macOS)

Create a LaunchAgent to run sync automatically:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.cloud-sync-folders</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/username/bin/sync_all_folders.sh</string>
    </array>
    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>2</integer>
        <key>Minute</key>
        <integer>0</integer>
    </dict>
    <key>StandardOutPath</key>
    <string>/Users/username/logs/sync-folders.log</string>
    <key>StandardErrorPath</key>
    <string>/Users/username/logs/sync-folders-error.log</string>
</dict>
</plist>
```

Load the agent:

```bash
launchctl load ~/Library/LaunchAgents/com.user.cloud-sync-folders.plist
```

## Troubleshooting

### Sync Pair Not Found

**Error**: `sync pair 'name' not found`

**Solution**: Check the sync pair name and list all pairs:

```bash
# Using Go API
pairs, _ := manager.ListSyncPairs()

# Or check config file
cat ~/.config/cloud-sync/sync-config.json
```

### Local Path Does Not Exist

**Error**: `local path does not exist: /path/to/folder`

**Solution**: Create the directory or fix the path:

```bash
mkdir -p /path/to/folder
```

### Remote Not Configured

**Error**: `remote test failed`

**Solution**: Configure the remote in rclone:

```bash
rclone config
```

### Permission Denied

**Error**: `failed to read directory: permission denied`

**Solution**: Check folder permissions:

```bash
ls -la /path/to/folder
chmod 755 /path/to/folder
```

### Sync Conflicts (Bidirectional)

**Issue**: Files being overwritten in bidirectional sync

**Solution**: 
- Use upload or download direction instead
- Consider using rclone's bisync feature
- Implement proper conflict resolution

## Examples

### Example 1: Backup Documents to Backblaze

```go
manager.AddSyncPair(
    "documents-backup",
    "/Users/john/Documents",
    "backblaze",
    "my-bucket/documents",
    "upload",
)
```

### Example 2: Download Photos from S3

```go
manager.AddSyncPair(
    "photo-download",
    "/Users/john/Pictures/Archive",
    "s3",
    "photo-bucket/archive",
    "download",
)
```

### Example 3: Bidirectional Sync with Google Drive

```go
manager.AddSyncPair(
    "work-sync",
    "/Users/john/Work",
    "gdrive",
    "work-backup",
    "bidirectional",
)
```

## Advanced Usage

### Custom Sync Scripts

You can create custom sync scripts based on the templates:

```bash
#!/bin/zsh
# Custom sync script with additional logic

# Pre-sync hook
echo "Running pre-sync checks..."
# ... custom logic ...

# Run sync
~/bin/sync_local_folder.sh documents

# Post-sync hook
echo "Running post-sync cleanup..."
# ... custom logic ...
```

### Integration with Other Tools

Integrate with monitoring tools:

```bash
#!/bin/zsh
# Sync with monitoring

# Send start notification
curl -X POST https://monitoring.example.com/start

# Run sync
~/bin/sync_all_folders.sh

# Send completion notification
if [ $? -eq 0 ]; then
    curl -X POST https://monitoring.example.com/success
else
    curl -X POST https://monitoring.example.com/failure
fi
```

## See Also

- [Rclone Documentation](https://rclone.org/docs/)
- [Cloud-Sync README](../README.md)
- [Project Plan](../project-plan.md)
