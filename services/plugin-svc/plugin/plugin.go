package plugin

import (
	"context"
	"errors"
	"sync"
)

// PluginMetadata contains basic information about a plugin.
type PluginMetadata struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Type        string `json:"type"` // e.g., "full", "partial"
	Author      string `json:"author"`
}

// MenuItem represents a menu item provided by a plugin.
type MenuItem struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Icon   string `json:"icon"`
	Path   string `json:"path"`
	Parent string `json:"parent,omitempty"`
	Order  int    `json:"order"`
}

// Plugin defines the interface that all plugins must implement.
type Plugin interface {
	// Metadata returns plugin metadata information.
	Metadata() PluginMetadata

	// Init initializes the plugin with the given context.
	Init(ctx context.Context) error

	// GetMenuItems returns the menu items provided by this plugin.
	GetMenuItems() []MenuItem

	// Close releases resources held by the plugin.
	Close() error
}

// PluginManager manages plugin registration and lifecycle.
type PluginManager struct {
	mu       sync.RWMutex
	plugins  map[string]Plugin
	metadata map[string]PluginMetadata
}

// NewPluginManager creates a new PluginManager instance.
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins:  make(map[string]Plugin),
		metadata: make(map[string]PluginMetadata),
	}
}

// Register registers a plugin with the manager.
func (pm *PluginManager) Register(p Plugin) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	meta := p.Metadata()
	if meta.Name == "" {
		return ErrInvalidPlugin
	}

	if _, exists := pm.plugins[meta.Name]; exists {
		return ErrPluginAlreadyRegistered
	}

	pm.plugins[meta.Name] = p
	pm.metadata[meta.Name] = meta
	return nil
}

// Unregister removes a plugin from the manager.
func (pm *PluginManager) Unregister(name string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	p, exists := pm.plugins[name]
	if !exists {
		return ErrPluginNotFound
	}

	if err := p.Close(); err != nil {
		return err
	}

	delete(pm.plugins, name)
	delete(pm.metadata, name)
	return nil
}

// Get retrieves a plugin by name.
func (pm *PluginManager) Get(name string) (Plugin, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	p, exists := pm.plugins[name]
	if !exists {
		return nil, ErrPluginNotFound
	}
	return p, nil
}

// List returns all registered plugins.
func (pm *PluginManager) List() []Plugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]Plugin, 0, len(pm.plugins))
	for _, p := range pm.plugins {
		result = append(result, p)
	}
	return result
}

// ListMetadata returns all registered plugin metadata.
func (pm *PluginManager) ListMetadata() []PluginMetadata {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]PluginMetadata, 0, len(pm.metadata))
	for _, meta := range pm.metadata {
		result = append(result, meta)
	}
	return result
}

// Errors for plugin management.
var (
	ErrPluginAlreadyRegistered = errors.New("plugin already registered")
	ErrPluginNotFound          = errors.New("plugin not found")
	ErrInvalidPlugin           = errors.New("invalid plugin: missing name")
)
