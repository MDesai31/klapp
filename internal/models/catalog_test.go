package models

import "testing"

func TestCatalogJobs(t *testing.T) {
	db := newTestDB(t)
	cm := &CatalogModel{DB: db}

	t.Run("CreateJob and ListJobs", func(t *testing.T) {
		if _, err := cm.CreateJob("Mow lawn"); err != nil {
			t.Fatalf("CreateJob: %v", err)
		}
		if _, err := cm.CreateJob("Trim hedges"); err != nil {
			t.Fatalf("CreateJob: %v", err)
		}

		jobs, err := cm.ListJobs()
		if err != nil {
			t.Fatalf("ListJobs: %v", err)
		}
		if len(jobs) != 2 {
			t.Fatalf("got %d jobs, want 2", len(jobs))
		}
		// Ordered by description
		if jobs[0].Description != "Mow lawn" || jobs[1].Description != "Trim hedges" {
			t.Errorf("unexpected order: %v", jobs)
		}
	})

	t.Run("CreateJob ignores duplicates", func(t *testing.T) {
		if _, err := cm.CreateJob("Mow lawn"); err != nil {
			t.Fatalf("duplicate CreateJob should not error: %v", err)
		}
		jobs, err := cm.ListJobs()
		if err != nil {
			t.Fatalf("ListJobs: %v", err)
		}
		if len(jobs) != 2 {
			t.Errorf("got %d jobs after duplicate insert, want 2", len(jobs))
		}
	})

	t.Run("SearchJobs", func(t *testing.T) {
		results, err := cm.SearchJobs("mow")
		if err != nil {
			t.Fatalf("SearchJobs: %v", err)
		}
		if len(results) != 1 || results[0] != "Mow lawn" {
			t.Errorf("got %v, want [Mow lawn]", results)
		}

		results, err = cm.SearchJobs("zzznomatch")
		if err != nil {
			t.Fatalf("SearchJobs: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("got %d results, want 0", len(results))
		}
	})

	t.Run("DeleteJob", func(t *testing.T) {
		jobs, _ := cm.ListJobs()
		idToDelete := jobs[0].ID

		if err := cm.DeleteJob(idToDelete); err != nil {
			t.Fatalf("DeleteJob: %v", err)
		}

		jobs, err := cm.ListJobs()
		if err != nil {
			t.Fatalf("ListJobs after delete: %v", err)
		}
		if len(jobs) != 1 {
			t.Errorf("got %d jobs after delete, want 1", len(jobs))
		}
		for _, j := range jobs {
			if j.ID == idToDelete {
				t.Errorf("deleted job still present: %+v", j)
			}
		}
	})
}

func TestCatalogMaterials(t *testing.T) {
	db := newTestDB(t)
	cm := &CatalogModel{DB: db}

	t.Run("CreateMaterial and ListMaterials", func(t *testing.T) {
		if _, err := cm.CreateMaterial("Fertilizer", "bag", 25.00); err != nil {
			t.Fatalf("CreateMaterial: %v", err)
		}
		if _, err := cm.CreateMaterial("Mulch", "yard", 40.00); err != nil {
			t.Fatalf("CreateMaterial: %v", err)
		}

		mats, err := cm.ListMaterials()
		if err != nil {
			t.Fatalf("ListMaterials: %v", err)
		}
		if len(mats) != 2 {
			t.Fatalf("got %d materials, want 2", len(mats))
		}
		// Ordered by name
		if mats[0].Name != "Fertilizer" || mats[1].Name != "Mulch" {
			t.Errorf("unexpected order: %v", mats)
		}
		if mats[0].Unit != "bag" || mats[0].Price != 25.00 {
			t.Errorf("unexpected fields: %+v", mats[0])
		}
	})

	t.Run("CreateMaterial ignores duplicates", func(t *testing.T) {
		if _, err := cm.CreateMaterial("Fertilizer", "bag", 25.00); err != nil {
			t.Fatalf("duplicate CreateMaterial should not error: %v", err)
		}
		mats, _ := cm.ListMaterials()
		if len(mats) != 2 {
			t.Errorf("got %d materials after duplicate insert, want 2", len(mats))
		}
	})

	t.Run("SearchMaterials", func(t *testing.T) {
		results, err := cm.SearchMaterials("fert")
		if err != nil {
			t.Fatalf("SearchMaterials: %v", err)
		}
		if len(results) != 1 || results[0] != "Fertilizer" {
			t.Errorf("got %v, want [Fertilizer]", results)
		}

		results, err = cm.SearchMaterials("zzznomatch")
		if err != nil {
			t.Fatalf("SearchMaterials: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("got %d results, want 0", len(results))
		}
	})

	t.Run("UpdateMaterial", func(t *testing.T) {
		mats, _ := cm.ListMaterials()
		id := mats[0].ID

		if err := cm.UpdateMaterial(id, "Fertilizer Pro", "lb", 30.00); err != nil {
			t.Fatalf("UpdateMaterial: %v", err)
		}

		mats, err := cm.ListMaterials()
		if err != nil {
			t.Fatalf("ListMaterials after update: %v", err)
		}
		var found bool
		for _, m := range mats {
			if m.ID == id {
				found = true
				if m.Name != "Fertilizer Pro" || m.Unit != "lb" || m.Price != 30.00 {
					t.Errorf("unexpected fields after update: %+v", m)
				}
			}
		}
		if !found {
			t.Error("updated material not found in list")
		}
	})

	t.Run("UpdateMaterial returns ErrNoRecord for missing ID", func(t *testing.T) {
		if err := cm.UpdateMaterial(99999, "Ghost", "", 0); !isErrNoRecord(err) {
			t.Errorf("got error %v, want ErrNoRecord", err)
		}
	})

	t.Run("DeleteMaterial", func(t *testing.T) {
		mats, _ := cm.ListMaterials()
		idToDelete := mats[0].ID

		if err := cm.DeleteMaterial(idToDelete); err != nil {
			t.Fatalf("DeleteMaterial: %v", err)
		}

		mats, err := cm.ListMaterials()
		if err != nil {
			t.Fatalf("ListMaterials after delete: %v", err)
		}
		for _, m := range mats {
			if m.ID == idToDelete {
				t.Errorf("deleted material still present: %+v", m)
			}
		}
	})
}

// isErrNoRecord is a helper because UpdateMaterial returns ErrNoRecord
// via RowsAffected, which doesn't wrap via errors.Is by default.
func isErrNoRecord(err error) bool {
	return err == ErrNoRecord
}
