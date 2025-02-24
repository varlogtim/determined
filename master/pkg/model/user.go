package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/guregu/null.v3"

	"github.com/pkg/errors"
	"github.com/uptrace/bun"

	"github.com/determined-ai/determined/proto/pkg/userv1"
)

var (
	// EmptyPassword is the empty password (i.e., the empty string).
	EmptyPassword = null.NewString("", false)

	// NoPasswordLogin is a password that prevents the user from logging in
	// directly. They can still login via external authentication methods like
	// OAuth.
	NoPasswordLogin = null.NewString("", true)
)

// BCryptCost is a stopgap until we implement sane master-configuration.
const BCryptCost = 15

// UserID is the type for user IDs.
type UserID int

// SessionID is the type for user session IDs.
type SessionID int

// User corresponds to a row in the "users" DB table.
type User struct {
	bun.BaseModel `bun:"table:users"`
	ID            UserID      `db:"id" bun:"id,pk,autoincrement" json:"id"`
	Username      string      `db:"username" json:"username"`
	PasswordHash  null.String `db:"password_hash" json:"-"`
	DisplayName   null.String `db:"display_name" json:"display_name"`
	Admin         bool        `db:"admin" json:"admin"`
	Active        bool        `db:"active" json:"active"`
	ModifiedAt    time.Time   `db:"modified_at" json:"modified_at"`
	Remote        bool        `db:"remote" json:"remote"`
	LastAuthAt    *time.Time  `db:"last_auth_at" json:"last_auth_at"`
}

// UserSession corresponds to a row in the "user_sessions" DB table.
type UserSession struct {
	bun.BaseModel   `bun:"table:user_sessions"`
	ID              SessionID         `db:"id" json:"id"`
	UserID          UserID            `db:"user_id" json:"user_id"`
	Expiry          time.Time         `db:"expiry" json:"expiry"`
	InheritedClaims map[string]string `bun:"-"` // InheritedClaims contains the OIDC raw ID token when OIDC is enabled
}

// A FullUser is a User joined with any other user relations.
type FullUser struct {
	ID          UserID      `db:"id" json:"id"`
	DisplayName null.String `db:"display_name" json:"display_name"`
	Username    string      `db:"username" json:"username"`
	Name        string      `db:"name" json:"name"`
	Admin       bool        `db:"admin" json:"admin"`
	Active      bool        `db:"active" json:"active"`
	ModifiedAt  time.Time   `db:"modified_at" json:"modified_at"`
	Remote      bool        `db:"remote" json:"remote"`
	LastAuthAt  *time.Time  `db:"last_auth_at" json:"last_auth_at"`

	AgentUID   null.Int    `db:"agent_uid" json:"agent_uid"`
	AgentGID   null.Int    `db:"agent_gid" json:"agent_gid"`
	AgentUser  null.String `db:"agent_user" json:"agent_user"`
	AgentGroup null.String `db:"agent_group" json:"agent_group"`
}

// ToUser converts a FullUser model to just a User model.
func (u FullUser) ToUser() User {
	return User{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: null.String{},
		DisplayName:  u.DisplayName,
		Admin:        u.Admin,
		Active:       u.Active,
		ModifiedAt:   u.ModifiedAt,
		Remote:       u.Remote,
		LastAuthAt:   u.LastAuthAt,
	}
}

// ValidatePassword checks that the supplied password is correct.
func (user User) ValidatePassword(password string) bool {
	// If an empty password was posted, we need to check that the
	// user is a password-less user.
	if password == "" {
		return !user.PasswordHash.Valid
	}

	// If the model's password is empty, then
	// supplied password must be incorrect
	if !user.PasswordHash.Valid {
		return false
	}

	err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash.ValueOrZero()),
		[]byte(password))

	return err == nil
}

// UpdatePasswordHash updates the model's password hash employing necessary cryptographic
// techniques.
func (user *User) UpdatePasswordHash(password string) error {
	if password == "" {
		user.PasswordHash = EmptyPassword
	} else {
		passwordHash, err := HashPassword(password)
		if err != nil {
			return errors.Wrap(err, "error updating user password")
		}

		user.PasswordHash = null.StringFrom(passwordHash)
	}
	return nil
}

// Proto converts a user to its protobuf representation.
func (user *User) Proto() *userv1.User {
	u := &userv1.User{
		Id:          int32(user.ID),
		Username:    user.Username,
		DisplayName: user.DisplayName.ValueOrZero(),
		Admin:       user.Admin,
		Active:      user.Active,
		ModifiedAt:  timestamppb.New(user.ModifiedAt),
		Remote:      user.Remote,
	}
	if user.LastAuthAt != nil {
		u.LastAuthAt = timestamppb.New(*user.LastAuthAt)
	}
	return u
}

// Users is a slice of User objects—primarily useful for its methods.
type Users []User

// Proto converts a slice of users to its protobuf representation.
func (users Users) Proto() []*userv1.User {
	out := make([]*userv1.User, len(users))
	for i, u := range users {
		out[i] = u.Proto()
	}
	return out
}

// ExternalSessions provides an integration point for an external service to issue JWTs to control
// access to the cluster.
type ExternalSessions struct {
	LoginURI  string `json:"login_uri"`
	LogoutURI string `json:"logout_uri"`
	JwtKey    string `json:"jwt_key"`
}

// Enabled returns whether or not external sessions are enabled.
func (e ExternalSessions) Enabled() bool {
	return len(e.LoginURI) > 1
}

// UserWebSetting is a record of user web setting.
type UserWebSetting struct {
	UserID      UserID
	Key         string
	Value       string
	StoragePath string
}

// Proto returns the protobuf representation.
func (s UserWebSetting) Proto() *userv1.UserWebSetting {
	return &userv1.UserWebSetting{
		Key:         s.Key,
		StoragePath: s.StoragePath,
		Value:       s.Value,
	}
}

// HashPassword hashes the user's password.
func HashPassword(password string) (string, error) {
	// truncate password to confirm the bcrypt length limit <=72bytes
	passwordBytes := []byte(password)
	var truncatedBytes []byte
	if len(passwordBytes) <= 72 {
		truncatedBytes = passwordBytes
	} else {
		truncatedBytes = passwordBytes[:72]
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		truncatedBytes,
		BCryptCost,
	)
	if err != nil {
		return "", err
	}
	return string(passwordHash), nil
}
