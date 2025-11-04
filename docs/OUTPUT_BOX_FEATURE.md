# Output Box Feature

## Overview

Added a separate output box to the installation view that displays command output without breaking the UI layout or borders.

## Features

### 1. Separate Output Panel
- **Location**: Appears below the main installation menu
- **Styling**: Rounded border with consistent color scheme
- **Layout**: Vertically stacked using `lipgloss.JoinVertical`
- **Size**: Automatically adjusts to screen width (width - 8 padding)

### 2. Dynamic Display
- **Show/Hide**: Output box only appears when there's output to display
- **Buffer Management**: Stores last 50 lines of output
- **Display Limit**: Shows last 15 lines in the UI to prevent overflow
- **Auto-scroll**: Always shows most recent output

### 3. Screen Layout
```
┌─────────────────────────────────────┐
│  Installation & Setup               │
│  1. Check rsync Installation        │
│  2. Check rsync Version             │
│  3. Install/Update rsync       ← Selected
│  ...                                │
│                                     │
│  ↑/↓: navigate • enter: execute    │
│  ✓ Status message here              │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│  Command Output:                    │
│                                     │
│  ==> Downloading rsync...           │
│  ==> Installing rsync...            │
│  🍺 rsync 3.2.7 installed           │
└─────────────────────────────────────┘
```

## Implementation Details

### Model Changes
```go
type InstallationModel struct {
    // ... existing fields
    outputBuffer []string // Buffer to store command output
    showOutput   bool     // Whether to show the output box
}
```

### Message Structure
```go
type installStepCompleteMsg struct {
    step    string
    success bool
    err     error
    message string
    output  string // Command output to display
}
```

### New Installer Methods
- `InstallRsyncWithOutput() (string, error)` - Captures install output
- `UpdateRsyncWithOutput() (string, error)` - Captures update output

### Styling
```go
// Output box styles
outputBoxStyle = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(colorPrimary).
    Padding(1, 2).
    MarginTop(1)

outputTitleStyle = lipgloss.NewStyle().
    Foreground(colorPrimary).
    Bold(true).
    MarginBottom(1)

outputContentStyle = lipgloss.NewStyle().
    Foreground(colorSecondary)
```

## Usage

When you select "3. Install/Update rsync" and press Enter:

1. The command executes (e.g., `brew install rsync`)
2. Output is captured in real-time
3. Output box appears below the main menu
4. Shows the last 15 lines of output
5. UI borders remain intact
6. Status message updates at the top

## Benefits

- **Non-intrusive**: Doesn't break existing UI layout
- **Informative**: Shows actual command output
- **Clean**: Maintains border consistency
- **Responsive**: Adjusts to screen size
- **Efficient**: Limits buffer size to prevent memory issues
- **User-friendly**: Only shows when relevant

## Future Enhancements

Potential improvements:
- Scrollable output box for longer outputs
- Color-coded output (errors in red, success in green)
- Copy output to clipboard functionality
- Save output to log file
- Real-time streaming output (currently shows after completion)
