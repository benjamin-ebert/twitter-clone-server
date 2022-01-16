package domain

import "time"

type OAuth struct {
	// UserID is the ID of the user that the oauth record belongs to.
	ID int `json:"id"`
	UserID int `json:"user_id" gorm:"notNull;uniqueIndex:user_id_provider"`
	User User `json:"user"`

	// Provider is the name of the oauth provider - Google, Facebook, Github etc.
	// UserID and Provider share the same database index to ensure the pair is unique,
	// since every user should be able to use every provider, but only have one
	// associated oauth record per provider.
	// ProviderUserID is the ID of the user in the provider's database.
	Provider string `json:"provider" gorm:"notNull;uniqueIndex:user_id_provider"`
	ProviderUserID string `json:"provider_user_id"`

	// OAuth token fields. If a provider issues an access token that never expires,
	// Expiry will be a nonsense date and RefreshToken will be empty.
	// They are still in here to be able to support providers who do use them.
	AccessToken string `json:"access_token"`
	TokenType string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Expiry time.Time `json:"expiry"`

	// Timestamps. No DeletedAt, because all OAuth deletions will be hard-deletions.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OAuthService interface {
	Find(userId int, provider string) (*OAuth, error)
	ByProviderUserId(provider, providerUserId string) (*OAuth, error)
	Create(oauth *OAuth) error
	Update(oauth *OAuth) error
}
