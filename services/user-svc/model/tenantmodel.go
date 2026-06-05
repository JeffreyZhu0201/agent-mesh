package model

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// sqlTenantModel implements TenantModel interface using sqlx
type sqlTenantModel struct {
	conn sqlx.SqlConn
}

// NewTenantModel creates a new sqlTenantModel with the given sqlx.SqlConn
func NewTenantModel(conn sqlx.SqlConn) TenantModel {
	return &sqlTenantModel{
		conn: conn,
	}
}

// Insert inserts a new tenant into the database
func (m *sqlTenantModel) Insert(ctx context.Context, tenant *Tenant) (int64, error) {
	query := `INSERT INTO agentmesh_public.tenants (name, code, status, plan, owner_id, settings, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.conn.ExecCtx(ctx, query,
		tenant.Name,
		tenant.Code,
		tenant.Status,
		tenant.Plan,
		tenant.OwnerID,
		tenant.Settings,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// FindOneByID finds a tenant by ID
func (m *sqlTenantModel) FindOneByID(ctx context.Context, id int64) (*Tenant, error) {
	query := "SELECT id, name, code, status, plan, owner_id, settings, created_at, updated_at FROM agentmesh_public.tenants WHERE id = ?"

	var tenant Tenant
	err := m.conn.QueryRowCtx(ctx, &tenant, query, id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// FindAll returns all tenants
func (m *sqlTenantModel) FindAll(ctx context.Context) ([]*Tenant, error) {
	query := "SELECT id, name, code, status, plan, owner_id, settings, created_at, updated_at FROM agentmesh_public.tenants ORDER BY id ASC"

	var tenants []*Tenant
	err := m.conn.QueryRowsCtx(ctx, &tenants, query)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return tenants, nil
}

// FindByCode finds a tenant by code
func (m *sqlTenantModel) FindByCode(ctx context.Context, code string) (*Tenant, error) {
	query := "SELECT id, name, code, status, plan, owner_id, settings, created_at, updated_at FROM agentmesh_public.tenants WHERE code = ?"

	var tenant Tenant
	err := m.conn.QueryRowCtx(ctx, &tenant, query, code)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &tenant, nil
}

// Update updates a tenant in the database
func (m *sqlTenantModel) Update(ctx context.Context, tenant *Tenant) error {
	query := `UPDATE agentmesh_public.tenants SET name = ?, code = ?, status = ?, plan = ?, owner_id = ?, settings = ?, updated_at = ? WHERE id = ?`

	_, err := m.conn.ExecCtx(ctx, query,
		tenant.Name,
		tenant.Code,
		tenant.Status,
		tenant.Plan,
		tenant.OwnerID,
		tenant.Settings,
		time.Now(),
		tenant.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes a tenant from the database
func (m *sqlTenantModel) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM agentmesh_public.tenants WHERE id = ?"

	_, err := m.conn.ExecCtx(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// Ensure sqlTenantModel implements TenantModel
var _ TenantModel = (*sqlTenantModel)(nil)
