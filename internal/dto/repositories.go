package dto

type UserCheckPhoneAndEmail struct {
	Exist bool   `json:"isExist"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}
