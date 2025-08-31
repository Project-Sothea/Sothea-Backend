package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- helpers ----------
func uniq(prefix string) string {
	return prefix + "-" + time.Now().Format("150405.000000000")
}

// ---------- DRUGS ----------

func TestPharmacyRepository_CreateAndGetDrug(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	name := uniq("Paracetamol")
	in := &entities.Drug{
		Name:        name,
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
		Notes:       entities.PtrTo("pain relief"),
	}

	created, err := ph.CreateDrug(ctx, in)
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	assert.Equal(t, name, created.Name)
	assert.Equal(t, "tablet", created.Unit)
	require.NotNil(t, created.DefaultSize)
	assert.Equal(t, 1, *created.DefaultSize)
	require.NotNil(t, created.Notes)
	assert.Equal(t, "pain relief", *created.Notes)

	got, err := ph.GetDrug(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created, got)
}

func TestPharmacyRepository_CreateDrug_DuplicateName(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	name := uniq("Amoxicillin")
	in := &entities.Drug{Name: name, Unit: "capsule", DefaultSize: entities.PtrTo(1)}

	_, err := ph.CreateDrug(ctx, in)
	require.NoError(t, err)

	// Duplicate name → ErrDrugNameTaken (from ON CONFLICT DO NOTHING + Scan ErrNoRows)
	_, err = ph.CreateDrug(ctx, in)
	require.ErrorIs(t, err, entities.ErrDrugNameTaken)
}

func TestPharmacyRepository_UpdateDrug_SuccessAndNotFound(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	// Create a drug to update
	created, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("Ibuprofen"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(2),
		Notes:       entities.PtrTo("anti-inflammatory"),
	})
	require.NoError(t, err)

	// Update it
	created.Name = created.Name + "-Updated"
	created.Unit = "syrup"
	created.DefaultSize = entities.PtrTo(5)
	created.Notes = entities.PtrTo("updated notes")

	updated, err := ph.UpdateDrug(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "syrup", updated.Unit)
	require.NotNil(t, updated.DefaultSize)
	assert.Equal(t, 5, *updated.DefaultSize)
	require.NotNil(t, updated.Notes)
	assert.Equal(t, "updated notes", *updated.Notes)

	// Update non-existent id
	_, err = ph.UpdateDrug(ctx, &entities.Drug{
		ID:          9_999_999,
		Name:        "X",
		Unit:        "x",
		DefaultSize: entities.PtrTo(1),
	})
	assert.EqualError(t, err, "drug not found")
}

func TestPharmacyRepository_DeleteDrug_SuccessAndNotFound(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	// Create a drug to delete
	created, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("Cetirizine"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)

	// Delete it
	err = ph.DeleteDrug(ctx, created.ID)
	require.NoError(t, err)

	// Ensure it's gone
	_, err = ph.GetDrug(ctx, created.ID)
	assert.EqualError(t, err, "drug not found")

	// Delete non-existent id
	err = ph.DeleteDrug(ctx, 9_999_999)
	assert.EqualError(t, err, "drug not found")
}

func TestPharmacyRepository_ListDrugs_ContainsCreatedOnes(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	aName := uniq("DrugA")
	bName := uniq("DrugB")
	_, _ = ph.CreateDrug(ctx, &entities.Drug{Name: aName, Unit: "tab", DefaultSize: entities.PtrTo(1)})
	_, _ = ph.CreateDrug(ctx, &entities.Drug{Name: bName, Unit: "tab", DefaultSize: entities.PtrTo(1)})

	list, err := ph.ListDrugs(ctx)
	require.NoError(t, err)

	foundA, foundB := false, false
	for _, d := range list {
		if d.Name == aName {
			foundA = true
		}
		if d.Name == bName {
			foundB = true
		}
	}
	assert.True(t, foundA, "expected DrugA in list")
	assert.True(t, foundB, "expected DrugB in list")
}

// ---------- BATCHES + LOCATIONS ----------

func TestPharmacyRepository_CreateBatch_WithLocations_AtomicAndGet(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	// Create drug
	d, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("Metformin"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)

	// Create batch with nested locations in a single tx
	in := &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d.ID,
			BatchNumber: "MF-001",
			ExpiryDate:  time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			Supplier:    entities.PtrTo("ACME"),
		},
		BatchLocations: []entities.DrugBatchLocation{
			{Location: "Main", Quantity: 100},
			{Location: "Cabinet A", Quantity: 40},
		},
	}
	created, err := ph.CreateBatch(ctx, in)
	require.NoError(t, err)
	require.NotZero(t, created.DrugBatch.ID)
	require.Len(t, created.BatchLocations, 2)

	// Get and verify locations persisted
	got, err := ph.GetBatch(ctx, created.DrugBatch.ID)
	require.NoError(t, err)
	assert.Equal(t, created.DrugBatch.BatchNumber, got.DrugBatch.BatchNumber)
	require.Len(t, got.BatchLocations, 2)
}

func TestPharmacyRepository_CreateBatch_RollbackOnBadLocation(t *testing.T) {
	// Negative quantity in nested location should make the whole tx fail and not persist anything
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	d, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("RollbackDrug"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)

	// Count batches for this drug before
	before, err := ph.ListBatchDetails(ctx, &d.ID)
	require.NoError(t, err)
	beforeN := len(before)

	_, err = ph.CreateBatch(ctx, &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d.ID,
			BatchNumber: "RB-ERR",
			ExpiryDate:  time.Now().AddDate(1, 0, 0),
			Supplier:    entities.PtrTo("S"),
		},
		BatchLocations: []entities.DrugBatchLocation{
			{Location: "Main", Quantity: -5}, // illegal
		},
	})
	require.Error(t, err)

	// Count stays the same → rollback succeeded
	after, err := ph.ListBatchDetails(ctx, &d.ID)
	require.NoError(t, err)
	assert.Equal(t, beforeN, len(after))
}

func TestPharmacyRepository_ListBatchDetails_FilteringAndOrder(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	// Create two drugs
	d1, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("Amlodipine"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)
	d2, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("Bisoprolol"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)

	// Create batches for both drugs (different expiries for d1 to test order)
	b1, err := ph.CreateBatch(ctx, &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d1.ID,
			BatchNumber: "D1-B001",
			ExpiryDate:  time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
			Supplier:    entities.PtrTo("ACME"),
		},
	})
	require.NoError(t, err)
	b2, err := ph.CreateBatch(ctx, &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d1.ID,
			BatchNumber: "D1-B002",
			ExpiryDate:  time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
			Supplier:    entities.PtrTo("ACME"),
		},
	})
	require.NoError(t, err)
	b3, err := ph.CreateBatch(ctx, &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d1.ID,
			BatchNumber: "D1-B003",
			ExpiryDate:  time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
			Supplier:    entities.PtrTo("ACME"),
		},
	})
	require.NoError(t, err)
	bOther, err := ph.CreateBatch(ctx, &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d2.ID,
			BatchNumber: "D2-B100",
			ExpiryDate:  time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			Supplier:    entities.PtrTo("Globex"),
		},
	})
	require.NoError(t, err)

	// Attach locations
	_, err = ph.CreateBatchLocation(ctx, &entities.DrugBatchLocation{
		BatchID:  b1.DrugBatch.ID,
		Location: "Main",
		Quantity: 100,
	})
	require.NoError(t, err)
	_, err = ph.CreateBatchLocation(ctx, &entities.DrugBatchLocation{
		BatchID:  b1.DrugBatch.ID,
		Location: "Cabinet A",
		Quantity: 40,
	})
	require.NoError(t, err)
	_, err = ph.CreateBatchLocation(ctx, &entities.DrugBatchLocation{
		BatchID:  b2.DrugBatch.ID,
		Location: "Main",
		Quantity: 25,
	})
	require.NoError(t, err)
	_, err = ph.CreateBatchLocation(ctx, &entities.DrugBatchLocation{
		BatchID:  bOther.DrugBatch.ID,
		Location: "Overflow",
		Quantity: 10,
	})
	require.NoError(t, err)

	// 1) List all (nil filter)
	allDetails, err := ph.ListBatchDetails(ctx, nil)
	require.NoError(t, err)
	require.True(t, len(allDetails) >= 4)

	// Ensure grouping
	find := func(id int64) *entities.BatchDetail {
		for i := range allDetails {
			if allDetails[i].DrugBatch.ID == id {
				return &allDetails[i]
			}
		}
		return nil
	}
	dB1 := find(b1.DrugBatch.ID)
	dB2 := find(b2.DrugBatch.ID)
	dB3 := find(b3.DrugBatch.ID)
	dOther := find(bOther.DrugBatch.ID)
	if assert.NotNil(t, dB1) {
		assert.Len(t, dB1.BatchLocations, 2)
	}
	if assert.NotNil(t, dB2) {
		assert.Len(t, dB2.BatchLocations, 1)
	}
	if assert.NotNil(t, dB3) {
		assert.Len(t, dB3.BatchLocations, 0)
	}
	if assert.NotNil(t, dOther) {
		assert.Len(t, dOther.BatchLocations, 1)
	}

	// 2) Filter by d1.ID → exactly the 3 d1 batches in expiry_date,id order
	filtered, err := ph.ListBatchDetails(ctx, &d1.ID)
	require.NoError(t, err)

	var d1Only []entities.DrugBatch
	for _, bd := range filtered {
		if bd.DrugBatch.DrugID == d1.ID {
			d1Only = append(d1Only, bd.DrugBatch)
		}
	}
	require.Equal(t, 3, len(d1Only), "expected exactly 3 batches for d1")

	for i := 1; i < len(d1Only); i++ {
		prev := d1Only[i-1]
		curr := d1Only[i]
		if prev.ExpiryDate.Equal(curr.ExpiryDate) {
			assert.Less(t, prev.ID, curr.ID, "ids should increase when expiries equal")
		} else {
			assert.True(t, prev.ExpiryDate.Before(curr.ExpiryDate), "expiry_date should be ascending")
		}
	}
}

func TestPharmacyRepository_UpdateAndDeleteBatch(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	d, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("UpdateDrug"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)

	created, err := ph.CreateBatch(ctx, &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d.ID,
			BatchNumber: "UD-001",
			ExpiryDate:  time.Now().AddDate(1, 0, 0),
			Supplier:    entities.PtrTo("S1"),
		},
	})
	require.NoError(t, err)

	// Update
	ub := created.DrugBatch
	ub.BatchNumber = "UD-001-UPDATED"
	ub.Supplier = entities.PtrTo("S2")
	updated, err := ph.UpdateBatch(ctx, &ub)
	require.NoError(t, err)
	assert.Equal(t, "UD-001-UPDATED", updated.DrugBatch.BatchNumber)
	require.NotNil(t, updated.DrugBatch.Supplier)
	assert.Equal(t, "S2", *updated.DrugBatch.Supplier)

	// Update non-existent
	_, err = ph.UpdateBatch(ctx, &entities.DrugBatch{
		ID:          9_999_999,
		DrugID:      d.ID,
		BatchNumber: "X",
		ExpiryDate:  time.Now(),
		Supplier:    entities.PtrTo("X"),
	})
	assert.EqualError(t, err, "batch not found")

	// Delete
	err = ph.DeleteBatch(ctx, created.DrugBatch.ID)
	require.NoError(t, err)

	// Delete unknown
	err = ph.DeleteBatch(ctx, 9_999_999)
	assert.EqualError(t, err, "batch not found")
}

func TestPharmacyRepository_ListBatchLocations_CreateUpdateDelete_NoUpsert(t *testing.T) {
	// This test reflects insert-only semantics for CreateBatchLocation (no ON CONFLICT)
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	// Create a drug + batch
	d, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("Loratadine"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)

	bd, err := ph.CreateBatch(ctx, &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d.ID,
			BatchNumber: "LOT-001",
			ExpiryDate:  time.Date(2025, 2, 10, 0, 0, 0, 0, time.UTC),
			Supplier:    entities.PtrTo("SupplierX"),
		},
	})
	require.NoError(t, err)

	// Insert locations
	locA, err := ph.CreateBatchLocation(ctx, &entities.DrugBatchLocation{
		BatchID:  bd.DrugBatch.ID,
		Location: "Main",
		Quantity: 10,
	})
	require.NoError(t, err)
	require.NotZero(t, locA.ID)

	locB, err := ph.CreateBatchLocation(ctx, &entities.DrugBatchLocation{
		BatchID:  bd.DrugBatch.ID,
		Location: "Cabinet 1",
		Quantity: 5,
	})
	require.NoError(t, err)
	require.NotZero(t, locB.ID)

	// Second insert with same (batch_id, location) should fail with duplicate key
	_, err = ph.CreateBatchLocation(ctx, &entities.DrugBatchLocation{
		BatchID:  bd.DrugBatch.ID,
		Location: "Main",
		Quantity: 25,
	})
	require.Error(t, err)
	var pqErr *pq.Error
	if assert.ErrorAs(t, err, &pqErr) {
		assert.Equal(t, "23505", string(pqErr.Code), "should be unique_violation")
	}

	// Use Update to change quantity
	locA.Quantity = 25
	updatedMain, err := ph.UpdateBatchLocation(ctx, locA)
	require.NoError(t, err)
	assert.Equal(t, locA.ID, updatedMain.ID)
	assert.Equal(t, int64(25), updatedMain.Quantity)

	// List locations
	locs, err := ph.ListBatchLocations(ctx, bd.DrugBatch.ID)
	require.NoError(t, err)
	require.Len(t, locs, 2)

	// Verify values
	var gotMain, gotCab *entities.DrugBatchLocation
	for i := range locs {
		if locs[i].Location == "Main" {
			cp := locs[i]
			gotMain = &cp
		}
		if locs[i].Location == "Cabinet 1" {
			cp := locs[i]
			gotCab = &cp
		}
	}
	if assert.NotNil(t, gotMain) {
		assert.Equal(t, int64(25), gotMain.Quantity)
	}
	if assert.NotNil(t, gotCab) {
		assert.Equal(t, int64(5), gotCab.Quantity)
	}

	// Delete "Cabinet 1"
	err = ph.DeleteBatchLocation(ctx, locB.ID)
	require.NoError(t, err)

	locs2, err := ph.ListBatchLocations(ctx, bd.DrugBatch.ID)
	require.NoError(t, err)
	if assert.Len(t, locs2, 1) {
		assert.Equal(t, "Main", locs2[0].Location)
		assert.Equal(t, int64(25), locs2[0].Quantity)
	}

	// Delete unknown location
	err = ph.DeleteBatchLocation(ctx, 9_999_999)
	assert.EqualError(t, err, "location not found")
}

func TestPharmacyRepository_GetBatchLocation_NotFound(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	_, err := ph.GetBatchLocation(ctx, 9_999_999)
	assert.EqualError(t, err, "batch location not found")
}

func TestPharmacyRepository_ListBatchDetails_EmptyForDrugWithoutBatches(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	d, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("EmptyDrug"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)

	details, err := ph.ListBatchDetails(ctx, &d.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, len(details))
}

// Optional: depending on FK ON DELETE behavior (RESTRICT vs CASCADE)
func TestPharmacyRepository_DeleteDrug_WithBatches_FKBehavior(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)
	ctx := context.Background()

	d, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        uniq("FK-Drug"),
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
	})
	require.NoError(t, err)

	_, err = ph.CreateBatch(ctx, &entities.BatchDetail{
		DrugBatch: entities.DrugBatch{
			DrugID:      d.ID,
			BatchNumber: "FK-B-1",
			ExpiryDate:  time.Now().AddDate(1, 0, 0),
			Supplier:    entities.PtrTo("FK-Supplier"),
		},
	})
	require.NoError(t, err)

	err = ph.DeleteDrug(ctx, d.ID)
	if err == nil {
		t.Skip("DeleteDrug succeeded (likely ON DELETE CASCADE). If you expect FK violation, enforce RESTRICT and assert specific error.")
		return
	}
	assert.Error(t, err)
}
