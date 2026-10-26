package validators

import (
	"errors"
	"regexp"
)

var ErrEmailInvalid = errors.New("email format is invalid")

// emailPattern requires a local part, "@", a dotted domain and a TLD of at
// least two letters, so typos like "user@gmail.com123" are rejected.
var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@.]+(\.[^\s@.]+)*\.[A-Za-z]{2,}$`)

// validateEmailFormat checks a non-empty email against emailPattern.
func validateEmailFormat(email string) error {
	if !emailPattern.MatchString(email) {
		return ErrEmailInvalid
	}
	return nil
}
