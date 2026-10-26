package validators

import (
	"errors"
	"testing"

	"tesina/backend/internal/models"
)

func Test_validateEmailFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{name: "simple", email: "gerrandonea@gmail.com"},
		{name: "dots and plus tag", email: "nombre.apellido+tag@empresa.com.ar"},
		{name: "two-letter tld", email: "a@b.io"},
		{name: "hyphenated subdomain", email: "usuario@sub-dominio.example.org"},
		{name: "digits after tld", email: "gerrandonea@gmail.com123", wantErr: ErrEmailInvalid},
		{name: "no tld", email: "gerrandonea@gmail", wantErr: ErrEmailInvalid},
		{name: "one-letter tld", email: "gerrandonea@gmail.c", wantErr: ErrEmailInvalid},
		{name: "empty domain label", email: "gerrandonea@.com", wantErr: ErrEmailInvalid},
		{name: "double dot", email: "gerrandonea@gmail..com", wantErr: ErrEmailInvalid},
		{name: "trailing dot", email: "gerrandonea@gmail.com.", wantErr: ErrEmailInvalid},
		{name: "no at sign", email: "gerrandoneagmail.com", wantErr: ErrEmailInvalid},
		{name: "no local part", email: "@gmail.com", wantErr: ErrEmailInvalid},
		{name: "inner space", email: "ger randonea@gmail.com", wantErr: ErrEmailInvalid},
		{name: "surrounding spaces", email: " gerrandonea@gmail.com ", wantErr: ErrEmailInvalid},
		{name: "two at signs", email: "ger@ran@gmail.com", wantErr: ErrEmailInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := validateEmailFormat(tt.email); !errors.Is(err, tt.wantErr) {
				t.Fatalf("validateEmailFormat(%q): got %v want %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func Test_emailFormatInRequests(t *testing.T) {
	t.Parallel()

	bad := "gerrandonea@gmail.com123"
	empty := ""

	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{
			name:    "register rejects malformed email",
			err:     ValidateCreateUserRequest(models.CreateUserRequest{Email: bad, Password: "x", FirstName: "a", LastName: "b"}),
			wantErr: ErrEmailInvalid,
		},
		{
			name:    "register still reports missing email first",
			err:     ValidateCreateUserRequest(models.CreateUserRequest{Password: "x", FirstName: "a", LastName: "b"}),
			wantErr: ErrEmailRequired,
		},
		{
			name:    "profile update rejects malformed email",
			err:     ValidateUpdateUserRequest(models.UpdateUserRequest{Email: &bad}),
			wantErr: ErrEmailInvalid,
		},
		{
			name:    "profile update keeps empty email error",
			err:     ValidateUpdateUserRequest(models.UpdateUserRequest{Email: &empty}),
			wantErr: ErrEmailEmpty,
		},
		{
			name:    "forgot password rejects malformed email",
			err:     ValidateForgotPasswordRequest(models.ForgotPasswordRequest{Email: bad}),
			wantErr: ErrEmailInvalid,
		},
		{
			name: "forgot password trims before checking",
			err:  ValidateForgotPasswordRequest(models.ForgotPasswordRequest{Email: "  gerrandonea@gmail.com "}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if !errors.Is(tt.err, tt.wantErr) {
				t.Fatalf("got %v want %v", tt.err, tt.wantErr)
			}
		})
	}
}
