# sdtop - systemd TUI Manager

A **terminal-based interactive dashboard** for managing and debugging systemd services on Linux.

## ✨ Features

### Service Management
- 🔍 View all systemd services with **color-coded status indicators**
  - ● Green = Running
  - ✗ Red = Failed
  - ○ Gray = Stopped/Dead
- ⚡ Control services: start, stop, restart with **visual feedback**
- 🎯 Enable/disable services on boot
- 🔎 Filter services: all, running, failed

### Real-time Monitoring
- 📊 **Live log streaming** from journald with **priority highlighting**
  - ✗ Red for errors
  - ⚠ Yellow for warnings
  - Auto-scrolling with timestamps
- 🌳 **Process Tree View** - See what's actually running!
  - Shows all processes for a service
  - Parent-child relationships
  - PIDs and command lines
  - Debug zombie processes
  - Understand CPU usage

### User Experience
- 🎨 **Context-aware help** and guidance
- 💡 **Intuitive empty states** that explain what to do
- ⌨️  **Keyboard-driven interface** with vi-style shortcuts
- 🚀 Fast, responsive, built with Go and Bubble Tea

## Requirements

- Linux with systemd
- journald enabled
- Go ≥ 1.21 (for building)

## Installation

### Package Managers

**Arch Linux (AUR):**
```bash
yay -S sdtop
# or
paru -S sdtop
```

**Debian/Ubuntu (PPA):**
```bash
sudo add-apt-repository ppa:YashSaini99/sdtop
sudo apt update
sudo apt install sdtop
```

**Fedora/RHEL (COPR):**
```bash
sudo dnf copr enable YashSaini99/sdtop
sudo dnf install sdtop
```

**Flatpak:**
```bash
flatpak install flathub io.github.YashSaini99.sdtop
flatpak run io.github.YashSaini99.sdtop
```

**Snap:**
```bash
sudo snap install sdtop --classic
```

**Homebrew (Linux):**
```bash
brew install sdtop
```

### Pre-built Binaries

**One-line install (recommended):**
```bash
curl -sL https://raw.githubusercontent.com/YashSaini99/sdtop/main/install.sh | bash
```

**Manual download:**
```bash
# Download from GitHub releases
wget https://github.com/YashSaini99/sdtop/releases/latest/download/sdtop-linux-amd64.tar.gz
tar -xzf sdtop-linux-amd64.tar.gz
sudo mv sdtop /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/YashSaini99/sdtop.git
cd sdtop
go build -o sdtop ./cmd/main.go
sudo mv sdtop /usr/local/bin/
```

### Run without installing

```bash
go run ./cmd/main.go
```

## Usage

Launch the application:

```bash
sdtop
```

### Keyboard Controls

| Key | Action |
|-----|--------|
| `↑` / `k` | Move selection up |
| `↓` / `j` | Move selection down |
| `Enter` | Select service and view logs |
| **Service Control** ||
| `r` | Restart selected service |
| `s` | Stop selected service |
| `t` | Start selected service |
| `e` | Enable service on boot |
| `d` | Disable service from boot |
| **View Modes** ||
| `p` | Show process tree 🌳 |
| `l` | Return to logs view |
| **Filtering** ||
| `f` | Cycle filters (all → running → failed) |
| `/` | Search/filter services |
| `1` | Show all services |
| `2` | Show only running |
| `3` | Show only failed |
| **Other** ||
| `q` | Quit application |

### Permissions

To view logs and control services, you may need appropriate permissions:

- **Read logs**: User must be in `systemd-journal` group or run as root
- **Control services**: Requires PolicyKit authorization or root access

Add user to journal group:
```bash
sudo usermod -a -G systemd-journal $USER
```

For service control without root, configure PolicyKit rules or use `sudo`:
```bash
sudo sdtop
```

## Architecture

```
sdtop/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── systemd/
│   │   ├── services.go      # DBus service operations (start/stop/restart)
│   │   ├── logs.go          # Journald log streaming
│   │   └── processes.go     # Process tree from /proc filesystem
│   ├── ui/
│   │   └── model.go         # Bubble Tea UI (MVC pattern)
│   └── types/
│       └── models.go        # Data structures (Service, LogEntry, Process)
├── go.mod
└── README.md
```

### How It Works

**Service Control** - Uses `go-systemd/dbus` to communicate with systemd:
- `ListUnits()` → fetches all services
- `StartUnit()`, `StopUnit()`, `RestartUnit()` → control services
- Real-time state updates

**Log Streaming** - Uses `go-systemd/sdjournal` to read logs:
- `AddMatch()` → filter by service name
- `SeekTail()` → start from recent logs
- Continuous polling for new entries

**Process Tree** - Reads `/proc` filesystem directly:
- Scans `/proc/[pid]/cgroup` → finds processes belonging to service
- Reads `/proc/[pid]/stat` → gets process name and parent PID
- Reads `/proc/[pid]/cmdline` → gets command line
- Builds parent-child tree structure

**UI Framework** - Bubble Tea (Elm Architecture):
- **Model** - Application state (services, logs, processes)
- **Update** - Handles events (keypresses, data updates)
- **View** - Renders UI from state

### Key Components

**Bubble Tea Pattern:**
```go
Init()      → Load initial data, start background tasks
Update(msg) → Handle events, return new state + commands
View()      → Render current state to terminal
```

**Concurrency:**
- Log ticker runs every 500ms (when viewing logs)
- Context cancellation prevents memory leaks
- Background goroutines for data fetching

## Use Cases

**Debug a failing service:**
1. Press `3` to show only failed services
2. Select the service, view error logs
3. Press `p` to see if processes are stuck
4. Press `r` to restart

**Monitor system services:**
1. Press `2` to show running services
2. Select a service to watch live logs
3. See real-time activity

**Understand what's running:**
1. Select any service
2. Press `p` to see process tree
3. Identify child processes and their relationships

**Control boot services:**
1. Select a service
2. Press `e` to enable on boot
3. Press `d` to disable from boot

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [go-systemd](https://github.com/coreos/go-systemd) - systemd integration

## License

[LICENSE](https://github.com/YashSaini99/sdtop/blob/main/LICENSE)
