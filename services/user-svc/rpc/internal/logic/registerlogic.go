package logic

import (
	"context"
	"errors"
	"time"

	"agentmesh/user-svc/internal/svc"
	"agentmesh/user-svc/rpc/pb"
	"agentmesh/user-svc/model"

	"golang.org/x/crypto/bcrypt"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(in *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	// Check if username already exists
	existingUser, err := l.svcCtx.UserModel.FindByUsername(l.ctx, in.Username)
	if err != nil && err != model.ErrNotFound {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists (if provided)
	if in.Email != "" {
		existingEmail, err := l.svcCtx.UserModel.FindByEmail(l.ctx, in.Email)
		if err != nil && err != model.ErrNotFound {
			return nil, err
		}
		if existingEmail != nil {
			return nil, errors.New("email already exists")
		}
	}

	// Hash password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Set default role if not provided
	role := in.Role
	if role == "" {
		role = "user"
	}

	// Set default tenant_id if not provided
	tenantID := in.TenantId
	if tenantID == 0 {
		tenantID = 1 // Default tenant
	}

	// Create new user
	now := time.Now()
	user := &model.User{
		TenantID:  tenantID,
		Username:  in.Username,
		Password:  string(hashedPassword),
		Email:     in.Email,
		Nickname:  in.Nickname,
		Avatar:    "",
		Status:    model.UserStatusActive, // 1 = active
		Role:      role,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Insert user into database
	id, err := l.svcCtx.UserModel.Insert(l.ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = id

	return &pb.RegisterResponse{
		User: &pb.User{
			Id:        user.ID,
			TenantId:  user.TenantID,
			Username:  user.Username,
			Email:     user.Email,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Status:    int32(user.Status),
			Role:      user.Role,
			CreatedAt: nil,
			UpdatedAt: nil,
		},
	}, nil
}
