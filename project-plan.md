# Cloud Backup Management Tool - Project Plan

## Project Overview
A terminal-based GUI application (using bubbletea) for managing automated cloud-to-cloud backups from Backblaze B2 to Scaleway Object Storage. The tool implements the 3-2-1 backup strategy with comprehensive management capabilities.

## Technology Stack
- **Language**: Go 1.21+
- **TUI Framework**: Charm Bubbletea
- **UI Components**: Charm Bubbles (list, viewport, spinner, progress)
- **Cloud Sync**: rclone
- **Automation**: macOS launchd
- **Version Control**: Git
- **Testing**: Go testing package + testify

---

## Architecture

### Three-Script Design (from backup.md)
1. **Engine Script** (`run_rclone_sync.sh`): Core rclone execution
2. **Automated Script** (`monthly_backup.sh`): Monthly check with deduplication
3. **Manual Script** (`sync_now.sh`): On-demand sync with progress

### Go Application Modules
```
cloud-sync/
├── cmd/
│   └── cloud-sync/          # Main application entry point
│       └── main.go
├── internal/
│   ├── installer/           # Install homebrew, rclone
│   ├── rclone/              # Rclone configuration and operations
│   ├── scripts/             # Script generation and management
│   ├── launchd/             # LaunchAgent management
│   ├── logs/                # Log parsing and viewing
│   ├── lockfile/            # Lockfile management
│   └── ui/                  # Bubbletea TUI components
│       ├── models/          # UI state models
│       ├── views/           # Screen views
│       └── styles/          # Styling constants
├── pkg/
│   └── backup/              # Public API for backup operations
├── scripts/                 # Embedded script templates
├── tests/
│   ├── unit/               # Unit tests
│   └── integration/        # Integration tests
├── go.mod
├── go.sum
├── README.md
├── project-plan.md         # This file
└── .gitignore
```

---

## Feature Implementation Plan

### Phase 1: Core Infrastructure ✓
- [x] Project structure setup
- [x] Git initialization
- [x] Go module initialization
- [x] Dependency management (go.mod)

### Phase 2: Installation Module ✓
**Module**: `internal/installer`

#### Features:
- [x] Check if Homebrew is installed
- [x] Install Homebrew with user confirmation
- [x] Check if rclone is installed
- [x] Install rclone via Homebrew
- [x] Verify installation success
- [x] Handle PATH issues for Apple Silicon vs Intel

#### Functions:
```go
- CheckHomebrewInstalled() bool
- InstallHomebrew() error
- CheckRcloneInstalled() bool
- InstallRclone() error
- GetRclonePath() string
```

#### Tests:
- Unit tests for each check function
- Mock command execution for install tests
- Integration tests on clean system (manual)

---

### Phase 3: Rclone Configuration Module ✓
**Module**: `internal/rclone`

#### Features:
- [x] Interactive remote configuration (B2, Scaleway, any other rsync supported platform)
- [x] List configured remotes
- [x] Test remote connectivity
- [x] List buckets for each remote
- [x] Validate bucket access
- [x] Store remote configuration
- [x] Parse rclone.conf file

#### Functions:
```go
- ConfigureRemote(remoteType, name string) error
- ListRemotes() ([]Remote, error)
- ListBuckets(remoteName string) ([]Bucket, error)
- TestRemote(remoteName string) error
- GetConfigPath() string
- ParseConfig() (*Config, error)
```

#### Structs:
```go
type Remote struct {
    Name     string
    Type     string
    Provider string
}

type Bucket struct {
    Name         string
    Size         int64
    ModifiedTime time.Time
}
```

#### Tests:
- Unit tests with mock config files
- Test config parsing
- Test bucket listing with mock responses

---

### Phase 4: Script Generation Module ✓
**Module**: `internal/scripts`

#### Features:
- [x] Generate `run_rclone_sync.sh` from template
- [x] Generate `monthly_backup.sh` from template
- [x] Generate `sync_now.sh` from template
- [x] Generate `show_transfers.sh` helper script
- [x] Customize scripts with user-specific paths
- [x] Customize scripts with bucket names
- [x] Automatically make scripts executable
- [x] Create required directories ($HOME/bin, $HOME/logs)

#### Functions:
```go
- GenerateEngineScript(config *ScriptConfig) error
- GenerateMonthlyScript(config *ScriptConfig) error
- GenerateManualScript(config *ScriptConfig) error
- GenerateShowTransfersScript(config *ScriptConfig) error
- MakeExecutable(filepath string) error
- CreateDirectories() error
```

#### Structs:
```go
type ScriptConfig struct {
    HomeDir          string
    Username         string
    RclonePath       string
    SourceRemote     string
    SourceBucket     string
    DestRemote       string
    DestBucket       string
    LogDir           string
    BinDir           string
}
```

#### Tests:
- Test template rendering
- Test script generation with various configs
- Verify executable permissions
- Test directory creation

---

### Phase 5: LaunchAgent Management Module ✓
**Module**: `internal/launchd`

#### Features:
- [x] Generate plist file for LaunchAgent
- [x] Dynamically determine user login name
- [x] Load LaunchAgent
- [x] Unload LaunchAgent
- [x] Check LaunchAgent status
- [x] Start LaunchAgent manually
- [x] Stop LaunchAgent
- [x] View LaunchAgent logs
- [x] Handle com.{username}.rclonebackup naming

#### Functions:
```go
- GeneratePlist(config *LaunchdConfig) error
- GetLaunchAgentPath(username string) string
- LoadAgent() error
- UnloadAgent() error
- IsLoaded() (bool, error)
- GetStatus() (*AgentStatus, error)
- StartAgent() error
- StopAgent() error
```

#### Structs:
```go
type LaunchdConfig struct {
    Label        string  // com.{username}.rclonebackup
    ScriptPath   string
    Hour         int
    Minute       int
    RunAtLoad    bool
}

type AgentStatus struct {
    Loaded   bool
    Running  bool
    LastExit int
    PID      int
}
```

#### Tests:
- Test plist generation
- Test status parsing
- Mock launchctl commands
- Test label generation

---

### Phase 6: Log Management Module ✓
**Module**: `internal/logs`

#### Features:
- [x] Parse rclone log file
- [x] Extract transfer information
- [x] Filter by date (today's transfers)
- [x] Show most recent N transfers
- [x] Parse automated check entries
- [x] Show sync statistics
- [x] Tail logs in real-time
- [x] Clear old logs

#### Functions:
```go
- ParseLogFile(filepath string) (*LogData, error)
- GetTodaysTransfers() ([]Transfer, error)
- GetRecentTransfers(count int) ([]Transfer, error)
- GetSyncStats() (*SyncStats, error)
- TailLog(lines int) ([]string, error)
- WatchLog(callback func(line string)) error
```

#### Structs:
```go
type LogData struct {
    Transfers []Transfer
    Syncs     []SyncSession
}

type Transfer struct {
    Timestamp time.Time
    Filename  string
    Size      int64
    Action    string  // Copied, Deleted, etc.
}

type SyncSession struct {
    StartTime time.Time
    EndTime   time.Time
    Success   bool
    Type      string  // Manual, Automated
    Transfers int
}

type SyncStats struct {
    TotalFiles     int
    TotalSize      int64
    LastSync       time.Time
    LastSuccess    time.Time
    SuccessRate    float64
}
```

#### Tests:
- Test log parsing with sample logs
- Test date filtering
- Test transfer extraction
- Test statistics calculation

---

### Phase 7: Lockfile Management Module ✓
**Module**: `internal/lockfile`

#### Features:
- [x] Check if lockfile exists
- [x] Create lockfile
- [x] Remove lockfile
- [x] Get lockfile age
- [x] Detect stale lockfiles
- [x] Force remove with confirmation

#### Functions:
```go
- Exists() bool
- Create() error
- Remove() error
- GetAge() (time.Duration, error)
- IsStale(maxAge time.Duration) bool
- ForceRemove() error
```

#### Tests:
- Test lockfile creation/removal
- Test age calculation
- Test stale detection

---

### Phase 8: Bubbletea UI - Main Menu ⚠️ (In Progress)
**Module**: `internal/ui`

#### Features:
- [ ] Welcome screen
- [x] Main menu with navigation (basic)
- [ ] Installation wizard
- [ ] Configuration wizard
- [ ] Status dashboard
- [x] Keyboard shortcuts help (basic)

#### Menu Items:
1. **Installation & Setup**
   - Install required tools
   - Configure rclone remotes
   - Identify buckets
   - Generate scripts
   - Setup LaunchAgent

2. **Backup Operations**
   - Run manual sync
   - Trigger automated job
   - View sync progress

3. **Log Viewer**
   - View all logs
   - Today's transfers
   - Recent transfers
   - Sync statistics

4. **LaunchAgent Management**
   - Check status
   - Load agent
   - Unload agent
   - Start manually
   - Stop agent

5. **Maintenance**
   - Remove lockfile
   - Reset monthly timestamp
   - Clear old logs

6. **Exit**

#### UI Components:
```go
type Model struct {
    state        AppState
    menu         list.Model
    viewport     viewport.Model
    spinner      spinner.Model
    progress     progress.Model
    width        int
    height       int
    err          error
}

type AppState int
const (
    StateMainMenu AppState = iota
    StateInstallation
    StateConfiguration
    StateBackupRunning
    StateLogViewer
    StateLaunchdManager
    StateMaintenance
)
```

#### Tests:
- Test navigation logic
- Test state transitions
- Test rendering (snapshots)

---

### Phase 9: UI Views - Installation Wizard
**Module**: `internal/ui/views`

#### Features:
- [ ] Step-by-step installation guide
- [ ] Progress indicators
- [ ] Success/failure feedback
- [ ] Automatic progression

#### Steps:
1. Check Homebrew → Install if needed
2. Check rclone → Install if needed
3. Success screen

#### Tests:
- Test each installation step
- Test error handling
- Test skip logic if already installed

---

### Phase 10: UI Views - Configuration Wizard
**Module**: `internal/ui/views`

#### Features:
- [ ] Remote configuration forms
- [ ] Input validation
- [ ] Bucket selection
- [ ] Configuration preview
- [ ] Test connectivity button

#### Steps:
1. Configure B2 remote (keyID, applicationKey)
2. List and select B2 bucket
3. Configure Scaleway remote (access key, secret key, region)
4. List and select Scaleway bucket
5. Review and confirm
6. Generate scripts
7. Setup LaunchAgent
8. Success screen

#### Form Components:
```go
type FormModel struct {
    inputs   []textinput.Model
    focused  int
    err      error
}
```

#### Tests:
- Test form validation
- Test input handling
- Test configuration save

---

### Phase 11: UI Views - Backup Operations
**Module**: `internal/ui/views`

#### Features:
- [ ] Live sync progress display
- [ ] Transfer speed indicator
- [ ] File count and size
- [ ] Cancel operation
- [ ] Post-sync summary

#### Display:
- Progress bar
- Current file being transferred
- Transfer rate (MB/s)
- Elapsed time
- ETA
- Total files/size

#### Tests:
- Test progress parsing
- Test real-time updates
- Test cancellation

---

### Phase 12: UI Views - Log Viewer
**Module**: `internal/ui/views`

#### Features:
- [ ] Scrollable log display
- [ ] Syntax highlighting for errors
- [ ] Filter options (date, type)
- [ ] Search functionality
- [ ] Export logs

#### View Options:
- All logs
- Today's transfers
- Recent N transfers
- Errors only
- Sync sessions only

#### Tests:
- Test log rendering
- Test filtering
- Test search

---

### Phase 13: UI Views - LaunchAgent Manager
**Module**: `internal/ui/views`

#### Features:
- [ ] Status display
- [ ] Load/Unload buttons
- [ ] Start/Stop buttons
- [ ] View agent logs
- [ ] Schedule configuration

#### Status Display:
- Loaded: Yes/No
- Running: Yes/No
- Last Run: timestamp
- Next Run: timestamp
- PID: process ID

#### Tests:
- Test status rendering
- Test action handlers

---

### Phase 14: Testing Infrastructure

#### Unit Tests
Location: `tests/unit/`

**Coverage Goals:**
- Installer: 90%+
- Rclone: 85%+
- Scripts: 90%+
- Launchd: 85%+
- Logs: 90%+
- Lockfile: 95%+

**Test Files:**
```
tests/unit/
├── installer_test.go
├── rclone_test.go
├── scripts_test.go
├── launchd_test.go
├── logs_test.go
└── lockfile_test.go
```

**Mocking Strategy:**
- Use testify/mock for external commands
- Mock filesystem operations
- Mock network calls

#### Integration Tests
Location: `tests/integration/`

**Test Scenarios:**
1. Full installation flow (clean system)
2. Configuration flow with real rclone
3. Script generation and execution
4. LaunchAgent lifecycle
5. End-to-end backup operation

**Test Files:**
```
tests/integration/
├── install_flow_test.go
├── config_flow_test.go
├── backup_flow_test.go
└── launchd_flow_test.go
```

**Requirements:**
- Docker container for isolated testing
- Test fixtures for config files
- Mock rclone remotes

#### Test Utilities
```go
// tests/testutil/
- MockCommandExecutor
- MockFilesystem
- TestDataGenerator
- AssertionHelpers
```

---

### Phase 15: Error Handling & Logging

#### Error Types
```go
type ErrorCode int
const (
    ErrInstallation ErrorCode = iota
    ErrConfiguration
    ErrScriptGeneration
    ErrLaunchdOperation
    ErrBackupFailed
    ErrLockfileExists
    ErrInvalidInput
)

type AppError struct {
    Code    ErrorCode
    Message string
    Cause   error
}
```

#### Logging Strategy
- Use structured logging (log/slog in Go 1.21+)
- Log levels: Debug, Info, Warn, Error
- Separate app log from rclone log
- Log rotation for app logs

#### Log Files:
```
$HOME/logs/
├── rclone_backup.log         # Rclone output
├── rclone_last_run_timestamp # Monthly tracking
├── rclone_backup.lock        # Lockfile
└── cloud-sync.log            # App log
```

---

### Phase 16: Documentation

#### README.md
- [x] Project description
- [x] Features overview
- [x] Installation instructions
- [x] Quick start guide
- [ ] Screenshots/GIFs
- [x] Troubleshooting
- [x] Contributing guidelines

#### User Guide
- [ ] Detailed feature documentation
- [ ] Configuration examples
- [ ] Common workflows
- [ ] FAQ

#### Developer Documentation
- [ ] Architecture overview
- [ ] Module documentation
- [ ] API reference
- [ ] Testing guide
- [ ] Build instructions

---

## Implementation Order

### Sprint 1: Foundation (Days 1-2)
1. ✓ Project setup and Git init
2. ✓ Go module initialization
3. ✓ Project plan documentation
4. Core directory structure
5. Dependency installation

### Sprint 2: Core Modules (Days 3-5)
1. Installer module + tests
2. Rclone module + tests
3. Scripts module + tests
4. Lockfile module + tests

### Sprint 3: Advanced Modules (Days 6-8)
1. Launchd module + tests
2. Logs module + tests
3. Integration between modules

### Sprint 4: Basic UI (Days 9-11)
1. Main menu structure
2. Navigation logic
3. Installation wizard
4. Configuration wizard

### Sprint 5: Advanced UI (Days 12-14)
1. Backup operations view
2. Log viewer
3. LaunchAgent manager
4. Maintenance tools

### Sprint 6: Testing & Polish (Days 15-17)
1. Comprehensive unit tests
2. Integration tests
3. Error handling improvements
4. UI polish and refinements

### Sprint 7: Documentation & Release (Days 18-20)
1. README and user guide
2. Developer documentation
3. Final testing
4. Release preparation

---

## Git Workflow

### Branch Strategy
- `main`: Production-ready code
- `develop`: Integration branch
- `feature/*`: Feature branches
- `bugfix/*`: Bug fix branches
- `test/*`: Testing improvements

### Commit Message Convention
```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- feat: New feature
- fix: Bug fix
- docs: Documentation
- test: Testing
- refactor: Code refactoring
- style: Formatting
- chore: Maintenance

**Example:**
```
feat(installer): add homebrew installation check

Implement CheckHomebrewInstalled() function that verifies
if Homebrew is installed by checking for the brew command.

Closes #12
```

### Commit Frequency
- Commit after each logical unit of work
- Commit after adding tests
- Commit before major refactoring
- Minimum: One commit per module/feature

---

## Testing Strategy

### Test-Driven Development (TDD)
1. Write test first (red)
2. Implement minimum code to pass (green)
3. Refactor (refactor)
4. Repeat

### Test Coverage Goals
- Overall: 80%+
- Critical paths: 95%+
- UI logic: 70%+
- Integration: Key workflows covered

### Continuous Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detector
go test -race ./...

# Run integration tests only
go test -tags=integration ./tests/integration/...
```

---

## Security Considerations

### Credentials Management
- Never log credentials
- Store in rclone.conf with appropriate permissions (0600)
- Support for environment variables
- Clear credential prompts in UI

### File Permissions
- Scripts: 0755 (executable by owner)
- Config files: 0600 (read/write by owner only)
- Log files: 0644 (readable by all, writable by owner)

### Input Validation
- Sanitize all user inputs
- Validate remote names
- Validate bucket names
- Prevent path traversal

---

## Performance Considerations

### Async Operations
- Run rclone in background goroutine
- Non-blocking UI updates
- Progress streaming

### Resource Management
- Limit concurrent operations
- Clean up goroutines
- Close file handles
- Release locks properly

### Large Log Files
- Implement log rotation
- Use streaming for large file reads
- Pagination for log viewer
- Indexing for quick searches

---

## Deployment

### Build Process
```bash
# Build for macOS
go build -o cloud-sync ./cmd/cloud-sync

# Build with version info
go build -ldflags "-X main.Version=1.0.0" -o cloud-sync ./cmd/cloud-sync

# Cross-compile for Intel Mac
GOARCH=amd64 go build -o cloud-sync-amd64 ./cmd/cloud-sync

# Cross-compile for Apple Silicon
GOARCH=arm64 go build -o cloud-sync-arm64 ./cmd/cloud-sync
```

### Installation Methods
1. **Direct download**: Binary with installer script
2. **Homebrew**: Custom tap (future)
3. **Go install**: `go install github.com/user/cloud-sync@latest`

### Distribution
- GitHub Releases with binaries
- Installation script for auto-detection
- Checksums for verification

---

## Success Criteria

### Functionality
- ✅ All features from backup.md implemented
- ✅ Fully functional TUI with intuitive navigation
- ✅ Successful backup operations (manual and automated)
- ✅ LaunchAgent working correctly
- ✅ Log viewing and management

### Quality
- ✅ Test coverage >80%
- ✅ No critical bugs
- ✅ Error handling for all failure modes
- ✅ Clean, maintainable code

### Documentation
- ✅ Complete README with examples
- ✅ User guide for all features
- ✅ Developer documentation
- ✅ Inline code comments

### User Experience
- ✅ Installation completes in <5 minutes
- ✅ Configuration wizard is clear and guided
- ✅ Real-time feedback during operations
- ✅ Helpful error messages
- ✅ Keyboard shortcuts documented

---

## Future Enhancements (Post-MVP)

1. **Multi-Profile Support**
   - Multiple backup configurations
   - Switch between profiles

2. **Notifications**
   - macOS notifications on completion
   - Email alerts on failure

3. **Scheduling Options**
   - Custom schedules (weekly, daily)
   - Time zone support

4. **Backup Verification**
   - Checksum validation
   - Integrity checks

5. **Restore Functionality**
   - Download from cloud
   - Selective restore

6. **Cloud Provider Support**
   - AWS S3
   - Google Cloud Storage
   - Azure Blob Storage

7. **Dashboard**
   - Graphical statistics
   - Trend analysis
   - Storage usage charts

8. **Web Interface**
   - Optional web UI
   - Remote management

---

## Risk Mitigation

### Technical Risks
1. **Risk**: rclone API changes
   - **Mitigation**: Pin rclone version, test upgrades

2. **Risk**: launchd behavior differences across macOS versions
   - **Mitigation**: Test on multiple macOS versions

3. **Risk**: Large file transfers timeout
   - **Mitigation**: Implement retry logic, progress tracking

### User Experience Risks
1. **Risk**: Complex configuration confuses users
   - **Mitigation**: Step-by-step wizard, sensible defaults

2. **Risk**: Users lose data due to sync errors
   - **Mitigation**: Dry-run option, confirmation prompts

---

## Maintenance Plan

### Regular Updates
- Monitor rclone updates
- Test with new macOS releases
- Update dependencies quarterly

### Bug Tracking
- GitHub Issues for bug reports
- Label and prioritize issues
- Weekly review of open issues

### Support
- Troubleshooting guide in docs
- Common issues FAQ
- GitHub Discussions for questions

---

## Metrics & Monitoring

### Application Metrics
- Backup success rate
- Average backup duration
- Error frequency by type
- User engagement (feature usage)

### Performance Metrics
- UI responsiveness
- Log parsing speed
- Memory usage
- CPU usage during sync

### Collection Method
- Local metrics file (opt-in)
- No external telemetry
- Privacy-focused

---

## License
MIT License (or user preference)

---

## Project Timeline
**Estimated Duration**: 20 working days
**Start Date**: November 2025
**Target Completion**: December 2025

---

## Status Legend
- ✓ Complete
- ⚠ In Progress
- ○ Not Started
- ✗ Blocked

---

## Notes
- This is a living document and will be updated as the project progresses
- All dates are estimates and subject to change
- Features may be reprioritized based on feedback
- Testing is continuous throughout development, not just at the end