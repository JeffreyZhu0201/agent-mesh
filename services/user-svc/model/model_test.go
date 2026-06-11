package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleConstants(t *testing.T) {
	require.Equal(t, "platform_admin", RolePlatformAdmin)
	require.Equal(t, "admin", RoleAdmin)
	require.Equal(t, "user", RoleUser)
	require.Equal(t, "viewer", RoleViewer)
}

func TestUserTableName(t *testing.T) {
	require.Equal(t, "users", User{}.TableName())
}

func TestTenantTableName(t *testing.T) {
	require.Equal(t, "tenants", Tenant{}.TableName())
}

func TestErrNotFound(t *testing.T) {
	require.Error(t, ErrNotFound)
	require.Equal(t, "record not found", ErrNotFound.Error())
}
