package errs

import "strings"

const (
	// NotFound is returned when a resource cannot be found
	// in the database.
	NotFound modelError = "models: resource not found"
	// PasswordIncorrect is returned when an invalid password
	// is used when attempting to authenticate a user.
	PasswordIncorrect modelError = "models: incorrect password provided"
	// EmailRequired is returned when an email address is
	// not provided when creating a user
	EmailRequired modelError = "models: email address is required"
	// EmailInvalid is returned when an email address provided
	// does not match any of our requirements
	EmailInvalid modelError = "models: email address is not valid"
	// EmailTaken is returned when an update or create is attempted
	// with an email address that is already in use.
	EmailTaken modelError = "models: email address is already taken"
	// PasswordRequired is returned when a create is attempted
	// without a user password provided.
	PasswordRequired modelError = "models: password is required"
	// PasswordTooShort is returned when an update or create is
	// attempted with a user password that is less than 8 characters.
	PasswordTooShort modelError = "models: password must be at least 8 characters long"
	TitleRequired    modelError = "models: title is required"

	TokenInvalid modelError = "models: token provided is not valid"

	ContentTooShort modelError = "models: content must not be empty"
	ContentTooLong modelError ="models: content must not have more than 280 characters"

	FollowAlreadyExists modelError = "models: this follow already exists"
	FollowedDoesNotExist modelError = "models: user to be followed does not exist"
	FollowedIsFollower modelError = "models: followed and follower are the same user"

	// IDInvalid is returned when an invalid ID is provided
	// to a method like Delete.
	IDInvalid privateError = "models: ID provided was invalid"
	// RememberRequired is returned when a create or update
	// is attempted without a user remember token hash
	RememberRequired privateError = "models: remember token is required"
	// RememberTooShort is returned when a remember token is
	// not at least 32 bytes
	RememberTooShort privateError = "models: remember token must be at least 32 bytes"
	UserIDRequired   privateError = "models: user ID is required"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	split := strings.Split(s, " ")
	split[0] = strings.Title(split[0])
	return strings.Join(split, " ")
}

type privateError string

func (e privateError) Error() string {
	return string(e)
}
