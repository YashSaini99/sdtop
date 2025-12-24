package systemd

import (
	"strings"

	"sdtop/internal/types"

	"github.com/coreos/go-systemd/v22/dbus"
)

// Manager handles systemd service operations
type Manager struct {
	conn *dbus.Conn
}

// NewManager creates a new systemd manager
func NewManager() (*Manager, error) {
	conn, err := dbus.NewSystemConnection()
	if err != nil {
		return nil, err
	}
	return &Manager{conn: conn}, nil
}

// Close closes the DBus connection
func (m *Manager) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}

// ListServices fetches all systemd services
func (m *Manager) ListServices() ([]types.Service, error) {
	units, err := m.conn.ListUnits()
	if err != nil {
		return nil, err
	}

	var services []types.Service
	for _, unit := range units {
		// Filter only .service units
		if !strings.HasSuffix(unit.Name, ".service") {
			continue
		}

		services = append(services, types.Service{
			Name:        unit.Name,
			Description: unit.Description,
			ActiveState: unit.ActiveState,
			SubState:    unit.SubState,
			LoadState:   unit.LoadState,
		})
	}

	return services, nil
}

// RestartService restarts a systemd service
func (m *Manager) RestartService(serviceName string) error {
	_, err := m.conn.RestartUnit(serviceName, "replace", nil)
	return err
}

// StopService stops a systemd service
func (m *Manager) StopService(serviceName string) error {
	_, err := m.conn.StopUnit(serviceName, "replace", nil)
	return err
}

// StartService starts a systemd service
func (m *Manager) StartService(serviceName string) error {
	_, err := m.conn.StartUnit(serviceName, "replace", nil)
	return err
}

// EnableService enables a service to start on boot
func (m *Manager) EnableService(serviceName string) error {
	_, _, err := m.conn.EnableUnitFiles([]string{serviceName}, false, true)
	if err != nil {
		return err
	}
	m.conn.Reload()
	return nil
}

// DisableService disables a service from starting on boot
func (m *Manager) DisableService(serviceName string) error {
	_, err := m.conn.DisableUnitFiles([]string{serviceName}, false)
	if err != nil {
		return err
	}
	m.conn.Reload()
	return nil
}

// GetServiceProperty gets a property of a service
func (m *Manager) GetServiceProperty(serviceName, property string) (interface{}, error) {
	return m.conn.GetServiceProperty(serviceName, property)
}
