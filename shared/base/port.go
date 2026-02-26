package base

import (
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/davinci"
	"base-be-golang/pkg/mailing"
	"base-be-golang/pkg/mapper"
	"base-be-golang/pkg/middleware"
	"base-be-golang/shared/payload"
	"bytes"
	"context"
	"html/template"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	md "github.com/rdhmuhammad/base-be-golang/iam-module/pkg/middleware"
	"gorm.io/gorm"
)

type Port struct {
	Security   Security
	ErrHandler ErrHandler
	Cache      Cache
	Env        Environment
	Davinci    Generator
	Mailing    Mailing
	Clock      Clock
}

type Security interface {
	Validate() gin.HandlerFunc
	GetUserContext(ctx context.Context) payload.UserData
	Authorize(roles ...string) gin.HandlerFunc
	SetSession(ctx context.Context, user payload.SessionDataUser) error
	GetSession(ctx context.Context, authCode string, sessionData *payload.SessionDataUser) error
	GetSessionLogin(ctx context.Context, sessionData *payload.SessionDataUser) error
}

type StorageService interface {
	GetFile(ctx context.Context, fileName string) (*bytes.Buffer, error)
	StoreFile(ctx context.Context, fileName string, file io.Reader, fileSize int64) (string, error)
	DeleteFile(ctx context.Context, fileName string) error
	HealthCheck(ctx context.Context) error
}

type ErrHandler interface {
	ErrorPrint(err error)
	DebugPrint(err string, v ...interface{})
	ErrorReturn(err error) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, keys ...string) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

type Mailing interface {
	NativeSendEmail(payload mailing.NativeSendEmailPayload) error
}

type Generator interface {
	GenerateUniqueKeyWithPredicate(
		secretKey string,
		uniqueID string,
		length int,
		isUnique davinci.UniquePredicate,
	) (string, error)
	GenerateUniqueKey(
		secretKey []byte,
		uniqueID string,
		length int,
	) (string, error)
	GenerateHash(secretKey []byte, uniqueID string) (string, error)
	GenerateHashValue(session string, id string, i int) (string, error)
	DecryptMessage(key []byte, data string) (string, error)
	EncryptMessage(key, data []byte) (string, error)
	GenerateOTPCode(
		secret string,
		counter uint64,
	) (int, error)
}

type Environment interface {
	CheckFlag(flag string) bool
	Get(key string) string
	GetInt(key string, defaultValue int) int
	GetUint(key string, defaultValue uint) uint
	GetFloat(key string, defaultValue float64) float64
	GetBranchID() uint
}

type Clock interface {
	ParseWithTzFromCtx(ctx context.Context, value string, format string) time.Time
	Now(ctx context.Context) time.Time
	NowUTC() time.Time
	NowUnix() int64
	GetTimeZoneByName(name string) *time.Location
	SetTimezoneToContext(ctx context.Context, val string) context.Context
	GetTimezoneFromContext(ctx context.Context) *time.Location
}

type SendInBlueInterface interface {
	NativeSendEmail(ctx context.Context, payload mailing.NativeSendEmailPayload) error
}

func (uc Port) GenerateEmailBodyVerifyOTP(
	ctx context.Context,
	payload payload.EmailBodyVerifyOTPPayload,
) (string, error) {
	htmlPath := "./resource/mailing/verification-email.html"
	tmpl := template.Must(template.ParseFiles(htmlPath))
	outWriter := bytes.Buffer{}

	err := tmpl.Execute(&outWriter, payload)
	if err != nil {
		return "", err
	}

	return outWriter.String(), nil
}

// ======================== BASE CONTROLLER ====================

type BaseController struct {
	Mapper   Mapper
	Enigma   Validator
	Security Security
	Idem     Idempotent
}

func NewBaseController(db *gorm.DB, dbCache cache.DbClient) BaseController {
	return BaseController{
		Mapper:   mapper.NewMapper(),
		Enigma:   middleware.NewEnigma(),
		Security: md.NewAuth(db, dbCache),
	}
}

type Mapper interface {
	NewResponse(c *gin.Context, res *payload.Response, err error)
}

type Validator interface {
	BindQueryToFilterAndValidate(c *gin.Context, payload interface{}) map[string][]string
	BindAndValidate(c *gin.Context, payload any) map[string][]string
	BindQueryToFilter(c *gin.Context, payload interface{}) error
}

type Idempotent interface {
	Idempotent(name string, paramKey string, lockTime time.Duration) gin.HandlerFunc
}
