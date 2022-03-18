package crud

import (
	"gorm.io/gorm"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// OAuthService manages oauth records.
// It implements the domain.OAuthService interface.
type OAuthService struct {
	oauthValidator
}

// oauthValidator runs validations on incoming OAuth data.
// On success, it passes the data on to oauthGorm.
// Otherwise, it returns the error of the validation that has failed.
type oauthValidator struct {
	oauthGorm
}

// oauthGorm runs CRUD operations on the database using incoming OAuth data.
// It assumes that data has been validated. On success, it returns nil.
// Otherwise, it returns the error of the operation that has failed.
type oauthGorm struct {
	db *gorm.DB
}

// NewOAuthService returns an instance of OAuthService.
func NewOAuthService(db *gorm.DB) *OAuthService {
	return &OAuthService{
		oauthValidator{
			oauthGorm{
				db: db,
			},
		},
	}
}

// Ensure the OAuthService struct properly implements the domain.OAuthService interface.
// If it does not, then this expression becomes invalid and won't compile.
var _ domain.OAuthService = &OAuthService{}

// Create runs validations needed for creating new OAuth database records.
func (ov *oauthValidator) Create(oauth *domain.OAuth) error {
	err := runOAuthValFns(oauth,
		ov.userIdValid,
		ov.providerRequired,
		ov.providerUserIdRequired)
	if err != nil {
		return err
	}
	return ov.oauthGorm.Create(oauth)
}

// Update runs validations needed for updating new OAuth database records.
func (ov *oauthValidator) Update(oauth *domain.OAuth) error {
	err := runOAuthValFns(oauth,
		ov.idValid,
		ov.userIdValid,
		ov.providerRequired,
		ov.providerUserIdRequired)
	if err != nil {
		return err
	}
	return ov.oauthGorm.Update(oauth)
}

// runOAuthValFns runs any number of functions of type oauthValFn on the passed in OAuth object.
// If none of them returns an error, it returns nil. Otherwise, it returns the respective error.
func runOAuthValFns(oauth *domain.OAuth, fns ...oauthValFn) error {
	for _, fn := range fns {
		if err := fn(oauth); err != nil {
			return err
		}
	}
	return nil
}

// A oauthValFn is any function that takes in a pointer to a domain.OAuth object and returns an error.
type oauthValFn = func(oauth *domain.OAuth) error

// idValid ensures that the passed in ID of a OAuth to be updated is greater than 0.
func (ov *oauthValidator) idValid(oauth *domain.OAuth) error {
	if oauth.ID <= 0 {
		return errs.IdInvalid
	}
	return nil
}

// providerRequired ensures that the name of a provider is not empty.
func (ov *oauthValidator) providerRequired(oauth *domain.OAuth) error {
	if oauth.Provider == "" {
		return errs.ProviderRequired
	}
	return nil
}

// providerUserIdRequired ensures that the user's unique identifier from the provider's system
// is not empty.
func (ov *oauthValidator) providerUserIdRequired(oauth *domain.OAuth) error {
	if oauth.ProviderUserID == "" {
		return errs.ProviderUserIdRequired
	}
	return nil
}

// userIdValid ensures that the userId is not empty.
func (ov *oauthValidator) userIdValid(oauth *domain.OAuth) error {
	if oauth.UserID <= 0 {
		return errs.UserIdValid
	}
	return nil
}

// ByProviderUserId takes the name of an oauth provider and a unique identifier of a user in
// a provider's system. The identifier is sent by an oauth provider after a user authorized
// this app's access on their provider profile. It's used to check if someone with that identifier
// on that provider has previously signed in here, which is true if an oauth record with that
// data exists. The record's user_id would then be used to identify the user in our database.
// It returns a pointer to an oauth object or an error.
func (og *oauthGorm) ByProviderUserId(provider, providerUserId string) (*domain.OAuth, error) {
	var oauth domain.OAuth
	err := og.db.
		Where("provider = ?", provider).
		Where("provider_user_id", providerUserId).
		First(&oauth).Error
	if err != nil {
		return nil, err
	}
	return &oauth, nil
}

// Create stores the data from the OAuth object in a new database record.
func (og *oauthGorm) Create(oauth *domain.OAuth) error {
	return og.db.Create(oauth).Error
}

// Update saves changes to an existing oauth record in the database.
// It's used to renew an existing user's existing oauth token data, when they sign in
// with a provider that they've already used in the past. Needed because some providers
// issue oauth tokens that expire after some time.
func (og *oauthGorm) Update(oauth *domain.OAuth) error {
	return og.db.Save(oauth).Error
}
