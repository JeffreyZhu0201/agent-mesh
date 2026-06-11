package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockPlugin struct {
	name string
}

func (m *mockPlugin) Metadata() PluginMetadata {
	return PluginMetadata{Name: m.name, Version: "1.0.0"}
}

func (m *mockPlugin) Init(ctx context.Context) error { return nil }
func (m *mockPlugin) GetMenuItems() []MenuItem       { return nil }
func (m *mockPlugin) Close() error                   { return nil }

func TestPluginManagerRegisterAndGet(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "chat"}

	require.NoError(t, pm.Register(p))

	got, err := pm.Get("chat")
	require.NoError(t, err)
	require.Equal(t, "chat", got.Metadata().Name)
}

func TestPluginManagerDuplicateRegister(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "chat"}
	require.NoError(t, pm.Register(p))
	require.ErrorIs(t, pm.Register(p), ErrPluginAlreadyRegistered)
}

func TestPluginManagerInvalidPlugin(t *testing.T) {
	pm := NewPluginManager()
	require.ErrorIs(t, pm.Register(&mockPlugin{name: ""}), ErrInvalidPlugin)
}

func TestPluginManagerUnregister(t *testing.T) {
	pm := NewPluginManager()
	p := &mockPlugin{name: "chat"}
	require.NoError(t, pm.Register(p))
	require.NoError(t, pm.Unregister("chat"))
	_, err := pm.Get("chat")
	require.ErrorIs(t, err, ErrPluginNotFound)
}

func TestPluginManagerList(t *testing.T) {
	pm := NewPluginManager()
	require.NoError(t, pm.Register(&mockPlugin{name: "a"}))
	require.NoError(t, pm.Register(&mockPlugin{name: "b"}))
	require.Len(t, pm.List(), 2)
	require.Len(t, pm.ListMetadata(), 2)
}
