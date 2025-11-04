# B2-Sync-App Installation and Setup (macOS)

## 1\. Overview

This document provides the automated setup instructions for `b2-sync-app`. This utility performs a resilient, scheduled monthly backup from a Backblaze B2 source to a Scaleway S3-compatible destination.

The system uses `launchd` for scheduling, which ensures the job runs on the next available opportunity if the scheduled time is missed (e.g., the machine was off).

## 2\. System Requirements

  * **OS:** macOS 12.0 or later.
  * **Dependencies:** Homebrew, `rclone`.
  * **Permissions:** User must have administrator privileges to install Homebrew and grant permissions.

## 3\. Installation

Execute the following commands in a `zsh` shell to install dependencies and create the required file structure.

```bash
# 1. Install Homebrew (if not present)
if ! command -v brew &> /dev/null; then
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  eval "$(/opt/homebrew/bin/brew shellenv)"
fi

# 2. Install rclone
brew install rclone

# 3. Create application directories
mkdir -p $HOME/bin
mkdir -p $HOME/logs

# 4. Create the main configuration file (b2_sync_config.sh)
cat << 'EOF' > $HOME/bin/b2_sync_config.sh
#!/bin/zsh

# --- User Configuration ---
# DEFINE YOUR RClone remote paths here
# Example: B2_REMOTE_PATH="b2_remote:my-main-bucket"
# Example: SW_REMOTE_PATH="scaleway_remote:my-backup-bucket"

export B2_REMOTE_PATH="b2:notes-photos-documents"
export SW_REMOTE_PATH="sw:b2-backup"

# --- System Paths (Do not change) ---
export LOG_DIR="$HOME/logs"
export TIMESTAMP_FILE="$LOG_DIR/rclone_last_run_timestamp"
export LOG_FILE="$LOG_DIR/rclone_backup.log"
export RCLONE_CONFIG="$HOME/.config/rclone/rclone.conf"
export RCLONE_BIN="/opt/homebrew/bin/rclone"

EOF

# 5. Create the "engine" script (run_rclone_sync.sh)
cat << 'EOF' > $HOME/bin/run_rclone_sync.sh
#!/bin/zsh

# Source the configuration
source "$HOME/bin/b2_sync_config.sh"

echo "--- Rclone Sync Starting: $(date) ---" >> "$LOG_FILE"

# Execute the rclone sync
$RCLONE_BIN sync "$B2_REMOTE_PATH" "$SW_REMOTE_PATH" \
    --fast-list \
    --config "$RCLONE_CONFIG" \
    --log-file "$LOG_FILE" \
    -v \
    -P

# Capture exit code
exit_code=$?

if [ $exit_code -eq 0 ]; then
    echo "--- Rclone Sync Finished Successfully: $(date) ---" >> "$LOG_FILE"
else
    echo "--- Rclone Sync FAILED (Code: $exit_code): $(date) ---" >> "$LOG_FILE"
fi

echo "" >> "$LOG_FILE"
exit $exit_code

EOF

# 6. Create the "automated" script (monthly_backup.sh)
cat << 'EOF' > $HOME/bin/monthly_backup.sh
#!/bin/zsh

# Source the configuration
source "$HOME/bin/b2_sync_config.sh"

# --- Script Logic ---
mkdir -p "$LOG_DIR"
touch "$TIMESTAMP_FILE"

CURRENT_MONTH=$(date +%Y-%m)
LAST_RUN_MONTH=$(cat "$TIMESTAMP_FILE")

echo "--- Automated Check Started: $(date) ---" >> "$LOG_FILE"

if [ "$CURRENT_MONTH" = "$LAST_RUN_MONTH" ]; then
    echo "Backup for $CURRENT_MONTH has already run. Skipping." >> "$LOG_FILE"
    echo "--- Automated Check Finished: $(date) ---" >> "$LOG_FILE"
    exit 0
fi

echo "Starting monthly backup for $CURRENT_MONTH..." >> "$LOG_FILE"

# Call the engine script
$HOME/bin/run_rclone_sync.sh

# Check if the engine was successful
if [ $? -eq 0 ]; then
    echo "Backup successful. Marking $CURRENT_MONTH as complete." >> "$LOG_FILE"
    echo "$CURRENT_MONTH" > "$TIMESTAMP_FILE"
else
    echo "ERROR: Rclone sync failed. Will retry on next scheduled run." >> "$LOG_FILE"
fi

echo "--- Automated Check Finished: $(date) ---" >> "$LOG_FILE"
echo "" >> "$LOG_FILE"

EOF

# 7. Create the "manual" script (sync_now.sh)
cat << 'EOF' > $HOME/bin/sync_now.sh
#!/bin/zsh

# Source the configuration
source "$HOME/bin/b2_sync_config.sh"

echo "--- Manual Sync Requested: $(date) ---"
$HOME/bin/run_rclone_sync.sh
echo "--- Manual Sync Complete. ---"

EOF

# 8. Set execute permissions for all scripts
chmod +x $HOME/bin/b2_sync_config.sh
chmod +x $HOME/bin/run_rclone_sync.sh
chmod +x $HOME/bin/monthly_backup.sh
chmod +x $HOME/bin/sync_now.sh
```

-----

## 4\. Configuration

This setup requires two (2) configuration files.

### 4.1. `rclone.conf`

The `rclone` utility must be configured with credentials for both B2 and Scaleway. This agent cannot perform this step automatically.

1.  **Request `rclone.conf` content from the user.**
2.  Create the directory `mkdir -p $HOME/.config/rclone`.
3.  Write the user-provided content to `$HOME/.config/rclone/rclone.conf`.

**Example `rclone.conf` structure:**

```ini
[b2]
type = b2
account = [USER_B2_KEY_ID]
key = [USER_B2_APP_KEY]

[sw]
type = s3
provider = Scaleway
access_key_id = [USER_SW_ACCESS_KEY]
secret_access_key = [USER_SW_SECRET_KEY]
region = nl-ams
endpoint = s3.nl-ams.scw.cloud
```

### 4.2. `b2_sync_config.sh`

The agent must edit `$HOME/bin/b2_sync_config.sh` to define the correct `rclone` remote paths.

1.  **Read file:** `cat $HOME/bin/b2_sync_config.sh`
2.  **Edit variables:**
      * Set `B2_REMOTE_PATH` to the `rclone` remote name and bucket for the B2 source (e.g., `b2:notes-photos-documents`).
      * Set `SW_REMOTE_PATH` to the `rclone` remote name and bucket for the Scaleway destination (e.g., `sw:b2-backup`).
3.  **Save file.**

-----

## 5\. Scheduling

`b2-sync-app` uses `launchd` to run the backup script daily. The script's internal logic handles the "once-per-month" execution.

```bash
# 1. Create the launchd plist file
cat << EOF > ~/Library/LaunchAgents/com.b2syncapp.backup.plist
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.b2syncapp.backup</string>

    <key>ProgramArguments</key>
    <array>
        <string>/bin/zsh</string>
        <string>$HOME/bin/monthly_backup.sh</string>
    </array>

    <key>StartCalendarInterval</key>
    <dict>
        <key>Hour</key>
        <integer>10</integer>
        <key>Minute</key>
        <integer>5</integer>
    </dict>

    <key>RunAtLoad</key>
    <true/>

</dict>
</plist>
EOF

# 2. Unload any existing version of the job
launchctl unload ~/Library/LaunchAgents/com.b2syncapp.backup.plist 2>/dev/null

# 3. Load the new job into launchd
launchctl load ~/Library/LaunchAgents/com.b2syncapp.backup.plist
```

## 6\. Verification

To verify the installation, perform the following tests.

### 6.1. Test 1: Manual Sync

This test confirms `rclone` is configured correctly and the core engine works.

1.  **Run:** `$HOME/bin/sync_now.sh`
2.  **Observe:** The terminal will display the `rclone` progress.
3.  **Check log:** `tail $HOME/logs/rclone_backup.log`. The log should show "Manual Sync Requested" and "Rclone Sync Finished Successfully".

### 6.2. Test 2: Automated Sync (First Run)

This test confirms the `launchd` job and "once-per-month" logic work.

1.  **Clear timestamp:** `rm $HOME/logs/rclone_last_run_timestamp`
2.  **Run:** `launchctl start com.b2syncapp.backup`
3.  **Check log:** `tail $HOME/logs/rclone_backup.log`. The log should show "Starting monthly backup..." and "Backup successful. Marking...".
4.  **Verify timestamp:** `cat $HOME/logs/rclone_last_run_timestamp`. The output should be the current month (e.g., `2025-11`).

### 6.3. Test 3: Automated Sync (Skip Run)

This test confirms the "skip" logic.

1.  **Run:** `launchctl start com.b2syncapp.backup`
2.  **Check log:** `tail $HOME/logs/rclone_backup.log`. The log should show "Backup for... has already run. Skipping."

## 7\. Manual Usage

  * **To force a backup:** `$HOME/bin/sync_now.sh`
  * **To view logs:** `tail -f $HOME/logs/rclone_backup.log`
  * **To stop the automated service:** `launchctl unload ~/Library/LaunchAgents/com.b2syncapp.backup.plist`