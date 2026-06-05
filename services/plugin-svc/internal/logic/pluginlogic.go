package logic

import (
	"context"
	"fmt"
	"plugin"

	"agentmesh/plugin-svc/plugin"
)

// PluginManager wraps the plugin manager for external use.
var PluginMgr = plugin.NewPluginManager()

// LoadPlugin opens a .so file, looks up the PluginInstance symbol,
// and initializes the plugin.
func LoadPlugin(ctx context.Context, soPath string) error {
	// Open the shared object file.
	p, err := plugin.Open(soPath)
	if err != nil {
		return fmt.Errorf("failed to open plugin file: %w", err)
	}

	// Look up the PluginInstance symbol.
	sym, err := p.Lookup("PluginInstance")
	if err != nil {
		return fmt.Errorf("failed to lookup PluginInstance symbol: %w", err)
	}

	// Type assert the symbol to our Plugin interface.
	pluginInst, ok := sym.(plugin.Plugin)
	if !ok {
		return fmt.Errorf("PluginInstance does not implement plugin.Plugin interface")
	}

	// Initialize the plugin.
	if err := pluginInst.Init(ctx); err != nil {
		return fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Register the plugin with the manager.
	if err := PluginMgr.Register(pluginInst); err != nil {
		return fmt.Errorf("failed to register plugin: %w", err)
	}

	return nil
}

// UnloadPlugin closes and unregisters a plugin by name.
func UnloadPlugin(name string) error {
	if err := PluginMgr.Unregister(name); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}
	return nil
}

// ListPlugins returns all loaded plugins.
func ListPlugins() []plugin.Plugin {
	return PluginMgr.List()
}

// ListPluginsMetadata returns metadata for all loaded plugins.
func ListPluginsMetadata() []plugin.PluginMetadata {
	return PluginMgr.ListMetadata()
}

// GetPlugin returns a plugin by name.
func GetPlugin(name string) (plugin.Plugin, error) {
	return PluginMgr.Get(name)
}

// GetMenuItems returns menu items for a specific plugin.
func GetMenuItems(name string) ([]plugin.MenuItem, error) {
	p, err := PluginMgr.Get(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin: %w", err)
	}
	return p.GetMenuItems(), nil
}

// GetAllMenuItems returns menu items from all loaded plugins.
func GetAllMenuItems() []plugin.MenuItem {
	plugins := PluginMgr.List()
	var allItems []plugin.MenuItem
	for _, p := range plugins {
		items := p.GetMenuItems()
		allItems = append(allItems, items...)
	}
	return allItems
}
