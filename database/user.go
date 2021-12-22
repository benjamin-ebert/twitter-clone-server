package database

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"unicode/utf8"

	"wtfTwitter/auth"
	"wtfTwitter/domain"
	"wtfTwitter/errs"
)

//var _ domain.UserService = (*UserService)(nil)
var _ domain.UserService = &UserService{}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		userValidator{
			hmac:     auth.NewHMAC("blerz"),
			pepper:   "blerz",
			emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),

			userGorm: userGorm{
				db: db,
			},
		},
	}
}

type UserService struct {
	userValidator
}

type userValidator struct {
	hmac auth.HMAC
	pepper string
	emailRegex *regexp.Regexp
	userGorm
}

type userGorm struct {
	db *gorm.DB
}

func (u *UserService) FindUserByID(ctx context.Context, id int) (*domain.User, error) {
	panic("implement me")
}

func (u *UserService) FindUsers(ctx context.Context, filter domain.UserFilter) ([]*domain.User, int, error) {
	panic("implement me")
}

func (u *UserService) CreateUser(ctx context.Context, user *domain.User) error {
	fmt.Println("INSIDE UserService@CREATE: ", user)
	//err := u.db.Create(user).Error
	//if err != nil {
	//	return err
	//}
	return u.userValidator.CreateUser(ctx, user)
}

func (u *UserService) UpdateUser(ctx context.Context, user *domain.User) error {
	fmt.Println("INSIDE UserService@UPDATE: ", user)
	return u.userValidator.UpdateUser(ctx, user)
}

func (u *UserService) DeleteUser(ctx context.Context, id int) error {
	panic("implement me")
}

// CreateUser will create the provided user and backfill data
// like the ID, CreatedAt, and UpdatedAt fields.
func (uv *userValidator) CreateUser(ctx context.Context, user *domain.User) error {
	fmt.Println("INSIDE userValidator@CREATE: ", user)
	err := runUserValFuncs(user,
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
	return uv.userGorm.CreateUser(ctx, user)
}

// UpdateUser will hash a remember token if it is provided.
func (uv *userValidator) UpdateUser(ctx context.Context, user *domain.User) error {
	fmt.Println("INSIDE userValidator@UPDATE: ", user)
	err := runUserValFuncs(user,
		uv.passwordMinLength,
		uv.passwordBcrypt,
		uv.passwordHashRequired,
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
	return uv.userGorm.UpdateUser(ctx, user)
}

func (uv *userValidator) FindUserByRemember(token string) (*domain.User, error) {
	user := domain.User{
		Remember: token,
	}
	if err := runUserValFuncs(&user, uv.rememberHmac); err != nil {
		return nil, err
	}
	return uv.userGorm.FindUserByRemember(user.RememberHash)
}

type userValFunc func(user *domain.User) error

func runUserValFuncs(user *domain.User, fns ...userValFunc) error {
	i := 0
	for _, fn := range fns {
		i++
		fmt.Println(i)
		fmt.Println("INSIDE runUserValFuncs: ", user)
		if err := fn(user); err != nil {
			return err
		}
	}
	return nil
}

func (uv *userValidator) emailFormat(user *domain.User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return errs.EmailInvalid
	}
	return nil
}

func (uv *userValidator) emailIsAvail(user *domain.User) error {
	existing, err := uv.userGorm.FindUserByEmail(user.Email)
	if err == errs.NotFound {
		return nil // Address is not taken.
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID { // If they are the same, it's just an update.
		return errs.EmailTaken
	}
	return nil
}

func (uv *userValidator) emailNormalize(user *domain.User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) emailRequired(user *domain.User) error {
	if user.Email == "" {
		return errs.EmailRequired
	}
	return nil
}

// passwordBcrypt hashes a user's password with a
// predefined pepper (userPwPepper) and bcrypts it,
// if the Password field is not the empty string.
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

func (uv *userValidator) passwordHashRequired(user *domain.User) error {
	if user.PasswordHash == "" {
		return errs.PasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordMinLength(user *domain.User) error {
	if user.Password == "" {
		return nil
	}
	if utf8.RuneCountInString(user.Password) < 8 {
		return errs.PasswordTooShort
	}
	return nil
}

func (uv *userValidator) passwordRequired(user *domain.User) error {
	if user.Password == "" {
		return errs.PasswordRequired
	}
	return nil
}

func (uv *userValidator) rememberHashRequired(user *domain.User) error {
	if user.RememberHash == "" {
		return errs.RememberRequired
	}
	return nil
}

func (uv *userValidator) rememberHmac(user *domain.User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) rememberMinBytes(user *domain.User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := auth.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return errs.RememberTooShort
	}
	return nil
}

func (uv *userValidator) rememberSetIfUnset(user *domain.User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := auth.MakeRememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (ug *userGorm) CreateUser(ctx context.Context, user *domain.User) error {
	fmt.Println("INSIDE userGorm@CREATE: ", user)
	err := ug.db.Create(user).Error
	if err != nil {
		 return err
	}
	return nil
}

func (ug *userGorm) UpdateUser(ctx context.Context, user *domain.User) error {
	fmt.Println("INSIDE userGorm@UPDATE: ", user)
	return ug.db.Save(user).Error
}

func (ug *userGorm) FindUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	db := ug.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func (ug *userGorm) FindUserByRemember(rememberHash string) (*domain.User, error) {
	var user domain.User
	db := ug.db.Where("remember_hash = ?", rememberHash)
	err := first(db, &user)
	return &user, err
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.
		Preload("Tweets.Replies").
		Preload("Tweets.Retweets").
		Preload("Tweets.Likes").
		Preload("Likes.Tweet").
		First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return errs.NotFound
	}
	return err
}
