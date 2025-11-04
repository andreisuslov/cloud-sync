# Cloud Sync - Backup Management Tool

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)

 A beautiful terminal-based GUI application built with [Bubbletea](https://github.com/charmbracelet/bubbletea) for managing automated cloud-to-cloud backups between any rclone-supported cloud storage providers.

## ✨ Features

### 🚀 Installation & Setup
- **One-click tool installation**: Automatically installs Homebrew and rclone
- **Guided configuration**: Interactive wizard for configuring cloud remotes
- **Bucket discovery**: List and select buckets from configured remotes
- **Script generation**: Automatically creates all required backup scripts
- **LaunchAgent setup**: Configures macOS automation without manual plist editing

### 💾 Backup Operations
- **Manual sync**: Run on-demand backups with live progress
- **Automated monthly backups**: Set-and-forget monthly cloud sync
- **Duplicate prevention**: Smart monthly check prevents redundant syncs
- **Lockfile protection**: Prevents simultaneous backup operations
- **Progress tracking**: Real-time transfer speed and file count

### 📊 Monitoring & Logs
- **Log viewer**: Browse all backup logs with syntax highlighting
- **Transfer history**: View today's transfers or recent N transfers
- **Sync statistics**: Success rate, total files, last sync time
- **Real-time tailing**: Watch logs as backups run
- **Search & filter**: Find specific log entries quickly

### ⚙️ LaunchAgent Management
- **Status monitoring**: Check if agent is loaded and running
- **Load/Unload**: Enable or disable automation
- **Manual triggers**: Start automated jobs on-demand
- **Schedule viewing**: See next scheduled run time
- **Dynamic configuration**: Automatically uses current user's login session

### 🛠️ Maintenance Tools
- **Lockfile management**: Remove stale lockfiles safely
- **Timestamp reset**: Force monthly backup to run again
- **Log cleanup**: Clear old logs to save space
- **Configuration review**: View current settings

## 🏗️ Architecture

Implements the proven **three-script design** from backup.md:

1. **Engine Script** (`run_rclone_sync.sh`): Core rclone execution
2. **Automated Script** (`monthly_backup.sh`): Monthly check with deduplication
3. **Manual Script** (`sync_now.sh`): On-demand sync with progress display

Plus additional helper scripts:
- `show_transfers.sh`: View transfer history
- LaunchAgent plist: macOS automation configuration

## 📋 Prerequisites

- **macOS**: 10.15 (Catalina) or later
- **Architecture**: Intel (x86_64) or Apple Silicon (arm64)
- **Cloud Storage**:
  - Two cloud storage accounts (source and destination)
  - Supported providers: AWS S3, Backblaze B2, Google Cloud Storage, Scaleway, DigitalOcean Spaces, Azure Blob Storage, Wasabi, and [many more](https://rclone.org/#providers)
  - API credentials/access keys for your chosen providers
- **Network**: Internet connection for cloud sync

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/cloud-sync.git
cd cloud-sync

# Build the application
go build -o cloud-sync ./cmd/cloud-sync

# Run the application
./cloud-sync
```

### First Time Setup

1. **Launch the app**: Run `./cloud-sync`
2. **Installation Wizard**: Follow prompts to install Homebrew and rclone
3. **Configure Remotes**:
   - Configure source remote using `rclone config` (any supported provider)
   - Configure destination remote using `rclone config` (any supported provider)
4. **Select Buckets**: Choose source and destination buckets/containers
5. **Generate Scripts**: App creates all scripts in `~/bin`
6. **Setup Automation**: LaunchAgent configured for monthly backups
7. **Done!** Your backups are now automated

### Daily Usage

```bash
# Run the TUI application
./cloud-sync

# Navigate with arrow keys
# Press Enter to select
# Press 'q' to quit
```

## 📖 Documentation

- **[Project Plan](project-plan.md)**: Detailed implementation roadmap
- **[Original Backup Guide](https://github.com/yourusername/cloud-sync/docs/backup.md)**: Manual setup instructions
- **[Architecture](docs/architecture.md)**: System design and modules *(coming soon)*
- **[API Reference](docs/api.md)**: Developer documentation *(coming soon)*

## 🧪 Development

### Project Structure

```
cloud-sync/
├── cmd/
│   └── cloud-sync/          # Main application entry point
├── internal/
│   ├── installer/           # Homebrew and rclone installation
│   ├── rclone/              # Rclone configuration and operations
│   ├── scripts/             # Script generation and management
│   ├── launchd/             # LaunchAgent management
│   ├── logs/                # Log parsing and viewing
│   ├── lockfile/            # Lockfile management
│   └── ui/                  # Bubbletea TUI components
├── pkg/
│   └── backup/              # Public API for backup operations
├── tests/
│   ├── unit/               # Unit tests
│   └── integration/        # Integration tests
├── scripts/                # Script templates
├── go.mod
├── go.sum
├── README.md               # This file
└── project-plan.md         # Detailed project plan
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...

# Run integration tests
go test -tags=integration ./tests/integration/...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Building

```bash
# Build for current platform
go build -o cloud-sync ./cmd/cloud-sync

# Build with version info
go build -ldflags "-X main.Version=1.0.0" -o cloud-sync ./cmd/cloud-sync

# Build for Intel Mac
GOARCH=amd64 go build -o cloud-sync-amd64 ./cmd/cloud-sync

# Build for Apple Silicon
GOARCH=arm64 go build -o cloud-sync-arm64 ./cmd/cloud-sync
```

## 🤝 Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Write tests** for your changes
4. **Ensure** all tests pass (`go test ./...`)
5. **Commit** your changes (`git commit -m 'feat: add amazing feature'`)
6. **Push** to the branch (`git push origin feature/amazing-feature`)
7. **Open** a Pull Request

### Commit Message Convention

We follow conventional commits:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation changes
- `test:` Test additions or updates
- `refactor:` Code refactoring
- `style:` Code style changes
- `chore:` Maintenance tasks

## 🐛 Troubleshooting

### Backup doesn't run automatically

```bash
# Check if LaunchAgent is loaded
launchctl list | grep rclone

# Reload the LaunchAgent
launchctl unload ~/Library/LaunchAgents/com.$(whoami).rclonebackup.plist
launchctl load ~/Library/LaunchAgents/com.$(whoami).rclonebackup.plist
```

### Lockfile error appears

```bash
# Remove stale lockfile (or use app's Maintenance menu)
rm $HOME/logs/rclone_backup.lock
```

### Force monthly backup to run again

```bash
# Reset timestamp (or use app's Maintenance menu)
echo "2025-10" > $HOME/logs/rclone_last_run_timestamp
```

### Rclone not found

```bash
# Verify installation
which rclone

# Reinstall if needed
brew install rclone
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Charm Bubbletea](https://github.com/charmbracelet/bubbletea) - Fantastic TUI framework
- [rclone](https://rclone.org/) - Powerful cloud sync tool supporting 70+ cloud storage providers

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/yourusername/cloud-sync/issues)
- **Discussions**: [GitHub Discussions](https://github.com/yourusername/cloud-sync/discussions)
- **Documentation**: See `docs/` folder

## 🗺️ Roadmap

See [project-plan.md](project-plan.md) for the complete implementation roadmap.

### Upcoming Features
- ✅ Core backup functionality
- ✅ LaunchAgent automation
- ✅ Log viewer
- 🚧 Multi-profile support
- 🚧 Email notifications
- 📋 Backup verification
- 📋 Restore functionality
- 📋 Additional cloud providers

---

**Status**: 🚧 Under active development - see [project-plan.md](project-plan.md) for current progress

**Version**: 0.1.0-dev

**Last Updated**: November 2025