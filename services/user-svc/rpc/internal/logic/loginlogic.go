package logic

import (
	"context"
	"errors"
	"time"

	"agentmesh/user-svc/internal/svc"
	"agentmesh/user-svc/rpc/pb"
	"agentmesh/user-svc/model"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// JWT claims for user authentication
type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	TenantID int64  `json:"tenant_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (l *LoginLogic) Login(in *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Find user by username
	user, err := l.svcCtx.UserModel.FindByUsername(l.ctx, in.Username)
	if err != nil {
		if err == model.ErrNotFound {
			return nil, errors.New("invalid username or password")
		}
		return nil, err
	}

	// Compare password with bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Check user status
	if user.Status != 1 { // 1 = active
		return nil, errors.New("user account is not active")
	}

	// Generate JWT token
	expiryDuration := time.Duration(l.svcCtx.Config.JWT.Expiry) * time.Second
	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID,
		TenantID: user.TenantID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiryDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "agentmesh",
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(l.svcCtx.Config.JWT.Secret))
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		Token: tokenString,
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
