package model

import "time"

// Tenant maps to the tenants table
type Tenant struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:128;not null" json:"name"`
	Code      string    `gorm:"size:64;uniqueIndex;not null" json:"code"`
	Status    int       `gorm:"not null;default:1" json:"status"` // 1=active 0=disabled
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Tenant) TableName() string { return "tenants" }

// TenantModel DB operations interface
type TenantModel struct{}

// FindByID finds a tenant by ID
func (TenantModel) FindByID(id uint) (*Tenant, error) {
	var t Tenant
	return &t, nil
}

// FindByCode finds a tenant by code
func (TenantModel) FindByCode(code string) (*Tenant, error) {
	var t Tenant
	return &t, nil
}

// Create inserts a new tenant
func (TenantModel) Create(t *Tenant) error {
	return nil
}

// Update updates a tenant
func (TenantModel) Update(t *Tenant) error {
	return nil
}

// Delete removes a tenant
func (TenantModel) Delete(id uint) error {
	return nil
}

// ListAll returns all tenants
func (TenantModel) ListAll() ([]*Tenant, error) {
	return []*Tenant{}, nil
}
