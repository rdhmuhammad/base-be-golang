package middleware

import (
	"base-be-golang/internal/dto"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/exp/slices"
	"net/http"
	"os"
	"strings"
	"time"
)

type Auth struct {
}

type AuthInterface interface {
	Authorize(roles ...string) gin.HandlerFunc
	Validate() gin.HandlerFunc
	GetAuthDataFromContext(c *gin.Context) UserData
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

		response := dto.DefaultErrorResponse()
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
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
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
