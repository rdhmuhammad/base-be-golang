package base

import (
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/logger"
	"base-be-golang/shared/payload"
	"context"

	"github.com/gin-gonic/gin"
	md "github.com/rdhmuhammad/base-be-golang/iam-module/pkg/middleware"

	"gorm.io/gorm"
)

func NewAuth(dbConn *gorm.DB, dbCache cache.DbClient) Security {
	return md.NewAuth(dbConn, dbCache)
}

// ======================= EMPTY AUTH ======================

// EmptyAuth implement if iam module is not used
type EmptyAuth struct {
}

func (e EmptyAuth) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("using empty auth")
	}
}

func (e EmptyAuth) GetUserContext(ctx context.Context) payload.UserData {
	return payload.UserData{}
}

func (e EmptyAuth) Authorize(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("using empty auth")
	}
}

func (e EmptyAuth) SetSession(ctx context.Context, user payload.SessionDataUser) error {
	return nil
}

func (e EmptyAuth) GetSession(ctx context.Context, authCode string, sessionData *payload.SessionDataUser) error {
	return nil
}

func (e EmptyAuth) GetSessionLogin(ctx context.Context, sessionData *payload.SessionDataUser) error {
	return nil
}
