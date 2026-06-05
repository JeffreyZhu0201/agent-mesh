package svc

import (
	"agentmesh/user-svc/internal/config"
	"agentmesh/user-svc/model"
)

type ServiceContext struct {
	Config *config.Config
	UserModel *model.UserModel
	TenantModel *model.TenantModel
}

func NewServiceContext(c *config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
