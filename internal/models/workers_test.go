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

	id, err := wm.Create("New Hire", "5555", "555-0199")
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
