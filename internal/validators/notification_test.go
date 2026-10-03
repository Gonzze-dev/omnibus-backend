package validators

import (
	"encoding/json"
	"errors"
	"testing"

	"tesina/backend/internal/roles"
)

func Test_ValidateAdminGlobalNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		role         string
		payload      string
		wantErr      error
		wantTimeLife int
	}{
		{name: "valid", role: roles.SuperAdmin, payload: `{"message":"Hola","time_life":12}`, wantTimeLife: 12},
		{name: "extra fields allowed", role: roles.SuperAdmin, payload: `{"title":"Aviso","message":"Hola","time_life":5}`, wantTimeLife: 5},
		{name: "admin role forbidden", role: roles.Admin, payload: `{"message":"Hola","time_life":12}`, wantErr: ErrNotificationGlobalSuperAdminOnly},
		{name: "empty payload", role: roles.SuperAdmin, payload: `  `, wantErr: ErrNotificationPayloadEmpty},
		{name: "malformed json", role: roles.SuperAdmin, payload: `{"message":`, wantErr: ErrNotificationPayloadInvalidJSON},
		{name: "payload not an object", role: roles.SuperAdmin, payload: `["Hola"]`, wantErr: ErrNotificationPayloadInvalidJSON},
		{name: "missing message", role: roles.SuperAdmin, payload: `{"time_life":12}`, wantErr: ErrNotificationMessageEmpty},
		{name: "empty message", role: roles.SuperAdmin, payload: `{"message":"","time_life":12}`, wantErr: ErrNotificationMessageEmpty},
		{name: "whitespace message", role: roles.SuperAdmin, payload: `{"message":"  \t\n ","time_life":12}`, wantErr: ErrNotificationMessageEmpty},
		{name: "missing time_life", role: roles.SuperAdmin, payload: `{"message":"Hola"}`, wantErr: ErrNotificationTimeLifeInvalid},
		{name: "non-positive time_life", role: roles.SuperAdmin, payload: `{"message":"Hola","time_life":0}`, wantErr: ErrNotificationTimeLifeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ValidateAdminGlobalNotification(tt.role, json.RawMessage(tt.payload))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: got %v want %v", err, tt.wantErr)
			}
			if got != tt.wantTimeLife {
				t.Fatalf("time_life: got %d want %d", got, tt.wantTimeLife)
			}
		})
	}
}

func Test_ValidateAdminLocalNotification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		payload      string
		wantErr      error
		wantMessage  string
		wantTimeLife int
	}{
		{name: "valid", payload: `{"message":"Anden 3 cerrado","time_life":30}`, wantMessage: "Anden 3 cerrado", wantTimeLife: 30},
		{name: "empty payload", payload: ``, wantErr: ErrNotificationPayloadEmpty},
		{name: "malformed json", payload: `{"message":`, wantErr: ErrNotificationPayloadInvalidJSON},
		{name: "missing message", payload: `{"time_life":30}`, wantErr: ErrNotificationMessageEmpty},
		{name: "empty message", payload: `{"message":"","time_life":30}`, wantErr: ErrNotificationMessageEmpty},
		{name: "whitespace message", payload: `{"message":"   ","time_life":30}`, wantErr: ErrNotificationMessageEmpty},
		{name: "missing time_life", payload: `{"message":"Hola"}`, wantErr: ErrNotificationTimeLifeInvalid},
		{name: "negative time_life", payload: `{"message":"Hola","time_life":-1}`, wantErr: ErrNotificationTimeLifeInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ValidateAdminLocalNotification(json.RawMessage(tt.payload))
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: got %v want %v", err, tt.wantErr)
			}
			if got.Message != tt.wantMessage {
				t.Fatalf("message: got %q want %q", got.Message, tt.wantMessage)
			}
			if got.TimeLife != tt.wantTimeLife {
				t.Fatalf("time_life: got %d want %d", got.TimeLife, tt.wantTimeLife)
			}
		})
	}
}
