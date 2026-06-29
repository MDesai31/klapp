package models

import (
	"errors"
	"testing"
)

func TestWorkerAuthenticate(t *testing.T) {
	db := newTestDB(t)
	wm := &WorkerModel{DB: db}

	activeID := mustInsertWorker(t, db, "Manthan", "1234", true)
	mustInsertWorker(t, db, "Retired Worker", "9999", false)

	t.Run("correct PIN", func(t *testing.T) {
		w, err := wm.Authenticate("1234")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if w.ID != activeID || w.WorkerName != "Manthan" {
			t.Errorf("got %+v, want worker %d named Manthan", w, activeID)
		}
	})

	t.Run("wrong PIN", func(t *testing.T) {
		_, err := wm.Authenticate("0000")
		if !errors.Is(err, ErrInvalidPIN) {
			t.Errorf("got error %v, want ErrInvalidPIN", err)
		}
	})

	t.Run("inactive worker's PIN is rejected", func(t *testing.T) {
		_, err := wm.Authenticate("9999")
		if !errors.Is(err, ErrInvalidPIN) {
			t.Errorf("got error %v, want ErrInvalidPIN", err)
		}
	})
}

func TestWorkerGet(t *testing.T) {
	db := newTestDB(t)
	wm := &WorkerModel{DB: db}

	id := mustInsertWorker(t, db, "Thomas", "4321", true)

	w, err := wm.Get(id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.WorkerName != "Thomas" {
		t.Errorf("got name %q, want Thomas", w.WorkerName)
	}

	_, err = wm.Get(id + 999)
	if !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord", err)
	}
}

func TestWorkerCreateListAndSetActive(t *testing.T) {
	db := newTestDB(t)
	wm := &WorkerModel{DB: db}

	id, err := wm.Create("New Hire", "5555", "555-0199", 0, "spanish", false)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	w, err := wm.Get(id)
	if err != nil {
		t.Fatalf("Get after Create: %v", err)
	}
	if !w.Active {
		t.Error("newly created worker should be active")
	}

	if _, err := wm.Authenticate("5555"); err != nil {
		t.Errorf("created worker's PIN should authenticate, got error %v", err)
	}

	list, err := wm.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d workers, want 1", len(list))
	}

	if err := wm.SetActive(id, false); err != nil {
		t.Fatalf("SetActive: %v", err)
	}
	if _, err := wm.Authenticate("5555"); !errors.Is(err, ErrInvalidPIN) {
		t.Errorf("deactivated worker's PIN should be rejected, got error %v", err)
	}

	if err := wm.SetActive(id+999, false); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for unknown worker", err)
	}
}

func TestWorkerUpdate(t *testing.T) {
	db := newTestDB(t)
	wm := &WorkerModel{DB: db}

	id := mustInsertWorker(t, db, "Original Name", "1111", true)

	if err := wm.Update(id, "New Name", "2222", "555-0100", 0, "spanish", false); err != nil {
		t.Fatalf("Update: %v", err)
	}

	w, err := wm.Get(id)
	if err != nil {
		t.Fatalf("Get after Update: %v", err)
	}
	if w.WorkerName != "New Name" || w.PIN != "2222" || w.Phone != "555-0100" {
		t.Errorf("got %+v, want updated name/pin/phone", w)
	}

	if _, err := wm.Authenticate("2222"); err != nil {
		t.Errorf("updated PIN should authenticate, got error %v", err)
	}

	if err := wm.Update(id+999, "Nobody", "3333", "", 0, "spanish", false); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for unknown worker", err)
	}
}

func TestWorkerDuplicatePIN(t *testing.T) {
	db := newTestDB(t)
	wm := &WorkerModel{DB: db}

	activeID := mustInsertWorker(t, db, "Active Worker", "1234", true)
	inactiveID := mustInsertWorker(t, db, "Inactive Worker", "5678", false)

	t.Run("Create rejects a PIN already used by an active worker", func(t *testing.T) {
		if _, err := wm.Create("New Hire", "1234", "", 0, "spanish", false); !errors.Is(err, ErrDuplicatePIN) {
			t.Errorf("got error %v, want ErrDuplicatePIN", err)
		}
	})

	t.Run("Create rejects a PIN already used by an inactive worker", func(t *testing.T) {
		if _, err := wm.Create("New Hire", "5678", "", 0, "spanish", false); !errors.Is(err, ErrDuplicatePIN) {
			t.Errorf("got error %v, want ErrDuplicatePIN", err)
		}
	})

	t.Run("Update rejects changing to another worker's PIN", func(t *testing.T) {
		if err := wm.Update(inactiveID, "Inactive Worker", "1234", "", 0, "spanish", false); !errors.Is(err, ErrDuplicatePIN) {
			t.Errorf("got error %v, want ErrDuplicatePIN", err)
		}
	})

	t.Run("Update allows a worker to keep its own PIN", func(t *testing.T) {
		if err := wm.Update(activeID, "Active Worker", "1234", "555-0101", 0, "spanish", false); err != nil {
			t.Errorf("unexpected error keeping own PIN: %v", err)
		}
	})
}

func TestWorkerRequireLocation(t *testing.T) {
	db := newTestDB(t)
	wm := &WorkerModel{DB: db}

	id, err := wm.Create("Field Worker", "7777", "555-0100", 15.0, "english", true)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Get round-trips the field.
	w, err := wm.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !w.RequireLocation {
		t.Error("Get: RequireLocation = false, want true")
	}

	// Authenticate returns it too.
	w, err = wm.Authenticate("7777")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if !w.RequireLocation {
		t.Error("Authenticate: RequireLocation = false, want true")
	}

	// List includes it.
	workers, err := wm.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(workers) != 1 || !workers[0].RequireLocation {
		t.Errorf("List: RequireLocation = false, want true")
	}

	// Update can clear it.
	if err := wm.Update(id, "Field Worker", "7777", "555-0100", 15.0, "english", false); err != nil {
		t.Fatalf("Update: %v", err)
	}
	w, err = wm.Get(id)
	if err != nil {
		t.Fatalf("Get after Update: %v", err)
	}
	if w.RequireLocation {
		t.Error("Get after Update: RequireLocation = true, want false")
	}
}
