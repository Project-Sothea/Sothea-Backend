// postgres_prescription_repository_test.go
package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/stretchr/testify/assert"
)

// ---------- helpers ----------

func unique(prefix string) string { return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()) }

func mustCreatePatientAndVisit(t *testing.T) (int32, int32) {
	t.Helper()
	ctx := context.Background()
	patRepo := NewPostgresPatientRepository(db).(*postgresPatientRepository)

	now := time.Now().UTC().Truncate(time.Second)
	admin := entities.Admin{
		FamilyGroup:         entities.PtrTo(unique("FG")),
		RegDate:             entities.PtrTo(now),
		QueueNo:             entities.PtrTo("Q1"),
		Name:                entities.PtrTo(unique("John Doe")),
		KhmerName:           entities.PtrTo("ខ្មែរ"),
		Dob:                 entities.PtrTo(now.AddDate(-30, 0, 0)),
		Age:                 entities.PtrTo(30),
		Gender:              entities.PtrTo("M"),
		Village:             entities.PtrTo("VillageX"),
		ContactNo:           entities.PtrTo("12345678"),
		Pregnant:            entities.PtrTo(false),
		LastMenstrualPeriod: nil,
		DrugAllergies:       entities.PtrTo("none"),
		SentToID:            entities.PtrTo(false),
		Photo:               nil,
	}

	id, err := patRepo.CreatePatient(ctx, &admin)
	assert.Nil(t, err)

	vid, err := patRepo.CreatePatientVisit(ctx, id, &admin)
	assert.Nil(t, err)

	return id, vid
}

func mustCreateDrug(t *testing.T, name string) *entities.Drug {
	t.Helper()
	ctx := context.Background()
	ph := NewPostgresPharmacyRepository(db).(*postgresPharmacyRepository)

	d, err := ph.CreateDrug(ctx, &entities.Drug{
		Name:        name,
		Unit:        "tablet",
		DefaultSize: entities.PtrTo(1),
		Notes:       entities.PtrTo(""),
	})
	assert.Nil(t, err)
	return d
}

// Creates a batch row and a location row with quantity, returns (batchID, locationID)
func mustCreateBatchAndLocation(t *testing.T, drugID int64, batchNo string, qty int, expiry time.Time, location string) (int64, int64) {
	t.Helper()
	ctx := context.Background()

	var batchID int64
	err := db.QueryRowContext(ctx, `
		INSERT INTO drug_batches (drug_id, batch_number, expiry_date, supplier)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, drugID, batchNo, expiry, "ACME").Scan(&batchID)
	assert.Nil(t, err)

	var locID int64
	err = db.QueryRowContext(ctx, `
		INSERT INTO batch_locations (batch_id, location, quantity)
		VALUES ($1, $2, $3)
		RETURNING id
	`, batchID, location, qty).Scan(&locID)
	assert.Nil(t, err)

	return batchID, locID
}

func getLocationQty(t *testing.T, locationID int64) int64 {
	t.Helper()
	var q int64
	err := db.QueryRow(`SELECT quantity FROM batch_locations WHERE id=$1`, locationID).Scan(&q)
	assert.Nil(t, err)
	return q
}

// ---------- tests ----------

func TestPrescription_CreateAndGet_ReducesStockAndHydrates(t *testing.T) {
	ctx := context.Background()
	id, vid := mustCreatePatientAndVisit(t)

	// stock
	drug := mustCreateDrug(t, unique("Paracetamol"))
	_, loc1 := mustCreateBatchAndLocation(t, drug.ID, "P-B1", 100, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "Main")
	_, loc2 := mustCreateBatchAndLocation(t, drug.ID, "P-B2", 50, time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC), "Main")

	before1 := getLocationQty(t, loc1)
	before2 := getLocationQty(t, loc2)

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	p := &entities.Prescription{
		PatientID: int64(id),
		VID:       vid,
		Notes:     entities.PtrTo("Take after meals"),
		PrescribedDrugs: []entities.DrugPrescription{
			{
				DrugID:  drug.ID,
				Remarks: entities.PtrTo("q8h"),
				Batches: []entities.PrescriptionBatchItem{
					{BatchLocationId: loc1, Quantity: 40},
					{BatchLocationId: loc2, Quantity: 30},
				},
			},
		},
	}

	created, err := repo.CreatePrescription(ctx, p)
	assert.Nil(t, err)
	assert.NotZero(t, created.ID)
	assert.Equal(t, int64(id), created.PatientID)
	assert.Equal(t, vid, created.VID)
	assert.Equal(t, 1, len(created.PrescribedDrugs))
	assert.Equal(t, 2, len(created.PrescribedDrugs[0].Batches))

	// stock reduced
	after1 := getLocationQty(t, loc1)
	after2 := getLocationQty(t, loc2)
	assert.Equal(t, before1-40, after1)
	assert.Equal(t, before2-30, after2)

	// hydrate via Get
	got, err := repo.GetPrescriptionByID(ctx, created.ID)
	assert.Nil(t, err)
	assert.Equal(t, created.ID, got.ID)
	assert.Equal(t, created.PatientID, got.PatientID)
	assert.Equal(t, 1, len(got.PrescribedDrugs))
	assert.Equal(t, 2, len(got.PrescribedDrugs[0].Batches))
}

func TestPrescription_Create_FailsWhenInsufficientStock(t *testing.T) {
	ctx := context.Background()
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Amoxicillin"))
	_, loc := mustCreateBatchAndLocation(t, drug.ID, "A-B1", 5, time.Now().AddDate(0, 6, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)
	_, err := repo.CreatePrescription(ctx, &entities.Prescription{
		PatientID: int64(id),
		VID:       vid,
		Notes:     entities.PtrTo("Too much"),
		PrescribedDrugs: []entities.DrugPrescription{
			{
				DrugID:  drug.ID,
				Remarks: nil,
				Batches: []entities.PrescriptionBatchItem{{BatchLocationId: loc, Quantity: 10}}, // exceeds 5
			},
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient stock in batch-location")
}

func TestPrescription_List_WithFiltersAndOrder(t *testing.T) {
	ctx := context.Background()
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Ibuprofen"))
	_, loc := mustCreateBatchAndLocation(t, drug.ID, "I-B1", 100, time.Now().AddDate(0, 12, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	// make two prescriptions for the same (id,vid)
	make := func(note string, q int) int64 {
		p, err := repo.CreatePrescription(ctx, &entities.Prescription{
			PatientID: int64(id),
			VID:       vid,
			Notes:     &note,
			PrescribedDrugs: []entities.DrugPrescription{
				{DrugID: drug.ID, Remarks: nil, Batches: []entities.PrescriptionBatchItem{{BatchLocationId: loc, Quantity: q}}},
			},
		})
		assert.Nil(t, err)
		return p.ID
	}
	firstID := make("first", 10)
	_ = firstID
	time.Sleep(5 * time.Millisecond) // ensure created_at ordering
	secondID := make("second", 5)
	_ = secondID

	// filter by (patient, vid)
	list, err := repo.ListPrescriptions(ctx, &[]int64{int64(id)}[0], &[]int32{vid}[0])
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(list), 2)
	// ordered desc by created_at → second first
	assert.Equal(t, "second", *list[0].Notes)
	assert.Equal(t, "first", *list[1].Notes)

	// list all
	all, err := repo.ListPrescriptions(ctx, nil, nil)
	assert.Nil(t, err)
	assert.GreaterOrEqual(t, len(all), 2)
}

func TestPrescription_Update_CanMarkPacked(t *testing.T) {
	ctx := context.Background()
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Metformin"))
	_, locA := mustCreateBatchAndLocation(t, drug.ID, "M-A", 100, time.Now().AddDate(0, 6, 0), "Main")
	_, locB := mustCreateBatchAndLocation(t, drug.ID, "M-B", 100, time.Now().AddDate(0, 9, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	p, err := repo.CreatePrescription(ctx, &entities.Prescription{
		PatientID: int64(id),
		VID:       vid,
		Notes:     entities.PtrTo("init"),
		PrescribedDrugs: []entities.DrugPrescription{
			{DrugID: drug.ID, Remarks: nil, Batches: []entities.PrescriptionBatchItem{
				{BatchLocationId: locA, Quantity: 30},
				{BatchLocationId: locB, Quantity: 20},
			}},
		},
	})
	assert.Nil(t, err)

	p.Notes = entities.PtrTo("updated")
	p.IsPacked = true
	upd, err := repo.UpdatePrescription(ctx, p)
	assert.Nil(t, err)
	assert.Equal(t, "updated", *upd.Notes)
	assert.Equal(t, true, upd.IsPacked)
}

func TestPrescription_Update_AppliesBatchDeltas(t *testing.T) {
	ctx := context.Background()
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Metformin"))
	_, locA := mustCreateBatchAndLocation(t, drug.ID, "M-A", 100, time.Now().AddDate(0, 6, 0), "Main")
	_, locB := mustCreateBatchAndLocation(t, drug.ID, "M-B", 100, time.Now().AddDate(0, 9, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	// initial: A=30, B=20 → leaves A=70, B=80
	p, err := repo.CreatePrescription(ctx, &entities.Prescription{
		PatientID: int64(id),
		VID:       vid,
		Notes:     entities.PtrTo("init"),
		PrescribedDrugs: []entities.DrugPrescription{
			{DrugID: drug.ID, Remarks: nil, Batches: []entities.PrescriptionBatchItem{
				{BatchLocationId: locA, Quantity: 30},
				{BatchLocationId: locB, Quantity: 20},
			}},
		},
	})
	assert.Nil(t, err)

	aAfterCreate := getLocationQty(t, locA) // expect 70
	bAfterCreate := getLocationQty(t, locB) // expect 80

	// update: A=10, B=50  → delta A = -20 (return 20), delta B = +30 (take 30)
	p.Notes = entities.PtrTo("updated")
	p.PrescribedDrugs = []entities.DrugPrescription{
		{DrugID: drug.ID, Remarks: entities.PtrTo("new"), Batches: []entities.PrescriptionBatchItem{
			{BatchLocationId: locA, Quantity: 10},
			{BatchLocationId: locB, Quantity: 50},
		}},
	}
	upd, err := repo.UpdatePrescription(ctx, p)
	assert.Nil(t, err)
	assert.Equal(t, "updated", *upd.Notes)
	assert.Equal(t, 1, len(upd.PrescribedDrugs))
	assert.Equal(t, 2, len(upd.PrescribedDrugs[0].Batches))

	// verify deltas applied
	aFinal := getLocationQty(t, locA)
	bFinal := getLocationQty(t, locB)
	assert.Equal(t, aAfterCreate+20, aFinal) // 70 + 20 = 90
	assert.Equal(t, bAfterCreate-30, bFinal) // 80 - 30 = 50
}

func TestPrescription_Delete_RestoresStockAndRemovesRows(t *testing.T) {
	ctx := context.Background()
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Loratadine"))
	_, loc := mustCreateBatchAndLocation(t, drug.ID, "L-B1", 40, time.Now().AddDate(0, 4, 0), "Main")

	before := getLocationQty(t, loc)

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)
	p, err := repo.CreatePrescription(ctx, &entities.Prescription{
		PatientID: int64(id),
		VID:       vid,
		Notes:     entities.PtrTo("to-delete"),
		PrescribedDrugs: []entities.DrugPrescription{
			{DrugID: drug.ID, Remarks: nil, Batches: []entities.PrescriptionBatchItem{
				{BatchLocationId: loc, Quantity: 30},
			}},
		},
	})
	assert.Nil(t, err)

	mid := getLocationQty(t, loc)
	assert.Equal(t, before-30, mid)

	// delete
	err = repo.DeletePrescription(ctx, p.ID)
	assert.Nil(t, err)

	// stock restored
	after := getLocationQty(t, loc)
	assert.Equal(t, before, after)

	// deleting again → not found
	err = repo.DeletePrescription(ctx, p.ID)
	assert.EqualError(t, err, "prescription not found")
}
