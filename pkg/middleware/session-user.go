package middleware

type SessionDataUser struct {
	BranchID            uint   `json:"branchId"`
	RoleName            string `json:"roleName"`
	TimeZone            string `json:"timeZone"`
	OwnerValidBranchIds string `json:"ownerValidBranchIds"`
}
