package model

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a user is not found
var ErrNotFound = errors.New("not found")

// User represents a user in the system
type User struct {
	ID        int64     `json:"id"`
	TenantID  int64     `json:"tenant_id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Status    int       `json:"status"` // 0: inactive, 1: active, 2: suspended
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"` // unique tenant code
	Status    int       `json:"status"` // 0: inactive, 1: active, 2: suspended
	Plan      string    `json:"plan"` // free, pro, enterprise
	OwnerID   int64     `json:"owner_id"`
	Settings  string    `json:"settings"` // JSON settings
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserStatus constants
const (
	UserStatusInactive  = 0
	UserStatusActive    = 1
	UserStatusSuspended = 2
)

// TenantStatus constants
const (
	TenantStatusInactive  = 0
	TenantStatusActive    = 1
	TenantStatusSuspended = 2
)

// UserModel interface for user database operations
type UserModel interface {
	Insert(ctx context.Context, user *User) (int64, error)
	FindOne(ctx context.Context, id int64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
}

// TenantModel interface for tenant database operations
type TenantModel interface {
	Insert(ctx context.Context, tenant *Tenant) (int64, error)
	FindOne(ctx context.Context, id int64) (*Tenant, error)
	FindByCode(ctx context.Context, code string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id int64) error
}

// UserModelFunc type for UserModel implementation
type UserModelFunc struct{}

// TenantModelFunc type for TenantModel implementation
type TenantModelFunc struct{}
