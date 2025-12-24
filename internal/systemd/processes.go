package systemd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sdtop/internal/types"
)

// ProcessManager handles process tree operations
type ProcessManager struct{}

// NewProcessManager creates a new process manager
func NewProcessManager() *ProcessManager {
	return &ProcessManager{}
}

// GetServiceProcesses returns the process tree for a service
func (pm *ProcessManager) GetServiceProcesses(serviceName string) ([]*types.Process, error) {
	// Get all PIDs for this service
	pids := pm.getAllServicePIDs(serviceName)

	if len(pids) == 0 {
		return nil, fmt.Errorf("service not running or no processes found")
	}

	// Build process tree
	var processes []*types.Process
	processMap := make(map[int]*types.Process)

	// First pass: create all process objects
	for _, pid := range pids {
		proc, err := pm.getProcessInfo(pid)
		if err != nil {
			continue
		}
		processMap[pid] = proc
		processes = append(processes, proc)
	}

	// Second pass: build parent-child relationships
	var roots []*types.Process
	for _, proc := range processes {
		parent, hasParent := processMap[proc.Parent]
		if hasParent {
			parent.Children = append(parent.Children, proc)
		} else {
			// This is a root process (parent not in our service)
			roots = append(roots, proc)
		}
	}

	if len(roots) == 0 && len(processes) > 0 {
		// If no roots found, use the first process
		return processes[:1], nil
	}

	return roots, nil
}

// getAllServicePIDs gets all PIDs for a service
func (pm *ProcessManager) getAllServicePIDs(serviceName string) []int {
	var pids []int
	cgroupPath := fmt.Sprintf("system.slice/%s", serviceName)

	procDir, err := os.ReadDir("/proc")
	if err != nil {
		return pids
	}

	for _, entry := range procDir {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Check if this process belongs to the service
		cgroupFile := fmt.Sprintf("/proc/%d/cgroup", pid)
		data, err := os.ReadFile(cgroupFile)
		if err != nil {
			continue
		}

		cgroupContent := string(data)
		if strings.Contains(cgroupContent, cgroupPath) ||
			strings.Contains(cgroupContent, serviceName) {
			pids = append(pids, pid)
		}
	}

	return pids
}

// getServiceMainPID gets the main PID of a service
func (pm *ProcessManager) getServiceMainPID(serviceName string) (int, error) {
	// Try multiple methods to find the service PID

	// Method 1: Check systemd cgroup
	cgroupPath := fmt.Sprintf("system.slice/%s", serviceName)

	procDir, err := os.ReadDir("/proc")
	if err != nil {
		return 0, err
	}

	// Collect all PIDs for this service
	var pids []int

	for _, entry := range procDir {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Check if this process belongs to the service
		cgroupFile := fmt.Sprintf("/proc/%d/cgroup", pid)
		data, err := os.ReadFile(cgroupFile)
		if err != nil {
			continue
		}

		cgroupContent := string(data)
		// Check for service in cgroup (handles both systemd v1 and v2)
		if strings.Contains(cgroupContent, cgroupPath) ||
			strings.Contains(cgroupContent, serviceName) {
			pids = append(pids, pid)
		}
	}

	if len(pids) == 0 {
		return 0, fmt.Errorf("no process found for service")
	}

	// Return the lowest PID (usually the parent)
	minPID := pids[0]
	for _, pid := range pids {
		if pid < minPID {
			minPID = pid
		}
	}

	return minPID, nil
}

// buildProcessTree builds a process tree from a root PID
func (pm *ProcessManager) buildProcessTree(pid int) (*types.Process, error) {
	proc, err := pm.getProcessInfo(pid)
	if err != nil {
		return nil, err
	}

	// Find children
	children := pm.findChildren(pid)
	proc.Children = make([]*types.Process, 0, len(children))

	for _, childPID := range children {
		child, err := pm.buildProcessTree(childPID)
		if err == nil {
			proc.Children = append(proc.Children, child)
		}
	}

	return proc, nil
}

// getProcessInfo reads process information from /proc
func (pm *ProcessManager) getProcessInfo(pid int) (*types.Process, error) {
	// Read command line
	cmdlineFile := fmt.Sprintf("/proc/%d/cmdline", pid)
	cmdlineData, err := os.ReadFile(cmdlineFile)
	if err != nil {
		return nil, err
	}

	cmdline := strings.ReplaceAll(string(cmdlineData), "\x00", " ")
	cmdline = strings.TrimSpace(cmdline)

	// Read stat for process name and parent
	statFile := fmt.Sprintf("/proc/%d/stat", pid)
	statData, err := os.ReadFile(statFile)
	if err != nil {
		return nil, err
	}

	// Parse stat file: PID (NAME) STATE PPID ...
	statStr := string(statData)
	startIdx := strings.Index(statStr, "(")
	endIdx := strings.LastIndex(statStr, ")")

	var name string
	var ppid int

	if startIdx != -1 && endIdx != -1 {
		name = statStr[startIdx+1 : endIdx]
		// Get PPID (4th field after closing paren)
		fields := strings.Fields(statStr[endIdx+1:])
		if len(fields) >= 2 {
			ppid, _ = strconv.Atoi(fields[1])
		}
	}

	if cmdline == "" {
		cmdline = fmt.Sprintf("[%s]", name)
	}

	return &types.Process{
		PID:     pid,
		Name:    name,
		Cmdline: cmdline,
		Parent:  ppid,
	}, nil
}

// findChildren finds all child processes of a given PID
func (pm *ProcessManager) findChildren(parentPID int) []int {
	var children []int

	procDir, err := os.ReadDir("/proc")
	if err != nil {
		return children
	}

	for _, entry := range procDir {
		if !entry.IsDir() {
			continue
		}

		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Read stat to check parent
		statFile := filepath.Join("/proc", entry.Name(), "stat")
		statData, err := os.ReadFile(statFile)
		if err != nil {
			continue
		}

		statStr := string(statData)
		endIdx := strings.LastIndex(statStr, ")")
		if endIdx == -1 {
			continue
		}

		fields := strings.Fields(statStr[endIdx+1:])
		if len(fields) >= 2 {
			ppid, _ := strconv.Atoi(fields[1])
			if ppid == parentPID {
				children = append(children, pid)
			}
		}
	}

	return children
}
