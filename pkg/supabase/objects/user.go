package objects

import "time"

type IdentityData struct {
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	PhoneVerified bool   `json:"phone_verified,omitempty"`
	Sub           string `json:"sub,omitempty"`
}

type Identity struct {
	IdentityID   string       `json:"identity_id,omitempty"`
	ID           string       `json:"id,omitempty"`
	UserID       string       `json:"user_id,omitempty"`
	IdentityData IdentityData `json:"identity_data,omitempty"`
	Provider     string       `json:"provider,omitempty"`
	LastSignInAt time.Time    `json:"last_sign_in_at,omitempty"`
	CreatedAt    time.Time    `json:"created_at,omitempty"`
	UpdatedAt    time.Time    `json:"updated_at,omitempty"`
	Email        string       `json:"email,omitempty"`
}

type AppMetadata struct {
	Provider  string   `json:"provider,omitempty"`
	Providers []string `json:"providers,omitempty"`
	Role      []string `json:"role,omitempty"`
}

type User struct {
	ID               string      `json:"id,omitempty"`
	Aud              string      `json:"aud,omitempty"`
	Role             string      `json:"role,omitempty"`
	Email            string      `json:"email,omitempty"`
	EmailConfirmedAt time.Time   `json:"email_confirmed_at,omitempty"`
	Phone            string      `json:"phone,omitempty"`
	ConfirmedAt      time.Time   `json:"confirmed_at,omitempty"`
	LastSignInAt     time.Time   `json:"last_sign_in_at,omitempty"`
	AppMetadata      AppMetadata `json:"app_metadata,omitempty"`
	UserMetadata     interface{} `json:"user_metadata,omitempty"`
	Identities       []Identity  `json:"identities,omitempty"`
	CreatedAt        time.Time   `json:"created_at,omitempty"`
	UpdatedAt        time.Time   `json:"updated_at,omitempty"`
	IsAnonymous      bool        `json:"is_anonymous,omitempty"`
}
