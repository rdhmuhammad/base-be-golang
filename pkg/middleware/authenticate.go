package middleware

import (
	"base-be-golang/internal/dto"
	"base-be-golang/internal/localerror"
	"base-be-golang/pkg/cache"
	"base-be-golang/pkg/davinci"
	"base-be-golang/pkg/environment"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"golang.org/x/exp/slices"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Auth struct {
	cache   cache.Cache
	davinci davinci.Generator
	mapper  mapperAuth
	env     environment.Environment
}

func NewAuth() AuthInterface {
	return Auth{
		cache:   cache.Default(),
		davinci: davinci.DefaultDavinci(),
		mapper:  SharedMapper{},
		env:     environment.NewEnvironment(),
	}
}

type mapperAuth interface {
	GetBodyJSON(c *gin.Context) map[string]any
}

func (receiver Auth) SignClaim(claim DefaultUserClaim) (string, error) {
	method := jwt.SigningMethodHS256
	token := &jwt.Token{
		Header: map[string]interface{}{
			"typ": "JWT",
			"alg": method.Alg(),
		},
		Claims: claim,
		Method: method,
	}
	secret := []byte(os.Getenv("SECRET"))
	tokenStr, err := token.SignedString(secret)
	if err != nil {

		return "", err
	}
	return tokenStr, nil
}

func (receiver Auth) GetSessionDataFromContext(c *gin.Context) (SessionDataUser, error) {
	authData := receiver.GetAuthDataFromContext(c)
	return receiver.GetSessionData(strconv.Itoa(int(authData.UserId)))
}

func (receiver Auth) GetSessionData(userId string) (SessionDataUser, error) {
	loginCacheKey := "LOGIN_KEY_"
	secretSession := receiver.env.Get("DEFAULT_SECRET_LOGIN_SESSION")
	sessionKey, err := receiver.davinci.GenerateHashValue(secretSession, userId, 10)
	if err != nil {

		return SessionDataUser{}, err
	}

	var sessionData SessionDataUser
	sessionStr, err := receiver.cache.Get(context.Background(), loginCacheKey+sessionKey)
	if err != nil {

		if errors.Is(redis.Nil, err) {
			return SessionDataUser{}, localerror.InvalidDataError{Msg: ""}
		}
		return SessionDataUser{}, err
	}

	err = json.Unmarshal([]byte(sessionStr), &sessionData)
	if err != nil {

		return SessionDataUser{}, err
	}

	return sessionData, nil
}

func (receiver Auth) SessionCheck(placeOn string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if os.Getenv("BYPASS_SESSION_CHECKING") == "1" {
			c.Next()
			return
		}
		var paramKey = KeyBranchID
		userData := receiver.GetAuthDataFromContext(c)
		sessionData, err := receiver.GetSessionData(strconv.Itoa(int(userData.UserId)))
		if err != nil {
			response := dto.DefaultErrorInvalidDataWithMessage(fmt.Sprintf("sessionCheck: %s", err.Error()))
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}
		var param string
		switch placeOn {
		case RequestParams:
			param = c.Param(paramKey)
		case RequestQuery:
			param = c.Query(paramKey)
		case RequestBodyJSON:
			body := receiver.mapper.GetBodyJSON(c)
			switch v := body[paramKey].(type) {
			case int:
				param = strconv.FormatInt(int64(v), 10)
			case uint:
				param = strconv.FormatUint(uint64(v), 10)
			case float32:
				param = strconv.FormatFloat(float64(v), 'f', -1, 32)
			case float64:
				param = strconv.FormatFloat(float64(v), 'f', -1, 64)
			case string:
				param = v
			}
		}

		paramBranchId, err := strconv.ParseUint(param, 10, 64)
		if err != nil {

			response := dto.DefaultErrorInvalidDataWithMessage(fmt.Sprintf("strconv.ParseUint(param): %s", err.Error()))
			c.JSON(http.StatusInternalServerError, response)
			c.Abort()
			return
		}

		if sessionData.BranchID == uint(paramBranchId) &&
			sessionData.RoleName == RoleUser {
			c.Next()
			return
		}
		response := dto.DefaultErrorInvalidDataWithMessage("Tidak memiliki akses")
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
		return
	}

}

func (receiver Auth) Authorize(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		var authData = UserData{}
		authDataStr, ok := c.Get("authData")
		if ok {
			authDataMap := authDataStr.(map[string]interface{})
			err := authData.LoadFromMap(authDataMap)
			if err != nil {

				c.JSON(http.StatusUnauthorized, "invalid token")
				c.Abort()
				return
			}
		}
		if slices.Contains(roles, authData.RoleName) {
			c.Next()
			return
		}

		response := dto.DefaultBadRequestResponse()
		response.Message = "Kamu tidak punya akses ke halaman ini"
		response.ResponseTime = fmt.Sprint(time.Since(start).Milliseconds(), " ms.")
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
	}
}

func (receiver Auth) Validate() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		tokenStr := strings.Replace(c.GetHeader("Authorization"), "Bearer ", "", -1)
		secret := os.Getenv("SECRET")
		token, err := receiver.parseToken(tokenStr, []byte(secret))
		if err != nil {

			response := dto.DefaultErrorResponseWithMessage(err.Error())
			response.ResponseTime = fmt.Sprint(time.Since(start).Milliseconds(), " ms.")
			c.JSON(http.StatusUnauthorized, response)
			c.Abort()
			return
		}

		authData, valid := receiver.getAuthData(token)

		userDataStruct := UserData{}
		err = userDataStruct.LoadFromMap(authData)
		if err != nil {

			if err != nil {

				response := dto.DefaultErrorResponse()
				response.Message = "Parse JWT payload failed."
				response.ResponseTime = fmt.Sprint(time.Since(start).Milliseconds(), " ms.")
				c.JSON(http.StatusUnauthorized, response)
				c.Abort()
				return
			}
		}
		if valid {
			c.Set("authData", authData)
			c.Next()
			return
		}

		response := dto.DefaultErrorResponse()
		response.ResponseTime = fmt.Sprint(time.Since(start).Milliseconds(), " ms.")
		c.JSON(http.StatusUnauthorized, response)
		c.Abort()
	}
}

func (receiver Auth) parseToken(tokenStr string, secret []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("invalid token format")
		}
		return secret, nil
	})
	if err != nil {

		return nil, err
	}
	return token, nil
}

func (receiver Auth) GetAuthDataFromContext(c *gin.Context) UserData {
	var authData = UserData{}
	authDataStr, ok := c.Get("authData")
	if !ok {
		return UserData{}
	}
	authDataMap := authDataStr.(map[string]interface{})
	err := authData.LoadFromMap(authDataMap)
	if err != nil {

		return UserData{}
	}
	if !ok {
		return UserData{}
	}
	return authData
}

func (receiver Auth) getAuthData(token *jwt.Token) (map[string]interface{}, bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	valid := ok && token.Valid
	if !ok {
		return nil, false
	}

	var authData map[string]interface{}

	if valid {
		authData = claims["userData"].(map[string]interface{})
	}

	return authData, valid
}

func (receiver Auth) GetSessionFromContext(ctx context.Context) SessionDataUser {
	var data SessionDataUser
	if oc, ok := ctx.Value(CtxKeySession).(SessionDataUser); ok {
		data = oc
	}

	return data
}

func (receiver Auth) SetSessionToContext(c *gin.Context, ctx context.Context) (context.Context, error) {
	fromContext, err := receiver.GetSessionDataFromContext(c)
	if err != nil {
		return nil, err
	}

	return context.WithValue(ctx, CtxKeySession, fromContext), nil
}

type AuthInterface interface {
	GetSessionFromContext(ctx context.Context) SessionDataUser
	SetSessionToContext(c *gin.Context, ctx context.Context) (context.Context, error)
	SignClaim(claim DefaultUserClaim) (string, error)
	Validate() gin.HandlerFunc
	Authorize(roles ...string) gin.HandlerFunc
	GetAuthDataFromContext(c *gin.Context) UserData
	GetSessionData(userId string) (SessionDataUser, error)
	GetSessionDataFromContext(c *gin.Context) (SessionDataUser, error)
	SessionCheck(placeOn string) gin.HandlerFunc
}
