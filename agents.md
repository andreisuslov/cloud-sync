# Agent Development Instructions

## Project Information

- **GitHub Repository**: `github.com/andreisuslov/cloud-sync`
- **Author**: andreisuslov
- **Language**: Go (requires installation)
- **Framework**: Bubbletea (TUI)

## Development Workflow

### Commit Strategy

**CRITICAL**: Commit each and every change individually when it is created.

- **One file per commit**: Each new file created must be committed before moving to the next file
- **One modification per commit**: Each modification to an existing file must be committed before making the next change
- **Small, atomic commits**: Break down work into the smallest logical units
- **No batch commits**: Never accumulate multiple changes before committing

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only changes
- `style`: Formatting, missing semicolons, etc (no code change)
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `test`: Adding or updating tests
- `chore`: Changes to build process or auxiliary tools

**Examples:**
```bash
git commit -m "feat(installer): add CheckHomebrewInstalled function"
git commit -m "test(installer): add unit tests for homebrew check"
git commit -m "docs: update README with installation instructions"
git commit -m "fix(ui): correct import path for models package"
```

### Git Workflow for Each Change

```bash
# 1. Create or modify a file
# 2. Stage the file
git add <filename>

# 3. Commit immediately with descriptive message
git commit -m "<type>(<scope>): <description>"

# 4. Then and only then, move to the next file/change
```

### Example Session

```bash
# Create agents.md
touch agents.md
# ... add content ...
git add agents.md
git commit -m "docs: add agent development instructions"

# Fix go.mod module name
# ... edit go.mod ...
git add go.mod
git commit -m "fix(module): correct GitHub username to andreisuslov"

# Fix import in main.go
# ... edit cmd/cloud-sync/main.go ...
git add cmd/cloud-sync/main.go
git commit -m "fix(main): update import path to use andreisuslov"

# And so on for each file...
```

## Prerequisites

### Required Software

1. **Go**: Must be installed before running any Go commands
   ```bash
   # Check if Go is installed
   which go
   
   # If not installed, install via Homebrew
   brew install go
   ```

2. **Git**: For version control (usually pre-installed on macOS)
   ```bash
   git --version
   ```

3. **Homebrew**: Package manager for macOS
   ```bash
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```

### First-Time Setup

```bash
# Navigate to project
cd /Users/ansuslov/Documents/Development/cloud-sync

# Verify Go is installed
go version

# Download dependencies
go mod download

# Verify dependencies
go mod tidy

# Run tests
go test ./...
```

## Development Guidelines

### Before Making Changes

1. Verify Go is installed: `go version`
2. Ensure you're on the correct branch
3. Pull latest changes: `git pull`
4. Create feature branch if needed: `git checkout -b feature/name`

### After Each Change

1. Stage the specific file: `git add <file>`
2. Commit with descriptive message: `git commit -m "type(scope): message"`
3. Verify commit was successful: `git log -1`
4. Continue to next change

### Before Pushing

1. Run tests: `go test ./...`
2. Check for lint errors: `go vet ./...`
3. Format code: `go fmt ./...`
4. Push commits: `git push origin <branch>`

## Testing

### Run All Tests
```bash
go test ./...
```

### Run Specific Package Tests
```bash
go test ./internal/installer
go test ./tests/unit/...
```

### Run with Coverage
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run with Verbose Output
```bash
go test -v ./...
```

## Building

### Development Build
```bash
go build -o cloud-sync ./cmd/cloud-sync
```

### Production Build with Version
```bash
go build -ldflags "-X main.Version=1.0.0" -o cloud-sync ./cmd/cloud-sync
```

### Cross-Compilation
```bash
# Intel Mac
GOARCH=amd64 go build -o cloud-sync-amd64 ./cmd/cloud-sync

# Apple Silicon
GOARCH=arm64 go build -o cloud-sync-arm64 ./cmd/cloud-sync
```

## Troubleshooting

### Go Not Found
```bash
# Install Go via Homebrew
brew install go

# Verify installation
go version

# If still not found, add to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
source ~/.zshrc
```

### Import Path Errors
- Ensure all imports use `github.com/andreisuslov/cloud-sync`
- Run `go mod tidy` to update dependencies
- Check go.mod has correct module name

### Permission Denied
```bash
# Make scripts executable
chmod +x script.sh
```

## Project Structure

```
cloud-sync/
├── cmd/cloud-sync/          # Main application entry
├── internal/                # Private application code
│   ├── installer/          # Tool installation
│   ├── rclone/             # Rclone operations
│   ├── scripts/            # Script generation
│   ├── launchd/            # LaunchAgent management
│   ├── logs/               # Log parsing
│   ├── lockfile/           # Lockfile management
│   └── ui/                 # TUI components
├── pkg/                    # Public API
├── scripts/                # Script templates
├── tests/                  # Test files
│   ├── unit/              # Unit tests
│   ├── integration/       # Integration tests
│   └── testutil/          # Test helpers
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
├── README.md               # Project README
├── project-plan.md         # Implementation roadmap
├── agents.md               # This file
└── .gitignore              # Git ignore rules
```

## Key Principles

1. **Commit Frequently**: Every single file creation or modification
2. **Small Commits**: One logical change per commit
3. **Clear Messages**: Descriptive commit messages following convention
4. **Test First**: Run tests before pushing
5. **Go Required**: Verify Go installation before any Go operations
6. **Correct Paths**: Always use `github.com/andreisuslov/cloud-sync` in imports

## Remember

- ✅ Commit each file individually
- ✅ Do not commit the changes before you write test cases for them and run them successfully
- ✅ Commit each modification individually  
- ✅ Use correct GitHub username: `andreisuslov`
- ✅ Verify Go is installed before Go commands
- ❌ Never batch multiple changes in one commit
- ❌ Never skip committing after a change
- ❌ Never use wrong username in imports