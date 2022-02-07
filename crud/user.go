package crud

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"hash"
	"regexp"
	"strings"
	"unicode/utf8"

	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

// UserService manages Users. It also contains the part of the authentication system
// that handles database interactions and token creation / hashing. It's basically
// the "backend" of the auth system, with http/auth.go dealing with requests, middleware
// and cookies being the "frontend". It implements the domain.UserService interface.
type UserService struct {
	userValidator
}

// userValidator runs validations on incoming User data.
// On success, it passes the data on to userGorm.
// Otherwise, it returns the error of the validation that has failed.
type userValidator struct {
	hmac HMAC
	pepper string
	emailRegex *regexp.Regexp
	userGorm
}

// userGorm runs CRUD operations on the database using incoming User data.
// It assumes that data has been validated. On success, it returns nil.
// Otherwise, it returns the error of the operation that has failed.
type userGorm struct {
	db *gorm.DB
}

// NewUserService returns an instance of UserService.
func NewUserService(db *gorm.DB, hmacKey, pepper string) *UserService {
	return &UserService{
		userValidator{
			hmac:     newHMAC(hmacKey),
			pepper:   pepper,
			emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
			userGorm: userGorm{
				db: db,
			},
		},
	}
}

// Ensure the UserService struct properly implements the domain.UserService interface.
// If it does not, then this expression becomes invalid and won't compile.
var _ domain.UserService = &UserService{}

// Authenticate checks a submitted email address and password for existence and correctness.
func (uv *userValidator) Authenticate(email, password string) (*domain.User, error) {
	// Look for a user database record containing the submitted email address.
	found, err := uv.userGorm.ByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.Errorf(errs.EINVALID, "The email address does not exist in our database.")
		} else {
			return nil, err
		}
	}

	// Append a predefined pepper to the submitted password, hash it, and compare the result to the
	// password hash stored in the user's database record. If they match, the submitted password is correct.
	err = bcrypt.CompareHashAndPassword([]byte(found.PasswordHash), []byte(password + uv.pepper))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return nil, errs.Errorf(errs.EINVALID, "The password is incorrect.")
		} else {
			return nil, err
		}
	}

	// Return the now authenticated user and a nil error.
	return found, nil
}

// MakeRememberToken is helper to generate remember tokens of a predetermined byte size.
func (uv *userValidator) MakeRememberToken() (string, error) {
	return bytesToString(RememberTokenBytes)
}

// ByRemember runs validations / normalizations on a user's remember token. It then passes
// the HASHED remember token on to userGorm.ByRemember, will look it up in the database.
func (uv *userValidator) ByRemember(token string) (*domain.User, error) {
	user := domain.User{
		Remember: token,
	}
	if err := runUserValFns(&user, uv.rememberHmac); err != nil {
		return nil, err
	}
	return uv.userGorm.ByRemember(user.RememberHash)
}

// Create runs validations needed for creating new User database records.
// It will create a remember token if none is provided.
func (uv *userValidator) Create(ctx context.Context, user *domain.User) error {
	err := runUserValFns(user,
		uv.passwordRequired,
		uv.passwordMinLength,
		uv.passwordBcrypt,
		uv.passwordHashRequired,
		uv.rememberSetIfUnset,
		uv.rememberMinBytes,
		uv.rememberHmac,
		uv.rememberHashRequired,
		uv.emailNormalize,
		uv.emailRequired,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}
	return uv.userGorm.Create(ctx, user)
}

// Update runs validations needed for updating a User record in the database.
// It will hash a remember token if it is provided (and will not return an error if it's not).
func (uv *userValidator) Update(ctx context.Context, user *domain.User) error {
	err := runUserValFns(user,
		uv.passwordMinLength,
		uv.passwordBcrypt,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.rememberHmac,
		uv.rememberHashRequired,
		uv.emailNormalize,
		uv.emailRequired,
		uv.emailFormat,
		uv.emailIsAvail,
		)
	if err != nil {
		return err
	}
	return uv.userGorm.Update(ctx, user)
}

// runUserValFns runs any number of functions of type userValFn on the passed in User object.
// If none of them returns an error, it returns nil. Otherwise, it returns the respective error.
func runUserValFns(user *domain.User, fns ...userValFn) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

// A userValFn is any function that takes in a pointer to a domain.User object and returns an error.
type userValFn func(user *domain.User) error

// emailFormat makes sure that a provided email address matches a predefined regex pattern.
func (uv *userValidator) emailFormat(user *domain.User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return errs.Errorf(errs.EINVALID, "The email address is invalid.")
	}
	return nil
}

// emailIsAvail makes sure that a provided email address is not yet taken.
func (uv *userValidator) emailIsAvail(user *domain.User) error {
	existing, err := uv.userGorm.ByEmail(user.Email)
	if err == gorm.ErrRecordNotFound {
		// Address is not taken.
		return nil
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID {
		// Email found, and the passed in user is not the owner of that email.
		return errs.Errorf(errs.EINVALID, "This email address is already taken.")
	}
	return nil
}

// emailNormalize converts the email to all lowercase and trims its whitespaces.
func (uv *userValidator) emailNormalize(user *domain.User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

// emailRequired makes sure that the email is not the empty string.
func (uv *userValidator) emailRequired(user *domain.User) error {
	if user.Email == "" {
		return errs.Errorf(errs.EINVALID, "An email address is required.")
	}
	return nil
}

// passwordBcrypt hashes a user's password with a predefined pepper.
// It bcrypts it, if the Password field is not the empty string.
// It then clears the password on the user object in memory for security reasons.
func (uv *userValidator) passwordBcrypt(user *domain.User) error {
	if user.Password == "" {
		return nil
	}
	pwBytes := []byte(user.Password + uv.pepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(pwBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""
	return nil
}

// passwordHashRequired makes sure that the user's password hash is not the empty string.
func (uv *userValidator) passwordHashRequired(user *domain.User) error {
	if user.NoPasswordNeeded == true {
		return nil
	}
	if user.PasswordHash == "" {
		return errs.Errorf(errs.EINVALID, "A password is required.")
	}
	return nil
}

// passwordMinLength makes sure that the user's password is at least 8 characters long.
func (uv *userValidator) passwordMinLength(user *domain.User) error {
	if user.Password == "" {
		return nil
	}
	if utf8.RuneCountInString(user.Password) < 8 {
		return errs.Errorf(errs.EINVALID, "The password must have at least 8 characters.")
	}
	return nil
}

// passwordRequired makes sure that the user's password is not the empty string.
func (uv *userValidator) passwordRequired(user *domain.User) error {
	if user.NoPasswordNeeded == true {
		return nil
	}
	if user.Password == "" {
		return errs.Errorf(errs.EINVALID, "A password is required.")
	}
	return nil
}

// rememberHashRequired makes sure the user's remember token hash is not the empty string.
func (uv *userValidator) rememberHashRequired(user *domain.User) error {
	if user.RememberHash == "" {
		return errs.RememberHashEmpty
	}
	return nil
}

// rememberHmac creates the user's remember token hash, if a remember token has been provided.
func (uv *userValidator) rememberHmac(user *domain.User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.hash(user.Remember)
	return nil
}

// rememberMinBytes makes sure that the user's remember token is not too short.
func (uv *userValidator) rememberMinBytes(user *domain.User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := nBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return errs.RememberTooShort
	}
	return nil
}

// rememberSetIfUnset creates the user's remember token if none is provided.
func (uv *userValidator) rememberSetIfUnset(user *domain.User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := uv.MakeRememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

// ByID retrieves a User database record by ID, along with its associated Tweets, Likes, Followers
// and "Followeds" (users whom the user is following), along with their most relevant associations.
// TODO: Only original tweets + relations, and followers and followeds without relations.
func (ug *userGorm) ByID(id int) (*domain.User, error) {
	var user domain.User
	err := ug.db.
		Preload("Tweets.Replies").
		Preload("Tweets.Retweets").
		Preload("Tweets.Likes").
		Preload("Likes.Tweet").
		Preload("Followers.Follower").
		Preload("Followeds.Followed").
		First(&user, "id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errs.Errorf(errs.ENOTFOUND, "The user does not exist")
		} else {
			return nil, err
		}
	}
	return &user, nil
}

// ByEmail retrieves a User database record by Email.
func (ug *userGorm) ByEmail(email string) (*domain.User, error) {
	var user domain.User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

// ByRemember retrieves a User database record by its hashed remember token.
// The checkUser middleware calls this on every request, trying to identify a user
// by matching a request cookie's remember token to a hashed remember token in the database.
func (ug *userGorm) ByRemember(rememberHash string) (*domain.User, error) {
	var user domain.User
	db := ug.db.Where("remember_hash = ?", rememberHash)
	err := first(db, &user)
	return &user, err
}

// Create stores the data from the User object in a new database record.
func (ug *userGorm) Create(ctx context.Context, user *domain.User) error {
	err := ug.db.Create(user).Error
	if err != nil {
		return err
	}
	return nil
}

// Update saves changes to an existing user record in the database.
func (ug *userGorm) Update(ctx context.Context, user *domain.User) error {
	return ug.db.Save(user).Error
}

// first is a helper for getting the first database record that matches a given query.
func first(db *gorm.DB, dst interface{}) error {
	return db.First(dst).Error
}

// HMAC is a wrapper around the crypto/hmac package making it easier to use.
type HMAC struct {
	hmac hash.Hash
}

// newHMAC creates and returns a new HMAC object.
func newHMAC(key string) HMAC {
	h := hmac.New(sha256.New, []byte(key))
	return HMAC{
		hmac: h,
	}
}

// hash hashes an input string using HMAC with the secret key
// provided when the HMAC object was created in NewUserService.
func (h HMAC) hash(input string) string {
	h.hmac.Reset()
	h.hmac.Write([]byte(input))
	b := h.hmac.Sum(nil)
	return base64.URLEncoding.EncodeToString(b)
}

const RememberTokenBytes = 32

// bytes generates n random bytes or returns an error. It uses the
// crypto/rand package, so it can be used for things like remember tokens.
func bytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// nBytes returns the number of bytes used in a base64 URL encoded string.
func nBytes(base64String string) (int, error) {
	b, err := base64.URLEncoding.DecodeString(base64String)
	if err != nil {
		return -1, err
	}
	return len(b), nil
}

// String generates a byte slice of size nBytes and then returns a
// string that is the base64 URL encoded version of that byte slice.
func bytesToString(nBytes int) (string, error) {
	b, err := bytes(nBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
