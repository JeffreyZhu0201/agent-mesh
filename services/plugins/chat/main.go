package main

import (
	"context"

	"agentmesh/plugin-svc/plugin"
)

// ChatPlugin implements the Plugin interface.
type ChatPlugin struct{}

// Metadata returns the plugin metadata.
func (c *ChatPlugin) Metadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:        "chat",
		Version:     "1.0.0",
		Description: "Chat plugin for real-time messaging",
		Type:        "full",
		Author:      "AgentMesh Team",
	}
}

// Init initializes the chat plugin.
func (c *ChatPlugin) Init(ctx context.Context) error {
	// Initialize chat plugin resources.
	return nil
}

// GetMenuItems returns the menu items for the chat plugin.
func (c *ChatPlugin) GetMenuItems() []plugin.MenuItem {
	return []plugin.MenuItem{
		{
			Key:    "chat",
			Label:  "Chat",
			Icon:   "message-circle",
			Path:   "/chat",
			Parent: "",
			Order:  10,
		},
	}
}

// Close releases resources held by the chat plugin.
func (c *ChatPlugin) Close() error {
	return nil
}

// PluginInstance is the plugin entry point that will be looked up.
var PluginInstance plugin.Plugin = &ChatPlugin{}
