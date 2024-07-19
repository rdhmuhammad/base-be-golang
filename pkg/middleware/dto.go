package middleware

import (
	"encoding/json"

	"github.com/golang-jwt/jwt/v4"
)

type DefaultUserClaim struct {
	UserData UserData `json:"userData"`
	jwt.RegisteredClaims
}

type TestStruct struct {
	Nama string `json:"nama" validate:"required"`
}

type UserData struct {
	Username string `json:"username"`
	UserId   uint   `json:"userId"`
	BranchID uint   `json:"branchId"`
	Email    string `json:"email"`
	RoleName string `json:"roleName"`
}

func (authData *UserData) LoadFromMap(m map[string]interface{}) error {
	data, err := json.Marshal(m)
	if err == nil {
		err = json.Unmarshal(data, authData)
	}
	return err
}
