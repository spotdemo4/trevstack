package v1

import (
	"regexp"

	"trev.zip/template/stack/server/validation"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// credentials are the fields LoginRequest and SignupRequest share.
type credentials interface {
	HasUsername() bool
	GetUsername() string
	HasPassword() bool
	GetPassword() string
}

func validateCredentials(req credentials) error {
	return validation.First(
		validation.Required("username", req.HasUsername()),
		validation.Length("username", req.GetUsername(), 3, 64),
		validation.Pattern("username", req.GetUsername(), usernamePattern,
			"start with a letter or digit and contain only letters, digits, underscores, and hyphens"),
		validation.Required("password", req.HasPassword()),
		validation.Length("password", req.GetPassword(), 8, 72),
		validation.MaxBytes("password", req.GetPassword(), 72),
	)
}
