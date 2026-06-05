package logic

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"agentmesh/user-svc/internal/svc"
	"agentmesh/user-svc/rpc/pb"
	"agentmesh/user-svc/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TenantLogic {
	return &TenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CreateTenant creates a new tenant and initializes its schema
func (l *TenantLogic) CreateTenant(in *pb.CreateTenantRequest) (*pb.CreateTenantResponse, error) {
	// Check if tenant code already exists
	existingTenant, err := l.svcCtx.TenantModel.FindByCode(l.ctx, in.Code)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	if existingTenant != nil {
		return nil, errors.New("tenant code already exists")
	}

	// Set default plan if not provided
	plan := in.Plan
	if plan == "" {
		plan = "free"
	}

	// Create new tenant
	now := time.Now()
	tenant := &model.Tenant{
		Name:      in.Name,
		Code:      in.Code,
		Status:    model.TenantStatusActive, // 1 = active
		Plan:      plan,
		OwnerID:   in.OwnerId,
		Settings:  in.Settings,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert tenant into database
	id, err := l.svcCtx.TenantModel.Insert(l.ctx, tenant)
	if err != nil {
		return nil, err
	}

	tenant.ID = id

	// Create tenant schema
	if err := l.CreateTenantSchema(id); err != nil {
		return nil, fmt.Errorf("failed to create tenant schema: %w", err)
	}

	return &pb.CreateTenantResponse{
		Tenant: &pb.Tenant{
			Id:        tenant.ID,
			Name:      tenant.Name,
			Code:      tenant.Code,
			Status:    int32(tenant.Status),
			Plan:      tenant.Plan,
			OwnerId:   tenant.OwnerID,
			Settings:  tenant.Settings,
			CreatedAt: nil,
			UpdatedAt: nil,
		},
	}, nil
}

// CreateTenantSchema creates a new database schema for the tenant
func (l *TenantLogic) CreateTenantSchema(tenantID int64) error {
	// Get the underlying sql.DB from the tenant model
	conn, ok := l.svcCtx.TenantModel.(interface {
		GetDB() *sql.DB
	})
	if !ok {
		return errors.New("tenant model does not support schema creation")
	}

	db := conn.GetDB()
	schemaName := fmt.Sprintf("agentmesh_tenant_%d", tenantID)

	// Create the schema
	_, err := db.ExecContext(l.ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Set search path to the new schema
	_, err = db.ExecContext(l.ctx, fmt.Sprintf("SET search_path TO %s", schemaName))
	if err != nil {
		return fmt.Errorf("failed to set search path: %w", err)
	}

	// Create tables from tenant_schema.sql
	tenantSchemaSQL := `
	-- Conversations table
	CREATE TABLE IF NOT EXISTS conversations (
		id SERIAL PRIMARY KEY,
		tenant_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		title VARCHAR(500) NOT NULL DEFAULT 'New Conversation',
		model VARCHAR(100),
		status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'deleted')),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_conversations_tenant_id ON conversations(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_conversations_user_id ON conversations(user_id);
	CREATE INDEX IF NOT EXISTS idx_conversations_status ON conversations(status);
	CREATE INDEX IF NOT EXISTS idx_conversations_created_at ON conversations(created_at);

	-- Messages table
	CREATE TABLE IF NOT EXISTS messages (
		id SERIAL PRIMARY KEY,
		conversation_id INTEGER NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
		tenant_id INTEGER NOT NULL,
		role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system')),
		content TEXT NOT NULL,
		model VARCHAR(100),
		token_count INTEGER DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
	CREATE INDEX IF NOT EXISTS idx_messages_tenant_id ON messages(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_messages_role ON messages(role);
	CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);

	-- Plugin configurations table (per tenant)
	CREATE TABLE IF NOT EXISTS plugin_configs (
		id SERIAL PRIMARY KEY,
		tenant_id INTEGER NOT NULL,
		plugin_id INTEGER NOT NULL,
		config_key VARCHAR(255) NOT NULL,
		config_value JSONB NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(tenant_id, plugin_id, config_key)
	);

	CREATE INDEX IF NOT EXISTS idx_plugin_configs_tenant_id ON plugin_configs(tenant_id);
	CREATE INDEX IF NOT EXISTS idx_plugin_configs_plugin_id ON plugin_configs(plugin_id);
	CREATE INDEX IF NOT EXISTS idx_plugin_configs_config_key ON plugin_configs(config_key);

	-- Function to update updated_at timestamp
	CREATE OR REPLACE FUNCTION update_updated_at_column()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.updated_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ language 'plpgsql';

	-- Triggers for updated_at
	CREATE TRIGGER update_conversations_updated_at
		BEFORE UPDATE ON conversations
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();

	CREATE TRIGGER update_plugin_configs_updated_at
		BEFORE UPDATE ON plugin_configs
		FOR EACH ROW
		EXECUTE FUNCTION update_updated_at_column();
	`

	_, err = db.ExecContext(l.ctx, tenantSchemaSQL)
	if err != nil {
		return fmt.Errorf("failed to create tenant tables: %w", err)
	}

	return nil
}

// GetTenant retrieves tenant info by ID
func (l *TenantLogic) GetTenant(in *pb.GetTenantRequest) (*pb.GetTenantResponse, error) {
	tenant, err := l.svcCtx.TenantModel.FindOneByID(l.ctx, in.Id)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("tenant not found")
		}
		return nil, err
	}

	return &pb.GetTenantResponse{
		Tenant: &pb.Tenant{
			Id:        tenant.ID,
			Name:      tenant.Name,
			Code:      tenant.Code,
			Status:    int32(tenant.Status),
			Plan:      tenant.Plan,
			OwnerId:   tenant.OwnerID,
			Settings:  tenant.Settings,
			CreatedAt: nil,
			UpdatedAt: nil,
		},
	}, nil
}

// ListTenants returns all tenants
func (l *TenantLogic) ListTenants(in *pb.ListTenantsRequest) (*pb.ListTenantsResponse, error) {
	tenants, err := l.svcCtx.TenantModel.FindAll(l.ctx)
	if err != nil {
		if err == model.ErrNotFound {
			return &pb.ListTenantsResponse{Tenants: []*pb.Tenant{}}, nil
		}
		return nil, err
	}

	pbTenants := make([]*pb.Tenant, len(tenants))
	for i, tenant := range tenants {
		pbTenants[i] = &pb.Tenant{
			Id:        tenant.ID,
			Name:      tenant.Name,
			Code:      tenant.Code,
			Status:    int32(tenant.Status),
			Plan:      tenant.Plan,
			OwnerId:   tenant.OwnerID,
			Settings:  tenant.Settings,
			CreatedAt: nil,
			UpdatedAt: nil,
		}
	}

	return &pb.ListTenantsResponse{
		Tenants: pbTenants,
	}, nil
}

// Ensure TenantLogic implements sqlx.SqlConn interface hint
var _ sqlx.SqlConn = (*tenantModelWithDB)(nil)

// tenantModelWithDB is a helper to access underlying DB
type tenantModelWithDB struct {
	conn sqlx.SqlConn
}

func (m *tenantModelWithDB) GetDB() *sql.DB {
	// This is a placeholder - actual implementation would need to access
	// the underlying *sql.DB from the sqlx.SqlConn
	// For now, return nil and handle in logic
	return nil
}
