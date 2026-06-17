package models

import (
	"errors"
	"testing"
)

func TestAdminAuthenticate(t *testing.T) {
	db := newTestDB(t)
	am := &AdminModel{DB: db}

	if err := am.Upsert("owner", "correct-password"); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	t.Run("correct credentials", func(t *testing.T) {
		a, err := am.Authenticate("owner", "correct-password")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if a.Username != "owner" {
			t.Errorf("got username %q, want owner", a.Username)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		_, err := am.Authenticate("owner", "wrong-password")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("got error %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("unknown username", func(t *testing.T) {
		_, err := am.Authenticate("nobody", "whatever")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("got error %v, want ErrInvalidCredentials", err)
		}
	})

	t.Run("upsert replaces password", func(t *testing.T) {
		if err := am.Upsert("owner", "new-password"); err != nil {
			t.Fatalf("Upsert: %v", err)
		}
		if _, err := am.Authenticate("owner", "correct-password"); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("old password should no longer work, got error %v", err)
		}
		if _, err := am.Authenticate("owner", "new-password"); err != nil {
			t.Errorf("new password should work, got error %v", err)
		}
	})
}
