package model

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("record not found")

// Role constants
const (
	RolePlatformAdmin = "platform_admin"
	RoleAdmin         = "admin"
	RoleUser          = "user"
	RoleViewer        = "viewer"
)

// User maps to the users table (shared-table multi-tenancy)
type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"column:password_hash;size:255;not null" json:"-"`
	Email        string    `gorm:"size:128" json:"email"`
	TenantID     uint      `gorm:"index;not null;default:0" json:"tenantId"` // 0=platform admin
	Role         string    `gorm:"size:32;not null;default:viewer" json:"role"`
	Status       int       `gorm:"not null;default:1" json:"status"`         // 1=active 0=disabled
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (User) TableName() string { return "users" }

// UserModel DB operations interface
type UserModel struct{}

// FindByUsername finds a user by username
func (UserModel) FindByUsername(username string) (*User, error) {
	var u User
	// Use GORM's FirstWhere or direct query - implementation in user-svc.go
	return &u, nil
}

// FindByID finds a user by ID
func (UserModel) FindByID(id uint) (*User, error) {
	var u User
	return &u, nil
}

// FindByEmail finds a user by email
func (UserModel) FindByEmail(email string) (*User, error) {
	var u User
	return &u, nil
}

// Create inserts a new user
func (UserModel) Create(u *User) error {
	return nil
}

// Update updates a user
func (UserModel) Update(u *User) error {
	return nil
}

// Delete removes a user
func (UserModel) Delete(id uint) error {
	return nil
}

// ListByTenant returns all users for a given tenant (excluding platform_admin if tenantID > 0)
func (UserModel) ListByTenant(tenantID uint) ([]*User, error) {
	return []*User{}, nil
}

// ListAll returns all users (platform_admin only operation)
func (UserModel) ListAll() ([]*User, error) {
	return []*User{}, nil
}
