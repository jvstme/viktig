package entities

type VkUser struct {
	VkUserId         int    `json:"id"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	CanAccessPrivate bool   `json:"can_access_closed"`
	IsPrivateProfile bool   `json:"is_closed"`
}
