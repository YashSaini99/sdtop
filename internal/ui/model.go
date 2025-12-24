package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sdtop/internal/systemd"
	"sdtop/internal/types"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model is the Bubble Tea model for the TUI
type Model struct {
	serviceList     list.Model
	logViewport     viewport.Model
	services        []types.Service
	allServices     []types.Service // Keep unfiltered list
	currentService  string
	logs            []types.LogEntry
	processes       []*types.Process
	manager         *systemd.Manager
	logReader       *systemd.LogReader
	processManager  *systemd.ProcessManager
	logCancel       context.CancelFunc
	statusMsg       string
	errMsg          string
	width           int
	height          int
	ready           bool
	filterMode      string // "all", "running", "failed"
	showProcessTree bool   // Toggle between logs and process tree
}

// serviceItem wraps a service for the list
type serviceItem struct {
	service types.Service
}

func (i serviceItem) Title() string {
	return i.service.Name
}

func (i serviceItem) Description() string {
	// Color-code the state
	state := i.service.SubState
	if state == "" {
		state = i.service.ActiveState
	}

	var stateColor lipgloss.Color
	var stateSymbol string

	switch state {
	case "running":
		stateColor = lipgloss.Color("42") // Green
		stateSymbol = "●"
	case "exited":
		stateColor = lipgloss.Color("240") // Gray
		stateSymbol = "○"
	case "failed":
		stateColor = lipgloss.Color("196") // Red
		stateSymbol = "✗"
	case "dead":
		stateColor = lipgloss.Color("240") // Gray
		stateSymbol = "○"
	case "active":
		stateColor = lipgloss.Color("42") // Green
		stateSymbol = "●"
	case "inactive":
		stateColor = lipgloss.Color("240") // Gray
		stateSymbol = "○"
	default:
		stateColor = lipgloss.Color("226") // Yellow
		stateSymbol = "◐"
	}

	styledState := lipgloss.NewStyle().
		Foreground(stateColor).
		Render(fmt.Sprintf("%s %s", stateSymbol, state))

	desc := i.service.Description
	if len(desc) > 40 {
		desc = desc[:37] + "..."
	}

	// Show if enabled on boot
	bootStatus := ""
	if i.service.LoadState == "loaded" {
		if strings.Contains(i.service.UnitFileState, "enabled") {
			bootStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(" [boot]")
		}
	}

	return fmt.Sprintf("%s %s%s", styledState, desc, bootStatus)
}

func (i serviceItem) FilterValue() string {
	return i.service.Name
}

// tickMsg is sent periodically to update logs
type tickMsg time.Time

// logUpdateMsg signals to refresh logs
type logUpdateMsg struct{}

// statusMsg shows temporary status messages
type statusMsgType string

// NewModel creates a new UI model
func NewModel(manager *systemd.Manager, logReader *systemd.LogReader) (*Model, error) {
	// Create list
	delegate := list.NewDefaultDelegate()
	serviceList := list.New([]list.Item{}, delegate, 0, 0)
	serviceList.Title = "SYSTEMD SERVICES"
	serviceList.SetShowStatusBar(false)
	serviceList.SetFilteringEnabled(true)

	// Customize list styles
	serviceList.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Background(lipgloss.Color("235")).
		Padding(0, 1)

	return &Model{
		serviceList:    serviceList,
		manager:        manager,
		logReader:      logReader,
		processManager: systemd.NewProcessManager(),
		logs:           []types.LogEntry{},
		filterMode:     "all",
	}, nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadServices,
		m.tickCmd(),
	)
}

// loadServices fetches services from systemd
func (m *Model) loadServices() tea.Msg {
	services, err := m.manager.ListServices()
	if err != nil {
		return systemd.ErrorMsg(fmt.Sprintf("Failed to list services: %v", err))
	}

	items := make([]list.Item, len(services))
	for i, svc := range services {
		items[i] = serviceItem{service: svc}
	}

	return servicesLoadedMsg{services: services, items: items}
}

// servicesLoadedMsg is sent when services are loaded
type servicesLoadedMsg struct {
	services []types.Service
	items    []list.Item
}

// tickCmd creates a ticker for log updates
func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.logCancel != nil {
				m.logCancel()
			}
			return m, tea.Quit

		case "enter":
			// Select service
			if item, ok := m.serviceList.SelectedItem().(serviceItem); ok {
				return m, m.selectService(item.service.Name)
			}

		case "r":
			// Restart service
			if m.currentService != "" {
				return m, m.restartService()
			}

		case "s":
			// Stop service
			if m.currentService != "" {
				return m, m.stopService()
			}

		case "t":
			// Start service
			if m.currentService != "" {
				return m, m.startService()
			}

		case "e":
			// Enable service on boot
			if m.currentService != "" {
				return m, m.enableService()
			}

		case "d":
			// Disable service on boot
			if m.currentService != "" {
				return m, m.disableService()
			}

		case "f":
			// Cycle through filters
			return m, m.cycleFilter()

		case "1":
			// Show all services
			m.filterMode = "all"
			return m, m.applyFilter()

		case "2":
			// Show only running
			m.filterMode = "running"
			return m, m.applyFilter()

		case "3":
			// Show only failed
			m.filterMode = "failed"
			return m, m.applyFilter()

		case "p":
			// Toggle process tree view
			if m.currentService != "" {
				m.showProcessTree = !m.showProcessTree
				if m.showProcessTree {
					return m, m.loadProcessTree()
				}
			}
			return m, nil

		case "l":
			// Back to logs view (if in process tree)
			if m.showProcessTree {
				m.showProcessTree = false
			}
			return m, nil

		case "up", "k":
			var cmd tea.Cmd
			m.serviceList, cmd = m.serviceList.Update(msg)
			return m, cmd

		case "down", "j":
			var cmd tea.Cmd
			m.serviceList, cmd = m.serviceList.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Split layout: 30% left, 70% right
		leftWidth := msg.Width * 30 / 100
		rightWidth := msg.Width - leftWidth - 2

		m.serviceList.SetSize(leftWidth, msg.Height-4)

		if !m.ready {
			m.logViewport = viewport.New(rightWidth, msg.Height-4)
			m.logViewport.YPosition = 0
			m.ready = true
		} else {
			m.logViewport.Width = rightWidth
			m.logViewport.Height = msg.Height - 4
		}

		return m, nil

	case servicesLoadedMsg:
		m.services = msg.services
		m.allServices = msg.services
		m.serviceList.SetItems(msg.items)
		return m, nil

	case systemd.ErrorMsg:
		m.errMsg = string(msg)
		return m, nil

	case statusMsgType:
		m.statusMsg = string(msg)
		// Clear status after 2 seconds
		return m, tea.Tick(time.Second*2, func(time.Time) tea.Msg {
			return clearStatusMsg{}
		})

	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil

	case tickMsg:
		// Update logs only if NOT viewing process tree
		if m.currentService != "" && !m.showProcessTree {
			return m, tea.Batch(m.tickCmd(), m.refreshLogs())
		}
		return m, m.tickCmd()

	case logsLoadedMsg:
		// Only update if not viewing process tree
		if !m.showProcessTree {
			m.logs = msg.logs
			m.logViewport.SetContent(m.formatLogs())
			m.logViewport.GotoBottom()
		}
		return m, nil

	case processesLoadedMsg:
		m.processes = msg.processes
		m.logViewport.SetContent(m.formatProcessTree())
		m.logViewport.GotoTop()
		return m, nil
	}

	// Update list
	var cmd tea.Cmd
	m.serviceList, cmd = m.serviceList.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport
	m.logViewport, cmd = m.logViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// selectService switches to viewing logs for a service
func (m *Model) selectService(serviceName string) tea.Cmd {
	// Cancel previous log stream
	if m.logCancel != nil {
		m.logCancel()
	}

	m.currentService = serviceName
	m.logs = []types.LogEntry{}

	// Create new context for log streaming
	_, cancel := context.WithCancel(context.Background())
	m.logCancel = cancel

	return m.refreshLogs()
}

// refreshLogs loads recent logs for the current service
func (m *Model) refreshLogs() tea.Cmd {
	return func() tea.Msg {
		logs, err := m.logReader.GetRecentLogs(m.currentService, 100)
		if err != nil {
			return systemd.ErrorMsg(fmt.Sprintf("Failed to read logs: %v", err))
		}
		return logsLoadedMsg{logs: logs}
	}
}

// logsLoadedMsg is sent when logs are loaded
type logsLoadedMsg struct {
	logs []types.LogEntry
}

// processesLoadedMsg is sent when process tree is loaded
type processesLoadedMsg struct {
	processes []*types.Process
}

// clearStatusMsg clears the status message
type clearStatusMsg struct{}

// loadProcessTree loads the process tree for current service
func (m *Model) loadProcessTree() tea.Cmd {
	return func() tea.Msg {
		processes, err := m.processManager.GetServiceProcesses(m.currentService)
		if err != nil {
			return systemd.ErrorMsg(fmt.Sprintf("Failed to load processes: %v", err))
		}
		return processesLoadedMsg{processes: processes}
	}
}

// restartService restarts the current service
func (m *Model) restartService() tea.Cmd {
	return func() tea.Msg {
		if err := m.manager.RestartService(m.currentService); err != nil {
			return systemd.ErrorMsg(fmt.Sprintf("Failed to restart: %v", err))
		}
		return statusMsgType(fmt.Sprintf("Restarting %s...", m.currentService))
	}
}

// stopService stops the current service
func (m *Model) stopService() tea.Cmd {
	return func() tea.Msg {
		if err := m.manager.StopService(m.currentService); err != nil {
			return systemd.ErrorMsg(fmt.Sprintf("Failed to stop: %v", err))
		}
		return statusMsgType(fmt.Sprintf("Stopping %s...", m.currentService))
	}
}

// startService starts the current service
func (m *Model) startService() tea.Cmd {
	return func() tea.Msg {
		if err := m.manager.StartService(m.currentService); err != nil {
			return systemd.ErrorMsg(fmt.Sprintf("Failed to start: %v", err))
		}
		return statusMsgType(fmt.Sprintf("Starting %s...", m.currentService))
	}
}

// enableService enables the current service on boot
func (m *Model) enableService() tea.Cmd {
	return func() tea.Msg {
		if err := m.manager.EnableService(m.currentService); err != nil {
			return systemd.ErrorMsg(fmt.Sprintf("Failed to enable: %v", err))
		}
		return statusMsgType(fmt.Sprintf("Enabled %s on boot ✓", m.currentService))
	}
}

// disableService disables the current service on boot
func (m *Model) disableService() tea.Cmd {
	return func() tea.Msg {
		if err := m.manager.DisableService(m.currentService); err != nil {
			return systemd.ErrorMsg(fmt.Sprintf("Failed to disable: %v", err))
		}
		return statusMsgType(fmt.Sprintf("Disabled %s from boot", m.currentService))
	}
}

// cycleFilter cycles through filter modes
func (m *Model) cycleFilter() tea.Cmd {
	switch m.filterMode {
	case "all":
		m.filterMode = "running"
	case "running":
		m.filterMode = "failed"
	case "failed":
		m.filterMode = "all"
	}
	return m.applyFilter()
}

// applyFilter applies the current filter mode
func (m *Model) applyFilter() tea.Cmd {
	return func() tea.Msg {
		var filtered []types.Service

		switch m.filterMode {
		case "running":
			for _, svc := range m.allServices {
				if svc.SubState == "running" || svc.ActiveState == "active" {
					filtered = append(filtered, svc)
				}
			}
		case "failed":
			for _, svc := range m.allServices {
				if svc.ActiveState == "failed" || svc.SubState == "failed" {
					filtered = append(filtered, svc)
				}
			}
		default: // "all"
			filtered = m.allServices
		}

		items := make([]list.Item, len(filtered))
		for i, svc := range filtered {
			items[i] = serviceItem{service: svc}
		}

		return servicesLoadedMsg{services: filtered, items: items}
	}
}

// formatLogs formats logs for display
func (m *Model) formatLogs() string {
	if len(m.logs) == 0 {
		if m.currentService == "" {
			return m.renderEmptyState()
		}
		return m.renderNoLogsState()
	}

	var sb strings.Builder
	for _, log := range m.logs {
		timestamp := log.Timestamp.Format("15:04:05")

		// Color-code by priority
		var lineStyle lipgloss.Style
		priorityIcon := "  "

		switch log.Priority {
		case "error":
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			priorityIcon = "✗ "
		case "warn":
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
			priorityIcon = "⚠ "
		default:
			lineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
			priorityIcon = "  "
		}

		timestampStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Render(timestamp)

		line := fmt.Sprintf("%s %s%s\n", timestampStyle, priorityIcon, log.Message)
		sb.WriteString(lineStyle.Render(line))
	}
	return sb.String()
}

// formatProcessTree formats the process tree for display
func (m *Model) formatProcessTree() string {
	if len(m.processes) == 0 {
		return m.renderNoProcessesState()
	}

	var sb strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	sb.WriteString(headerStyle.Render(fmt.Sprintf("Process Tree for %s\n\n", m.currentService)))

	for _, proc := range m.processes {
		m.renderProcess(&sb, proc, "", true)
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("Press 'l' to return to logs view"))

	return sb.String()
}

// renderProcess recursively renders a process and its children
func (m *Model) renderProcess(sb *strings.Builder, proc *types.Process, prefix string, isLast bool) {
	pidStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	cmdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	// Tree characters
	var connector string
	if prefix == "" {
		connector = "┌─"
	} else if isLast {
		connector = "└─"
	} else {
		connector = "├─"
	}

	// Format: ├─ [PID] name: cmdline
	line := fmt.Sprintf("%s%s %s %s: %s\n",
		prefix,
		connector,
		pidStyle.Render(fmt.Sprintf("[%d]", proc.PID)),
		nameStyle.Render(proc.Name),
		cmdStyle.Render(truncate(proc.Cmdline, 60)),
	)
	sb.WriteString(line)

	// Render children
	childPrefix := prefix
	if prefix == "" {
		childPrefix = "  "
	} else if isLast {
		childPrefix = prefix + "   "
	} else {
		childPrefix = prefix + "│  "
	}

	for i, child := range proc.Children {
		m.renderProcess(sb, child, childPrefix, i == len(proc.Children)-1)
	}
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// renderNoProcessesState shows message when no processes found
func (m *Model) renderNoProcessesState() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(m.logViewport.Width).
		MarginTop(m.logViewport.Height / 3)

	content := "NO PROCESSES FOUND\n\n" +
		fmt.Sprintf("Service: %s\n\n", m.currentService) +
		"The service may not be running,\n" +
		"or you may need root permissions\n" +
		"to view process information.\n\n" +
		"Try: sudo sdtop\n\n" +
		"Press 'l' to return to logs"

	return style.Render(content)
}

// renderEmptyState shows helpful message when no service selected
func (m *Model) renderEmptyState() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("170")).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("42")).
		Bold(true)

	// Build the content
	var content strings.Builder
	content.WriteString("\n\n")
	content.WriteString(titleStyle.Render("👈 SELECT A SERVICE TO GET STARTED"))
	content.WriteString("\n\n\n")
	content.WriteString(labelStyle.Render("Navigation:\n"))
	content.WriteString("  " + keyStyle.Render("↑↓") + labelStyle.Render(" or ") + keyStyle.Render("j/k") + labelStyle.Render(" - Move up/down\n"))
	content.WriteString("  " + keyStyle.Render("Enter") + labelStyle.Render(" - View logs for selected service\n"))
	content.WriteString("  " + keyStyle.Render("/") + labelStyle.Render(" - Filter/search services\n\n"))
	content.WriteString(labelStyle.Render("Service Actions:\n"))
	content.WriteString("  " + keyStyle.Render("r") + labelStyle.Render(" - Restart service\n"))
	content.WriteString("  " + keyStyle.Render("s") + labelStyle.Render(" - Stop service\n"))
	content.WriteString("  " + keyStyle.Render("t") + labelStyle.Render(" - Start service\n"))
	content.WriteString("  " + keyStyle.Render("e") + labelStyle.Render(" - Enable on boot\n"))
	content.WriteString("  " + keyStyle.Render("d") + labelStyle.Render(" - Disable from boot\n\n"))
	content.WriteString(labelStyle.Render("View Modes:\n"))
	content.WriteString("  " + keyStyle.Render("p") + labelStyle.Render(" - Show process tree (see what's running!)\n"))
	content.WriteString("  " + keyStyle.Render("l") + labelStyle.Render(" - Return to logs view\n\n"))
	content.WriteString(labelStyle.Render("Filters:\n"))
	content.WriteString("  " + keyStyle.Render("f") + labelStyle.Render(" - Cycle filter (all → running → failed)\n"))
	content.WriteString("  " + keyStyle.Render("1") + labelStyle.Render(" - Show all services\n"))
	content.WriteString("  " + keyStyle.Render("2") + labelStyle.Render(" - Show only running\n"))
	content.WriteString("  " + keyStyle.Render("3") + labelStyle.Render(" - Show only failed\n\n"))
	content.WriteString(labelStyle.Render("Other:\n"))
	content.WriteString("  " + keyStyle.Render("q") + labelStyle.Render(" - Quit application\n"))

	style := lipgloss.NewStyle().
		Width(m.logViewport.Width).
		Align(lipgloss.Center).
		MarginTop(m.logViewport.Height / 5)

	return style.Render(content.String())
}

// renderNoLogsState shows message when service has no logs
func (m *Model) renderNoLogsState() string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Align(lipgloss.Center).
		Width(m.logViewport.Width).
		MarginTop(m.logViewport.Height / 3)

	content := "NO LOGS FOUND\n\n" +
		fmt.Sprintf("Service: %s\n\n", m.currentService) +
		"The service may not have generated\n" +
		"any logs recently, or you may need\n" +
		"additional permissions.\n\n" +
		"Try: sudo usermod -a -G systemd-journal $USER"

	return style.Render(content)
}

// View renders the UI
func (m *Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Styles
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Left pane: service list
	leftWidth := m.width * 30 / 100
	leftPane := borderStyle.
		Width(leftWidth).
		Height(m.height - 4).
		Render(m.serviceList.View())

	// Right pane: logs
	rightWidth := m.width - leftWidth - 2

	var logTitle string
	if m.currentService != "" {
		// Show service name and actions
		serviceName := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("170")).
			Render(m.currentService)

		var modeAndActions string
		if m.showProcessTree {
			modeAndActions = lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Render(" 🌳 PROCESS TREE [l]ogs")
		} else {
			modeAndActions = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render(" [r]estart [s]top [t]art [p]rocesses")
		}

		logTitle = lipgloss.NewStyle().
			Background(lipgloss.Color("235")).
			Width(rightWidth).
			Padding(0, 1).
			Render(fmt.Sprintf("LOGS: %s %s", serviceName, modeAndActions))
	} else {
		logTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("240")).
			Background(lipgloss.Color("235")).
			Width(rightWidth).
			Padding(0, 1).
			Render("LOGS")
	}

	rightPane := borderStyle.
		Width(rightWidth).
		Height(m.height - 4).
		Render(lipgloss.JoinVertical(lipgloss.Left, logTitle, m.logViewport.View()))

	// Combine panes
	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, rightPane)

	// Status bar
	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

// renderStatusBar renders the bottom status bar
func (m *Model) renderStatusBar() string {
	// If there's a status or error message, show it prominently
	if m.statusMsg != "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true).
			Render(fmt.Sprintf("✓ %s", m.statusMsg))
	}

	if m.errMsg != "" {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Render(fmt.Sprintf("✗ Error: %s", m.errMsg))
	}

	// Build context-aware help text
	var helpParts []string

	// Navigation always available
	helpParts = append(helpParts,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Navigate: "),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("↑↓/jk"),
	)

	// Selection
	helpParts = append(helpParts,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" • Select: "),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("enter"),
	)

	// Actions (only if service selected)
	if m.currentService != "" {
		if m.showProcessTree {
			helpParts = append(helpParts,
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" • View: "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("l"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("ogs"),
			)
		} else {
			helpParts = append(helpParts,
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" • Actions: "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("r"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("estart "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("s"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("top "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("t"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("start "),
				lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("p"),
				lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("rocesses"),
			)
		}
	}

	// Filter
	helpParts = append(helpParts,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" • Filter: "),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("f"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("/"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("1"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("all "),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("2"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("run "),
		lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render("3"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("fail"),
	)

	// Quit
	helpParts = append(helpParts,
		lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(" • Quit: "),
		lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("q"),
	)

	// Service count with filter indicator
	filterIndicator := ""
	switch m.filterMode {
	case "running":
		filterIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(" [RUNNING]")
	case "failed":
		filterIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(" [FAILED]")
	}

	serviceCount := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(fmt.Sprintf(" │ Services: %d%s", len(m.services), filterIndicator))

	helpParts = append(helpParts, serviceCount)

	return lipgloss.JoinHorizontal(lipgloss.Left, helpParts...)
}
