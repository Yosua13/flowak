package auth_test

import (
	"testing"

	"backend/internal/domain/auth"
)

func TestAuthSentinelErrors(t *testing.T) {
	errs := []error{
		auth.ErrInvalidCredentials,
		auth.ErrEmailAlreadyRegistered,
		auth.ErrUserNotFound,
		auth.ErrInvalidSession,
		auth.ErrNoActiveOrganization,
		auth.ErrTokenBlacklisted,
		auth.ErrAllFieldsRequired,
		auth.ErrInvalidNameLength,
		auth.ErrInvalidEmailFormat,
		auth.ErrPasswordTooShort,
		auth.ErrElevatedRoleRestricted,
	}

	for _, err := range errs {
		if err == nil || err.Error() == "" {
			t.Errorf("expected non-empty error, got %v", err)
		}
	}
}
