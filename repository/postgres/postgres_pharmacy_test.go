// postgres_pharmacy_repository_test.go
package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/stretchr/testify/assert"
)

func uniqueName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func TestPharmacyRepository_CreateAndGetDrug(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph, ok := repo.(*postgresPharmacyRepository)
	if !ok {
		t.Fatal("failed to assert repo")
	}

	name := uniqueName("Paracetamol")
	in := &entities.Drug{
		Name:        name,
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
		Notes:       entities.PtrTo("pain relief"),
	}

	created, err := ph.CreateDrug(context.Background(), in)
	assert.Nil(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, name, created.Name)
	assert.Equal(t, "tablet", created.Unit)
	assert.Equal(t, 1, *created.DefaultSize)
	assert.Equal(t, "pain relief", *created.Notes)

	// Get by id
	got, err := ph.GetDrug(context.Background(), created.ID)
	assert.Nil(t, err)
	assert.Equal(t, created, got)
}

func TestPharmacyRepository_CreateDrug_DuplicateName(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)

	name := uniqueName("Amoxicillin")
	in := &entities.Drug{Name: name, Unit: "capsule", DefaultSize: entities.PtrTo(1), Notes: nil}

	_, err := ph.CreateDrug(context.Background(), in)
	assert.Nil(t, err)

	// Duplicate name → ErrDrugNameTaken
	_, err = ph.CreateDrug(context.Background(), in)
	assert.ErrorIs(t, err, entities.ErrDrugNameTaken)
}

func TestPharmacyRepository_UpdateDrug_SuccessAndNotFound(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)

	// Create a drug to update
	name := uniqueName("Ibuprofen")
	created, err := ph.CreateDrug(context.Background(), &entities.Drug{
		Name:        name,
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(2),
		Notes:       entities.PtrTo("anti-inflammatory"),
	})
	assert.Nil(t, err)

	// Update it
	created.Name = created.Name + "-Updated"
	created.Unit = "syrup"
	created.DefaultSize = entities.PtrTo(5)
	created.Notes = entities.PtrTo("updated notes")
	updated, err := ph.UpdateDrug(context.Background(), created)
	assert.Nil(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "syrup", updated.Unit)
	assert.Equal(t, 5, *updated.DefaultSize)
	assert.Equal(t, "updated notes", *updated.Notes)

	// Update non-existent id
	_, err = ph.UpdateDrug(context.Background(), &entities.Drug{
		ID:          9_999_999,
		Name:        "X",
		Unit:        "x",
		DefaultSize: entities.PtrTo(1),
		Notes:       nil,
	})
	assert.EqualError(t, err, "drug not found")
}

func TestPharmacyRepository_DeleteDrug_SuccessAndNotFound(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)

	// Create a drug to delete
	name := uniqueName("Cetrizine")
	created, err := ph.CreateDrug(context.Background(), &entities.Drug{
		Name:        name,
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
		Notes:       nil,
	})
	assert.Nil(t, err)

	// Delete it
	err = ph.DeleteDrug(context.Background(), created.ID)
	assert.Nil(t, err)

	// Ensure it's gone
	_, err = ph.GetDrug(context.Background(), created.ID)
	assert.EqualError(t, err, "drug not found")

	// Delete non-existent id
	err = ph.DeleteDrug(context.Background(), 9_999_999)
	assert.EqualError(t, err, "drug not found")
}

func TestPharmacyRepository_ListDrugs_ContainsCreatedOnes(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)

	aName := uniqueName("DrugA")
	bName := uniqueName("DrugB")
	_, _ = ph.CreateDrug(context.Background(), &entities.Drug{Name: aName, Unit: "tab", DefaultSize: entities.PtrTo(1), Notes: nil})
	_, _ = ph.CreateDrug(context.Background(), &entities.Drug{Name: bName, Unit: "tab", DefaultSize: entities.PtrTo(1), Notes: nil})

	list, err := ph.ListDrugs(context.Background())
	assert.Nil(t, err)

	// Just verify that our two uniques are present (list may contain seeded rows).
	foundA := false
	foundB := false
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

func TestPharmacyRepository_BatchCRUDAndList(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)

	// Create a parent drug
	name := uniqueName("Metformin")
	d, err := ph.CreateDrug(context.Background(), &entities.Drug{
		Name:        name,
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
		Notes:       nil,
	})
	assert.Nil(t, err)

	// Create three batches with different expiries
	b1 := entities.DrugBatch{
		DrugID:      d.ID,
		BatchNumber: "B001",
		Location:    "Main",
		Quantity:    100,
		ExpiryDate:  time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Supplier:    entities.PtrTo("ACME"),
	}
	b2 := entities.DrugBatch{
		DrugID:      d.ID,
		BatchNumber: "B002",
		Location:    "Main",
		Quantity:    50,
		ExpiryDate:  time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		Supplier:    entities.PtrTo("ACME"),
	}
	b3 := entities.DrugBatch{
		DrugID:      d.ID,
		BatchNumber: "B003",
		Location:    "Overflow",
		Quantity:    200,
		ExpiryDate:  time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC),
		Supplier:    entities.PtrTo("ACME"),
	}

	var err1, err2, err3 error
	b1.ID, err1 = ph.CreateBatch(context.Background(), &b1)
	b2.ID, err2 = ph.CreateBatch(context.Background(), &b2)
	b3.ID, err3 = ph.CreateBatch(context.Background(), &b3)
	assert.Nil(t, err1)
	assert.Nil(t, err2)
	assert.Nil(t, err3)

	// List filtered by drug → should be ordered by expiry_date
	list, err := ph.ListBatches(context.Background(), &d.ID)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(list), 3)
	// pick the last 3 for order check in case of seeds
	var last3 []entities.DrugBatch
	for _, x := range list {
		if x.DrugID == d.ID {
			last3 = append(last3, x)
		}
	}
	assert.Equal(t, 3, len(last3))
	assert.True(t, last3[0].ExpiryDate.Before(last3[1].ExpiryDate) || last3[0].ExpiryDate.Equal(last3[1].ExpiryDate))
	assert.True(t, last3[1].ExpiryDate.Before(last3[2].ExpiryDate) || last3[1].ExpiryDate.Equal(last3[2].ExpiryDate))

	// Update a batch
	b1.Location = "Cabinet A"
	b1.Quantity = 80
	err = ph.UpdateBatch(context.Background(), &b1)
	assert.Nil(t, err)

	// Update non-existent batch
	err = ph.UpdateBatch(context.Background(), &entities.DrugBatch{
		ID:          9_999_999,
		DrugID:      d.ID,
		BatchNumber: "X",
		Location:    "X",
		Quantity:    1,
		ExpiryDate:  time.Now(),
		Supplier:    entities.PtrTo("X"),
	})
	assert.EqualError(t, err, "batch not found")

	// Delete a batch
	err = ph.DeleteBatch(context.Background(), b3.ID)
	assert.Nil(t, err)

	// Delete non-existent
	err = ph.DeleteBatch(context.Background(), 9_999_999)
	assert.EqualError(t, err, "batch not found")
}

func TestPharmacyRepository_EarliestBatches_FEFOHelper(t *testing.T) {
	repo := NewPostgresPharmacyRepository(db)
	ph := repo.(*postgresPharmacyRepository)

	// Create a parent drug
	name := uniqueName("Loratadine")
	d, err := ph.CreateDrug(context.Background(), &entities.Drug{
		Name:        name,
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
		Notes:       nil,
	})
	assert.Nil(t, err)

	// Mix of quantities (including zero) and dates
	bEarly := entities.DrugBatch{
		DrugID:      d.ID,
		BatchNumber: "E1",
		Location:    "Shelf 1",
		Quantity:    5,
		ExpiryDate:  time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		Supplier:    entities.PtrTo("S"),
	}
	bZero := entities.DrugBatch{
		DrugID:      d.ID,
		BatchNumber: "Z0",
		Location:    "Shelf 2",
		Quantity:    0, // should be filtered out by earliestBatches
		ExpiryDate:  time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC),
		Supplier:    entities.PtrTo("S"),
	}
	bLater := entities.DrugBatch{
		DrugID:      d.ID,
		BatchNumber: "L2",
		Location:    "Shelf 3",
		Quantity:    10,
		ExpiryDate:  time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		Supplier:    entities.PtrTo("S"),
	}
	var e1, e2, e3 error
	bEarly.ID, e1 = ph.CreateBatch(context.Background(), &bEarly)
	bZero.ID, e2 = ph.CreateBatch(context.Background(), &bZero)
	bLater.ID, e3 = ph.CreateBatch(context.Background(), &bLater)
	assert.Nil(t, e1)
	assert.Nil(t, e2)
	assert.Nil(t, e3)

	// Call the FEFO helper directly (it's in the same package)
	batches, err := ph.earliestBatches(context.Background(), d.ID)
	assert.Nil(t, err)

	// Should exclude zero-quantity and be sorted by expiry_date
	for _, b := range batches {
		assert.NotEqual(t, 0, b.Quantity, "zero-quantity batch should be excluded")
	}
	// Find our two non-zero entries
	var seen []entities.DrugBatch
	for _, b := range batches {
		if b.BatchNumber == "E1" || b.BatchNumber == "L2" {
			seen = append(seen, b)
		}
	}
	assert.Equal(t, 2, len(seen))
	assert.True(t, seen[0].ExpiryDate.Before(seen[1].ExpiryDate) || seen[0].ExpiryDate.Equal(seen[1].ExpiryDate))
}
