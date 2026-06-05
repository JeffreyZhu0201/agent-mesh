package model

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// sqlUserModel implements UserModel interface using sqlx
type sqlUserModel struct {
	conn sqlx.SqlConn
}

// NewUserModel creates a new sqlUserModel with the given sqlx.SqlConn
func NewUserModel(conn sqlx.SqlConn) UserModel {
	return &sqlUserModel{
		conn: conn,
	}
}

// Insert inserts a new user into the database
func (m *sqlUserModel) Insert(ctx context.Context, user *User) (int64, error) {
	query := `INSERT INTO users (tenant_id, username, email, password, nickname, avatar, status, role, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.conn.ExecCtx(ctx, query,
		user.TenantID,
		user.Username,
		user.Email,
		user.Password,
		user.Nickname,
		user.Avatar,
		user.Status,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
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

// FindOne finds a user by ID
func (m *sqlUserModel) FindOne(ctx context.Context, id int64) (*User, error) {
	query := "SELECT id, tenant_id, username, email, password, nickname, avatar, status, role, created_at, updated_at FROM users WHERE id = ?"

	var user User
	err := m.conn.QueryRowCtx(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// FindByUsername finds a user by username
func (m *sqlUserModel) FindByUsername(ctx context.Context, username string) (*User, error) {
	query := "SELECT id, tenant_id, username, email, password, nickname, avatar, status, role, created_at, updated_at FROM users WHERE username = ?"

	var user User
	err := m.conn.QueryRowCtx(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// FindByEmail finds a user by email
func (m *sqlUserModel) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := "SELECT id, tenant_id, username, email, password, nickname, avatar, status, role, created_at, updated_at FROM users WHERE email = ?"

	var user User
	err := m.conn.QueryRowCtx(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Update updates a user in the database
func (m *sqlUserModel) Update(ctx context.Context, user *User) error {
	query := `UPDATE users SET email = ?, nickname = ?, avatar = ?, status = ?, role = ?, updated_at = ? WHERE id = ?`

	_, err := m.conn.ExecCtx(ctx, query,
		user.Email,
		user.Nickname,
		user.Avatar,
		user.Status,
		user.Role,
		time.Now(),
		user.ID,
	)
	if err != nil {
		return err
	}

	return nil
}

// Delete deletes a user from the database
func (m *sqlUserModel) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = ?"

	_, err := m.conn.ExecCtx(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// Ensure sqlUserModel implements UserModel
var _ UserModel = (*sqlUserModel)(nil)
