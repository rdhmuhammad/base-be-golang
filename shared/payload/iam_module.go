package payload

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type DefaultUserClaim struct {
	UserData  UserData     `json:"userData"`
	Issuer    string       `json:"iss,omitempty"`
	IssuedAt  *NumericDate `json:"iat,omitempty"`
	Subject   string       `json:"sub,omitempty"`
	ExpiresAt *NumericDate `json:"exp,omitempty"`
}
type NumericDate jwt.NumericDate

type UserData struct {
	UserId   string         `json:"userId"`
	Lang     string         `json:"lang"`
	Timezone string         `json:"timezone"`
	Tz       *time.Location `json:"tz"`
	Email    string         `json:"email"`
	RoleName string         `json:"roleName"`
}

func (authData *UserData) LoadFromMap(m map[string]interface{}) error {
	data, err := json.Marshal(m)

	if err == nil {
		err = json.Unmarshal(data, authData)
	}
	return err
}

type SessionDataUser struct {
	ID            uint      `json:"id"`
	Code          string    `json:"code"`
	UserReference string    `json:"userReference"`
	RoleName      string    `json:"roleName"`
	TimeZone      string    `json:"timeZone"`
	Lang          string    `json:"lang"`
	PhoneNumber   string    `json:"phoneNumber"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	IsVerified    bool      `json:"isVerified"`
	ProfileImage  string    `json:"profileImage"`
	LastActive    time.Time `json:"lastActive"`
}
