package models

import(
	"github.com/google/uuid"
	"time"
	"encoding/json"
)

type Users struct {
	InstanceId *uuid.UUID `json:"instance_id,omitempty" column:"instance_id"`
	Id uuid.UUID `json:"id,omitempty" column:"id"`
	Aud *string `json:"aud,omitempty" column:"aud"`
	Role *string `json:"role,omitempty" column:"role"`
	Email *string `json:"email,omitempty" column:"email"`
	EncryptedPassword *string `json:"encrypted_password,omitempty" column:"encrypted_password"`
	EmailConfirmedAt *time.Time `json:"email_confirmed_at,omitempty" column:"email_confirmed_at"`
	InvitedAt *time.Time `json:"invited_at,omitempty" column:"invited_at"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" column:"confirmation_token"`
	ConfirmationSentAt *time.Time `json:"confirmation_sent_at,omitempty" column:"confirmation_sent_at"`
	RecoveryToken *string `json:"recovery_token,omitempty" column:"recovery_token"`
	RecoverySentAt *time.Time `json:"recovery_sent_at,omitempty" column:"recovery_sent_at"`
	EmailChangeTokenNew *string `json:"email_change_token_new,omitempty" column:"email_change_token_new"`
	EmailChange *string `json:"email_change,omitempty" column:"email_change"`
	EmailChangeSentAt *time.Time `json:"email_change_sent_at,omitempty" column:"email_change_sent_at"`
	LastSignInAt *time.Time `json:"last_sign_in_at,omitempty" column:"last_sign_in_at"`
	RawAppMetaData *json.RawMessage `json:"raw_app_meta_data,omitempty" column:"raw_app_meta_data"`
	RawUserMetaData *json.RawMessage `json:"raw_user_meta_data,omitempty" column:"raw_user_meta_data"`
	IsSuperAdmin *bool `json:"is_super_admin,omitempty" column:"is_super_admin"`
	CreatedAt *time.Time `json:"created_at,omitempty" column:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" column:"updated_at"`
	Phone *string `json:"phone,omitempty" column:"phone"`
	PhoneConfirmedAt *time.Time `json:"phone_confirmed_at,omitempty" column:"phone_confirmed_at"`
	PhoneChange *string `json:"phone_change,omitempty" column:"phone_change"`
	PhoneChangeToken *string `json:"phone_change_token,omitempty" column:"phone_change_token"`
	PhoneChangeSentAt *time.Time `json:"phone_change_sent_at,omitempty" column:"phone_change_sent_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty" column:"confirmed_at"`
	EmailChangeTokenCurrent *string `json:"email_change_token_current,omitempty" column:"email_change_token_current"`
	EmailChangeConfirmStatus *int16 `json:"email_change_confirm_status,omitempty" column:"email_change_confirm_status"`
	BannedUntil *time.Time `json:"banned_until,omitempty" column:"banned_until"`
	ReauthenticationToken *string `json:"reauthentication_token,omitempty" column:"reauthentication_token"`
	ReauthenticationSentAt *time.Time `json:"reauthentication_sent_at,omitempty" column:"reauthentication_sent_at"`
	IsSsoUser bool `json:"is_sso_user,omitempty" column:"is_sso_user"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" column:"deleted_at"`

	Metadata string `schema:"auth"`
	Acl string `read:"" write:""`
}
