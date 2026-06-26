package models

import (
	"errors"
	"testing"
)

func mustCreateInvoice(t *testing.T, m *InvoiceModel, workerID int, date, houseNumber, customerName string, jobs, materials []string) int {
	t.Helper()
	id, err := m.Create(workerID, date, houseNumber, customerName, nil, "08:00", "16:00", 2, "", jobs, materials)
	if err != nil {
		t.Fatalf("Create invoice: %v", err)
	}
	return id
}

func TestInvoiceCreate(t *testing.T) {
	db := newTestDB(t)
	im := &InvoiceModel{DB: db}
	workerID := mustInsertWorker(t, db, "Worker A", "1111", true)
	customerID := mustInsertCustomer(t, db, "Jane Smith", "42", "42 Oak St")

	t.Run("with jobs, materials, and customer link", func(t *testing.T) {
		id, err := im.Create(
			workerID, "2026-06-26", "42", "Jane Smith", &customerID,
			"08:00", "16:00", 3, "Looks good",
			[]string{"Mow lawn", "Trim hedges"},
			[]string{"Fertilizer", "Mulch"},
		)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		inv, err := im.Get(id)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}

		if inv.Date != "2026-06-26" {
			t.Errorf("Date: got %q, want 2026-06-26", inv.Date)
		}
		if inv.HouseNumber != "42" {
			t.Errorf("HouseNumber: got %q, want 42", inv.HouseNumber)
		}
		if inv.CustomerName != "Jane Smith" {
			t.Errorf("CustomerName: got %q, want Jane Smith", inv.CustomerName)
		}
		if inv.CustomerID == nil || *inv.CustomerID != customerID {
			t.Errorf("CustomerID: got %v, want %d", inv.CustomerID, customerID)
		}
		if inv.NoOfWorkers != 3 {
			t.Errorf("NoOfWorkers: got %d, want 3", inv.NoOfWorkers)
		}
		if inv.Comments != "Looks good" {
			t.Errorf("Comments: got %q, want 'Looks good'", inv.Comments)
		}
		if inv.TimeArrived != "08:00" || inv.TimeLeft != "16:00" {
			t.Errorf("Times: got arrived=%q left=%q", inv.TimeArrived, inv.TimeLeft)
		}
		if inv.Reviewed {
			t.Error("newly created invoice should not be reviewed")
		}

		if len(inv.Jobs) != 2 || inv.Jobs[0] != "Mow lawn" || inv.Jobs[1] != "Trim hedges" {
			t.Errorf("Jobs: got %v, want [Mow lawn Trim hedges]", inv.Jobs)
		}
		if len(inv.Materials) != 2 || inv.Materials[0] != "Fertilizer" || inv.Materials[1] != "Mulch" {
			t.Errorf("Materials: got %v, want [Fertilizer Mulch]", inv.Materials)
		}
	})

	t.Run("without jobs or materials", func(t *testing.T) {
		id, err := im.Create(workerID, "2026-06-26", "99", "", nil, "09:00", "12:00", 1, "", nil, nil)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		inv, err := im.Get(id)
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if len(inv.Jobs) != 0 {
			t.Errorf("Jobs: got %v, want empty", inv.Jobs)
		}
		if len(inv.Materials) != 0 {
			t.Errorf("Materials: got %v, want empty", inv.Materials)
		}
		if inv.CustomerID != nil {
			t.Errorf("CustomerID: got %v, want nil", inv.CustomerID)
		}
	})
}

func TestInvoiceGet(t *testing.T) {
	db := newTestDB(t)
	im := &InvoiceModel{DB: db}
	workerID := mustInsertWorker(t, db, "Worker B", "2222", true)

	id := mustCreateInvoice(t, im, workerID, "2026-06-26", "5", "Bob", nil, nil)

	inv, err := im.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if inv.WorkerName != "Worker B" {
		t.Errorf("WorkerName: got %q, want Worker B", inv.WorkerName)
	}

	_, err = im.Get(id + 999)
	if !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for missing invoice", err)
	}
}

func TestInvoiceList(t *testing.T) {
	db := newTestDB(t)
	im := &InvoiceModel{DB: db}
	workerID := mustInsertWorker(t, db, "Worker C", "3333", true)

	mustCreateInvoice(t, im, workerID, "2026-06-24", "1", "A", nil, nil)
	mustCreateInvoice(t, im, workerID, "2026-06-25", "2", "B", nil, nil)
	mustCreateInvoice(t, im, workerID, "2026-06-26", "3", "C", nil, nil)

	t.Run("returns all invoices newest first", func(t *testing.T) {
		invoices, total, err := im.List(1)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 3 {
			t.Errorf("total: got %d, want 3", total)
		}
		if len(invoices) != 3 {
			t.Fatalf("got %d invoices, want 3", len(invoices))
		}
		// Newest first
		if invoices[0].Date != "2026-06-26" {
			t.Errorf("first invoice date: got %q, want 2026-06-26", invoices[0].Date)
		}
		if invoices[2].Date != "2026-06-24" {
			t.Errorf("last invoice date: got %q, want 2026-06-24", invoices[2].Date)
		}
	})

	t.Run("out-of-range page returns empty", func(t *testing.T) {
		invoices, total, err := im.List(999)
		if err != nil {
			t.Fatalf("List page 999: %v", err)
		}
		if total != 3 {
			t.Errorf("total: got %d, want 3", total)
		}
		if len(invoices) != 0 {
			t.Errorf("got %d invoices, want 0 for out-of-range page", len(invoices))
		}
	})
}

func TestInvoiceSetReviewed(t *testing.T) {
	db := newTestDB(t)
	im := &InvoiceModel{DB: db}
	workerID := mustInsertWorker(t, db, "Worker D", "4444", true)

	id := mustCreateInvoice(t, im, workerID, "2026-06-26", "7", "Carol", nil, nil)

	inv, _ := im.Get(id)
	if inv.Reviewed {
		t.Fatal("invoice should start unreviewed")
	}

	if err := im.SetReviewed(id); err != nil {
		t.Fatalf("SetReviewed: %v", err)
	}

	inv, _ = im.Get(id)
	if !inv.Reviewed {
		t.Error("invoice should be reviewed after SetReviewed")
	}

	if err := im.SetReviewed(id + 999); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for missing invoice", err)
	}
}

func TestInvoiceListByCustomer(t *testing.T) {
	db := newTestDB(t)
	im := &InvoiceModel{DB: db}
	workerID := mustInsertWorker(t, db, "Worker E", "5555", true)
	customerA := mustInsertCustomer(t, db, "Alice", "10", "")
	customerB := mustInsertCustomer(t, db, "Bob", "20", "")

	// Two invoices for customer A, one for customer B.
	im.Create(workerID, "2026-06-24", "10", "Alice", &customerA, "08:00", "12:00", 1, "", nil, nil)
	im.Create(workerID, "2026-06-26", "10", "Alice", &customerA, "08:00", "16:00", 2, "", nil, nil)
	im.Create(workerID, "2026-06-25", "20", "Bob", &customerB, "09:00", "13:00", 1, "", nil, nil)

	t.Run("returns only invoices for the given customer", func(t *testing.T) {
		invoices, err := im.ListByCustomer(customerA)
		if err != nil {
			t.Fatalf("ListByCustomer: %v", err)
		}
		if len(invoices) != 2 {
			t.Fatalf("got %d invoices for customer A, want 2", len(invoices))
		}
		for _, inv := range invoices {
			if inv.CustomerName != "Alice" {
				t.Errorf("unexpected customer name: %q", inv.CustomerName)
			}
		}
		// Newest first
		if invoices[0].Date != "2026-06-26" {
			t.Errorf("first invoice date: got %q, want 2026-06-26", invoices[0].Date)
		}
	})

	t.Run("returns empty for customer with no invoices", func(t *testing.T) {
		unknown := mustInsertCustomer(t, db, "Nobody", "99", "")
		invoices, err := im.ListByCustomer(unknown)
		if err != nil {
			t.Fatalf("ListByCustomer: %v", err)
		}
		if len(invoices) != 0 {
			t.Errorf("got %d invoices, want 0", len(invoices))
		}
	})
}
