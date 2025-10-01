//go:build !linux

package watcher

import (
	"fmt"
	scanService "vm-server/services/scan"
)

type Manager struct{}

type Options struct{ Mount, Verbose bool }

func NewManager(_ *scanService.Service, _ Options) (*Manager, error) { return &Manager{}, nil }
func (m *Manager) ActivatePath(_ string) error                       { return fmt.Errorf("watcher disabled: non-linux") }
func (m *Manager) DeactivatePath(_ string) error                     { return nil }
func (m *Manager) Stop()                                             {}
