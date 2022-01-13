package crud

import (
	"gorm.io/gorm"
	"time"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

type OAuthService struct {
	oauthValidator
}

type oauthValidator struct {
	oauthGorm
}

type oauthGorm struct {
	db *gorm.DB
}

func NewOAuthService(db *gorm.DB) *OAuthService {
	return &OAuthService{
		oauthValidator{
			oauthGorm{
				db: db,
			},
		},
	}
}

var _ domain.OAuthService = &OAuthService{}

func (ov *oauthValidator) Find(userId int, provider string) (*domain.OAuth, error) {
	return ov.oauthGorm.Find(userId, provider)
}

func (ov *oauthValidator) Create(oauth *domain.OAuth) error {
	err := runOAuthValFns(oauth,
		ov.userIdRequired,
		ov.providerRequired,
		ov.providerUserIdRequired)
	if err != nil {
		return err
	}
	return ov.oauthGorm.Create(oauth)
}

func (ov *oauthValidator) Delete(oauth *domain.OAuth) error {
	err := runOAuthValFns(oauth, ov.idValid)
	if err != nil {
		return err
	}
	return ov.oauthGorm.Delete(oauth)
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

// idValid makes sure that the passed in ID of a OAuth to be deleted is greater than 0.
func (ov *oauthValidator) idValid(oauth *domain.OAuth) error {
	if oauth.ID <= 0 {
		return errs.Errorf(errs.EINVALID, "OAuth ID is invalid.")
	}
	return nil
}

func (ov *oauthValidator) providerRequired(oauth *domain.OAuth) error {
	if oauth.Provider == "" {
		return errs.ProviderRequired
	}
	return nil
}

func (ov *oauthValidator) providerUserIdRequired(oauth *domain.OAuth) error {
	if oauth.ProviderUserID == "" {
		return errs.ProviderUserIdRequired
	}
	return nil
}

func (ov *oauthValidator) userIdRequired(oauth *domain.OAuth) error {
	if oauth.UserID <= 0 {
		return errs.UserIDRequired
	}
	return nil
}

func (og *oauthGorm) Find(userId int, provider string) (*domain.OAuth, error) {
	var oauth domain.OAuth
	err := og.db.
		Where("user_id = ?", userId).
		Where("provider = ?", provider).
		First(&oauth).Error
	if err != nil {
		return nil, err
	}
	return &oauth, nil
}

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

func (og *oauthGorm) update(existing *domain.OAuth, accessT, refreshT string, expiry time.Time) (*domain.OAuth, error) {
	return nil, nil
}

func (og *oauthGorm) Create(oauth *domain.OAuth) error {
	return og.db.Create(oauth).Error
}

func (og *oauthGorm) Update(oauth *domain.OAuth) error {
	return og.db.Save(oauth).Error
}

// Delete permanently deletes the database record matching the data from the OAuth object.
// TODO: So far I don't even need this, since oauths are only either created or updated.
func (og *oauthGorm) Delete(oauth *domain.OAuth) error {
	// TODO: If this works, do it like that on all models.
	return og.db.Delete(oauth).Error
}
