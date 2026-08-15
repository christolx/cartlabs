package identity

import "time"

type Role string

const (
	RoleBuyer  Role = "buyer"
	RoleSeller Role = "seller"
	RoleAdmin  Role = "admin"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"displayName"`
	Role         Role      `json:"role"`
	Status       string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type AdminUser struct {
	ID          string
	Email       string
	DisplayName string
	Role        Role
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RefreshSession struct {
	ID        string
	FamilyID  string
	UserID    string
	TokenHash []byte
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Session struct {
	AccessToken string `json:"accessToken"`
	TokenType   string `json:"tokenType"`
	ExpiresIn   int    `json:"expiresIn"`
	User        User   `json:"user"`
}

type Principal struct {
	UserID      string
	Role        Role
	UserVersion int64
}

func (p Principal) Require(roles ...Role) bool {
	for _, role := range roles {
		if p.Role == role {
			return true
		}
	}
	return false
}
