package database

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"unicode/utf8"
	"wtfTwitter/domain"
	"wtfTwitter/security"
)

var _ domain.UserService = (*UserService)(nil)

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		userValidator{
			hmac:     security.NewHMAC("blerz"),
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
	hmac security.HMAC
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
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.setRememberIfUnset,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
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
		uv.bcryptPassword,
		uv.passwordHashRequired,
		uv.rememberMinBytes,
		uv.hmacRemember,
		uv.rememberHashRequired,
		uv.normalizeEmail,
		uv.requireEmail,
		uv.emailFormat,
		uv.emailIsAvail)
	if err != nil {
		return err
	}
	return uv.userGorm.UpdateUser(ctx, user)
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

func (uv *userValidator) passwordRequired(user *domain.User) error {
	if user.Password == "" {
		return security.ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) passwordMinLength(user *domain.User) error {
	if user.Password == "" {
		return nil
	}
	if utf8.RuneCountInString(user.Password) < 8 {
		return security.ErrPasswordTooShort
	}
	return nil
}

// bcryptPassword hashes a user's password with a
// predefined pepper (userPwPepper) and bcrypts it,
// if the Password field is not the empty string.
func (uv *userValidator) bcryptPassword(user *domain.User) error {
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
		return security.ErrPasswordRequired
	}
	return nil
}

func (uv *userValidator) setRememberIfUnset(user *domain.User) error {
	if user.Remember != "" {
		return nil
	}
	token, err := security.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uv *userValidator) rememberMinBytes(user *domain.User) error {
	if user.Remember == "" {
		return nil
	}
	n, err := security.NBytes(user.Remember)
	if err != nil {
		return err
	}
	if n < 32 {
		return security.ErrRememberTooShort
	}
	return nil
}

func (uv *userValidator) hmacRemember(user *domain.User) error {
	if user.Remember == "" {
		return nil
	}
	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) rememberHashRequired(user *domain.User) error {
	if user.RememberHash == "" {
		return security.ErrRememberRequired
	}
	return nil
}

func (uv *userValidator) normalizeEmail(user *domain.User) error {
	user.Email = strings.ToLower(user.Email)
	user.Email = strings.TrimSpace(user.Email)
	return nil
}

func (uv *userValidator) requireEmail(user *domain.User) error {
	if user.Email == "" {
		return security.ErrEmailRequired
	}
	return nil
}

func (uv *userValidator) emailFormat(user *domain.User) error {
	if user.Email == "" {
		return nil
	}
	if !uv.emailRegex.MatchString(user.Email) {
		return security.ErrEmailInvalid
	}
	return nil
}

func (uv *userValidator) emailIsAvail(user *domain.User) error {
	existing, err := uv.userGorm.FindUserByEmail(user.Email)
	if err == security.ErrNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID {
		return security.ErrEmailTaken
	}
	return nil
}

func (g *userGorm) CreateUser(ctx context.Context, user *domain.User) error {
	fmt.Println("INSIDE userGorm@CREATE: ", user)
	err := g.db.Create(user).Error
	if err != nil {
		 return err
	}
	return nil
}

func (g *userGorm) UpdateUser(ctx context.Context, user *domain.User) error {
	fmt.Println("INSIDE userGorm@UPDATE: ", user)
	return g.db.Save(user).Error
}

func (g *userGorm) FindUserByEmail(email string) (*domain.User, error) {
	var user domain.User
	db := g.db.Where("email = ?", email)
	err := first(db, &user)
	return &user, err
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return security.ErrNotFound
	}
	return err
}
