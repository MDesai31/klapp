package models

import (
	"errors"
	"testing"
)

func TestCustomerCreate(t *testing.T) {
	db := newTestDB(t)
	cm := &CustomerModel{DB: db}

	id, err := cm.Create("Jane Smith", "555-0100", "42", "42 Oak Street")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero ID")
	}

	c, err := cm.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if c.Name != "Jane Smith" || c.HouseNumber != "42" || c.Address != "42 Oak Street" || c.Phone != "555-0100" {
		t.Errorf("unexpected customer fields: %+v", c)
	}
}

func TestCustomerGet(t *testing.T) {
	db := newTestDB(t)
	cm := &CustomerModel{DB: db}

	id := mustInsertCustomer(t, db, "Alice", "10", "10 Main St")

	c, err := cm.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if c.ID != id || c.Name != "Alice" {
		t.Errorf("got %+v, want id=%d name=Alice", c, id)
	}

	_, err = cm.Get(id + 999)
	if !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for missing customer", err)
	}
}

func TestCustomerGetByHouseNumber(t *testing.T) {
	db := newTestDB(t)
	cm := &CustomerModel{DB: db}

	mustInsertCustomer(t, db, "Alice", "10", "10 Oak St")
	mustInsertCustomer(t, db, "Bob", "10", "10 Elm St")
	mustInsertCustomer(t, db, "Carol", "20", "20 Oak St")

	t.Run("single match", func(t *testing.T) {
		customers, err := cm.GetByHouseNumber("20")
		if err != nil {
			t.Fatalf("GetByHouseNumber: %v", err)
		}
		if len(customers) != 1 || customers[0].Name != "Carol" {
			t.Errorf("got %v, want [Carol]", customers)
		}
	})

	t.Run("multiple matches", func(t *testing.T) {
		customers, err := cm.GetByHouseNumber("10")
		if err != nil {
			t.Fatalf("GetByHouseNumber: %v", err)
		}
		if len(customers) != 2 {
			t.Fatalf("got %d customers, want 2", len(customers))
		}
		// Results are ordered by name
		if customers[0].Name != "Alice" || customers[1].Name != "Bob" {
			t.Errorf("unexpected order: %v", customers)
		}
	})

	t.Run("no match returns empty slice", func(t *testing.T) {
		customers, err := cm.GetByHouseNumber("999")
		if err != nil {
			t.Fatalf("GetByHouseNumber: %v", err)
		}
		if len(customers) != 0 {
			t.Errorf("got %d customers, want 0", len(customers))
		}
	})
}

func TestCustomerList(t *testing.T) {
	db := newTestDB(t)
	cm := &CustomerModel{DB: db}

	mustInsertCustomer(t, db, "Zara", "5", "")
	mustInsertCustomer(t, db, "Aaron", "6", "")
	mustInsertCustomer(t, db, "Maria", "7", "")

	customers, err := cm.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(customers) != 3 {
		t.Fatalf("got %d customers, want 3", len(customers))
	}
	// List is ordered by name
	if customers[0].Name != "Aaron" || customers[1].Name != "Maria" || customers[2].Name != "Zara" {
		t.Errorf("unexpected order: %v", customers)
	}
}

func TestCustomerSearch(t *testing.T) {
	db := newTestDB(t)
	cm := &CustomerModel{DB: db}

	mustInsertCustomer(t, db, "John Doe", "100", "100 Pine Ave")
	mustInsertCustomer(t, db, "Jane Doe", "101", "101 Pine Ave")
	mustInsertCustomer(t, db, "Bob Smith", "200", "200 Elm St")

	t.Run("search by name", func(t *testing.T) {
		results, err := cm.Search("Doe")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 2 {
			t.Errorf("got %d results, want 2", len(results))
		}
	})

	t.Run("search by house number", func(t *testing.T) {
		results, err := cm.Search("100")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 1 || results[0].Name != "John Doe" {
			t.Errorf("got %v, want [John Doe]", results)
		}
	})

	t.Run("search by address", func(t *testing.T) {
		results, err := cm.Search("Elm")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 1 || results[0].Name != "Bob Smith" {
			t.Errorf("got %v, want [Bob Smith]", results)
		}
	})

	t.Run("no match returns empty", func(t *testing.T) {
		results, err := cm.Search("zzznomatch")
		if err != nil {
			t.Fatalf("Search: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("got %d results, want 0", len(results))
		}
	})
}

func TestCustomerUpdate(t *testing.T) {
	db := newTestDB(t)
	cm := &CustomerModel{DB: db}

	id := mustInsertCustomer(t, db, "Old Name", "50", "50 Old Rd")

	if err := cm.Update(id, "New Name", "555-9999", "51", "51 New Rd"); err != nil {
		t.Fatalf("Update: %v", err)
	}

	c, err := cm.Get(id)
	if err != nil {
		t.Fatalf("Get after Update: %v", err)
	}
	if c.Name != "New Name" || c.Phone != "555-9999" || c.HouseNumber != "51" || c.Address != "51 New Rd" {
		t.Errorf("unexpected fields after update: %+v", c)
	}

	if err := cm.Update(id+999, "Ghost", "", "0", ""); !errors.Is(err, ErrNoRecord) {
		t.Errorf("got error %v, want ErrNoRecord for missing customer", err)
	}
}
