// postgres_prescription_repository_test.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- helpers ----------

func unique(prefix string) string { return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano()) }

// runInTxAsUser opens a single transaction, sets the session user GUC,
// injects the tx into context (so repo picks it up), runs fn, and commits.
func runInTxAsUser(t *testing.T, userID int64, fn func(ctx context.Context)) {
	t.Helper()

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	require.NoError(t, err)
	defer tx.Rollback()

	// set GUC inside this tx
	_, err = tx.ExecContext(
		context.Background(),
		`SELECT set_config('sothea.user_id', $1, true)`,
		strconv.FormatInt(userID, 10),
	)
	require.NoError(t, err)

	// sanity check
	var got sql.NullString
	err = tx.QueryRowContext(context.Background(),
		`SELECT current_setting('sothea.user_id', true)`).Scan(&got)
	require.NoError(t, err)
	require.Equal(t, strconv.FormatInt(userID, 10), got.String)

	ctx := CtxWithTx(context.Background(), tx)

	// optional assert to verify wiring
	if txx, ok := TxFromCtx(ctx); !ok || txx == nil {
		t.Fatalf("no tx found in ctx")
	}

	fn(ctx)
	require.NoError(t, tx.Commit())
}

func mustAdminID(t *testing.T) int64 {
	t.Helper()
	repo := NewPostgresPatientRepository(db).(*postgresPatientRepository)

	u, err := repo.GetDBUser(context.Background(), "admin")
	require.NoError(t, err)
	require.NotNil(t, u)

	return u.Id
}

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
	require.NoError(t, err)

	vid, err := patRepo.CreatePatientVisit(ctx, id, &admin)
	require.NoError(t, err)

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
	require.NoError(t, err)
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
	require.NoError(t, err)

	var locID int64
	err = db.QueryRowContext(ctx, `
		INSERT INTO batch_locations (batch_id, location, quantity)
		VALUES ($1, $2, $3)
		RETURNING id
	`, batchID, location, qty).Scan(&locID)
	require.NoError(t, err)

	return batchID, locID
}

func getLocationQty(t *testing.T, locationID int64) int64 {
	t.Helper()
	var q int64
	err := db.QueryRow(`SELECT quantity FROM batch_locations WHERE id=$1`, locationID).Scan(&q)
	require.NoError(t, err)
	return q
}

// ---------- tests ----------

func TestPrescription_CreateAndGet_ReducesStockAndHydrates(t *testing.T) {
	adminID := mustAdminID(t)
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Paracetamol"))
	_, loc1 := mustCreateBatchAndLocation(t, drug.ID, "P-B1", 100, time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC), "Main")
	_, loc2 := mustCreateBatchAndLocation(t, drug.ID, "P-B2", 50, time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC), "Main")

	before1 := getLocationQty(t, loc1)
	before2 := getLocationQty(t, loc2)

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	var created *entities.Prescription
	runInTxAsUser(t, adminID, func(ctx context.Context) {
		p := &entities.Prescription{
			PatientID: int64(id),
			VID:       vid,
			Notes:     entities.PtrTo("Take after meals"),
			PrescribedDrugs: []entities.DrugPrescription{
				{
					DrugID:       drug.ID,
					Remarks:      entities.PtrTo("q8h"),
					RequestedQty: 70, // sum of batches below
					Batches: []entities.PrescriptionBatchItem{
						{BatchLocationId: loc1, Quantity: 40},
						{BatchLocationId: loc2, Quantity: 30},
					},
				},
			},
		}
		var err error
		created, err = repo.CreatePrescription(ctx, p)
		require.NoError(t, err)
		require.NotNil(t, created)

		// created_by should be stamped to adminID; default is_dispensed should be false
		var createdBy sql.NullInt64
		var isDispensed bool
		tx, _ := TxFromCtx(ctx)
		err = tx.QueryRowContext(ctx,
			`SELECT created_by, is_dispensed FROM prescriptions WHERE id = $1`, created.ID).
			Scan(&createdBy, &isDispensed)
		require.NoError(t, err)
		require.True(t, createdBy.Valid)
		assert.EqualValues(t, adminID, createdBy.Int64)
		assert.False(t, isDispensed)

		// hydrate via Get inside same GUC/tx path
		got, err := repo.GetPrescriptionByID(ctx, created.ID)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, created.ID, got.ID)
	})

	require.NotZero(t, created.ID)
	assert.Equal(t, int64(id), created.PatientID)
	assert.Equal(t, vid, created.VID)
	require.Len(t, created.PrescribedDrugs, 1)
	require.Len(t, created.PrescribedDrugs[0].Batches, 2)

	after1 := getLocationQty(t, loc1)
	after2 := getLocationQty(t, loc2)
	assert.Equal(t, before1-40, after1)
	assert.Equal(t, before2-30, after2)
}

func TestPrescription_Create_FailsWhenInsufficientStock(t *testing.T) {
	adminID := mustAdminID(t)
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Amoxicillin"))
	_, loc := mustCreateBatchAndLocation(t, drug.ID, "A-B1", 5, time.Now().AddDate(0, 6, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	runInTxAsUser(t, adminID, func(ctx context.Context) {
		_, err := repo.CreatePrescription(ctx, &entities.Prescription{
			PatientID: int64(id),
			VID:       vid,
			Notes:     entities.PtrTo("Too much"),
			PrescribedDrugs: []entities.DrugPrescription{
				{
					DrugID:       drug.ID,
					Remarks:      nil,
					RequestedQty: 10, // > stock available (5)
					Batches:      []entities.PrescriptionBatchItem{{BatchLocationId: loc, Quantity: 10}},
				},
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient stock in batch-location")
	})
}

func TestPrescription_List_WithFiltersAndOrder(t *testing.T) {
	adminID := mustAdminID(t)
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Ibuprofen"))
	_, loc := mustCreateBatchAndLocation(t, drug.ID, "I-B1", 100, time.Now().AddDate(0, 12, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	make := func(note string, q int) int64 {
		var pid int64
		runInTxAsUser(t, adminID, func(ctx context.Context) {
			p, err := repo.CreatePrescription(ctx, &entities.Prescription{
				PatientID: int64(id),
				VID:       vid,
				Notes:     &note,
				PrescribedDrugs: []entities.DrugPrescription{
					{
						DrugID:       drug.ID,
						Remarks:      nil,
						RequestedQty: int64(q), // match batch total
						Batches:      []entities.PrescriptionBatchItem{{BatchLocationId: loc, Quantity: q}},
					},
				},
			})
			require.NoError(t, err)
			require.NotNil(t, p)
			pid = p.ID
		})
		return pid
	}
	firstID := make("first", 10)
	_ = firstID
	time.Sleep(5 * time.Millisecond) // ensure created_at ordering
	secondID := make("second", 5)
	_ = secondID

	list, err := repo.ListPrescriptions(context.Background(), &[]int64{int64(id)}[0], &[]int32{vid}[0])
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(list), 2)
	assert.Equal(t, "second", *list[0].Notes) // most recent first
	assert.Equal(t, "first", *list[1].Notes)

	all, err := repo.ListPrescriptions(context.Background(), nil, nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(all), 2)
}

// Transition some drug lines to packed → stamp packed_by/packed_at at the line level
func TestPrescription_Update_MarksPackedLines_WithStamps(t *testing.T) {
	adminID := mustAdminID(t)
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Metformin"))
	_, locA := mustCreateBatchAndLocation(t, drug.ID, "M-A", 100, time.Now().AddDate(0, 6, 0), "Main")
	_, locB := mustCreateBatchAndLocation(t, drug.ID, "M-B", 100, time.Now().AddDate(0, 9, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	var p *entities.Prescription
	runInTxAsUser(t, adminID, func(ctx context.Context) {
		var err error
		p, err = repo.CreatePrescription(ctx, &entities.Prescription{
			PatientID: int64(id),
			VID:       vid,
			Notes:     entities.PtrTo("init"),
			PrescribedDrugs: []entities.DrugPrescription{
				{
					DrugID:       drug.ID,
					Remarks:      nil,
					RequestedQty: 50, // 30 + 20
					Batches: []entities.PrescriptionBatchItem{
						{BatchLocationId: locA, Quantity: 30},
						{BatchLocationId: locB, Quantity: 20},
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, p)
	})

	// Mark the line as packed (transition false->true triggers stamp in Step 6)
	p.Notes = entities.PtrTo("packed-now")
	for i := range p.PrescribedDrugs {
		if p.PrescribedDrugs[i].DrugID == drug.ID {
			p.PrescribedDrugs[i].IsPacked = true
		}
	}

	runInTxAsUser(t, adminID, func(ctx context.Context) {
		upd, err := repo.UpdatePrescription(ctx, p)
		require.NoError(t, err)
		require.NotNil(t, upd)
		assert.Equal(t, "packed-now", *upd.Notes)

		// Verify stamp at the line level
		tx, _ := TxFromCtx(ctx)
		type packedRow struct {
			isPacked   bool
			packedBy   sql.NullInt64
			packedAt   sql.NullTime
			prescrID   int64
			drugID     int64
			requestedQ int64
		}
		var got packedRow
		err = tx.QueryRowContext(ctx, `
			SELECT is_packed, packed_by, packed_at, prescription_id, drug_id, quantity_requested
			FROM drug_prescriptions
			WHERE prescription_id = $1 AND drug_id = $2
			ORDER BY id LIMIT 1
		`, upd.ID, drug.ID).Scan(&got.isPacked, &got.packedBy, &got.packedAt, &got.prescrID, &got.drugID, &got.requestedQ)
		require.NoError(t, err)
		assert.True(t, got.isPacked)
		require.True(t, got.packedBy.Valid)
		assert.EqualValues(t, adminID, got.packedBy.Int64)
		require.True(t, got.packedAt.Valid)
	})
}

func TestPrescription_Update_AppliesBatchDeltas(t *testing.T) {
	adminID := mustAdminID(t)
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Metformin"))
	_, locA := mustCreateBatchAndLocation(t, drug.ID, "M-A", 100, time.Now().AddDate(0, 6, 0), "Main")
	_, locB := mustCreateBatchAndLocation(t, drug.ID, "M-B", 100, time.Now().AddDate(0, 9, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	var p *entities.Prescription
	runInTxAsUser(t, adminID, func(ctx context.Context) {
		var err error
		p, err = repo.CreatePrescription(ctx, &entities.Prescription{
			PatientID: int64(id),
			VID:       vid,
			Notes:     entities.PtrTo("init"),
			PrescribedDrugs: []entities.DrugPrescription{
				{
					DrugID:       drug.ID,
					Remarks:      nil,
					RequestedQty: 50, // 30 + 20
					Batches: []entities.PrescriptionBatchItem{
						{BatchLocationId: locA, Quantity: 30},
						{BatchLocationId: locB, Quantity: 20},
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, p)
	})

	aAfterCreate := getLocationQty(t, locA) // expect 70
	bAfterCreate := getLocationQty(t, locB) // expect 80

	// update: A=10, B=50  → delta A = -20 (return 20), delta B = +30 (take 30)
	p.Notes = entities.PtrTo("updated")
	p.PrescribedDrugs = []entities.DrugPrescription{
		{
			DrugID:       drug.ID,
			Remarks:      entities.PtrTo("new"),
			RequestedQty: 60, // 10 + 50
			Batches: []entities.PrescriptionBatchItem{
				{BatchLocationId: locA, Quantity: 10},
				{BatchLocationId: locB, Quantity: 50},
			},
		},
	}

	runInTxAsUser(t, adminID, func(ctx context.Context) {
		upd, err := repo.UpdatePrescription(ctx, p)
		require.NoError(t, err)
		require.NotNil(t, upd)
		assert.Equal(t, "updated", *upd.Notes)
		require.Len(t, upd.PrescribedDrugs, 1)
		require.Len(t, upd.PrescribedDrugs[0].Batches, 2)
	})

	aFinal := getLocationQty(t, locA)
	bFinal := getLocationQty(t, locB)
	assert.Equal(t, aAfterCreate+20, aFinal) // 70 + 20 = 90
	assert.Equal(t, bAfterCreate-30, bFinal) // 80 - 30 = 50
}

func TestPrescription_Update_StatusDispensed_StampsDispenserAndBlocksFurtherEdits(t *testing.T) {
	adminID := mustAdminID(t)
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Lisinopril"))
	_, loc := mustCreateBatchAndLocation(t, drug.ID, "L-A", 100, time.Now().AddDate(0, 6, 0), "Main")

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	var p *entities.Prescription
	runInTxAsUser(t, adminID, func(ctx context.Context) {
		var err error
		p, err = repo.CreatePrescription(ctx, &entities.Prescription{
			PatientID: int64(id),
			VID:       vid,
			Notes:     entities.PtrTo("to-dispense"),
			PrescribedDrugs: []entities.DrugPrescription{
				{
					DrugID:       drug.ID,
					Remarks:      nil,
					RequestedQty: 10, // 10 allocated below
					Batches: []entities.PrescriptionBatchItem{
						{BatchLocationId: loc, Quantity: 10},
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, p)
	})

	// Move directly to DISPENSED (repo updates is_dispensed + stamps by/at)
	p.IsDispensed = true

	runInTxAsUser(t, adminID, func(ctx context.Context) {
		upd, err := repo.UpdatePrescription(ctx, p)
		require.NoError(t, err)
		require.NotNil(t, upd)
		assert.True(t, upd.IsDispensed)

		// verify dispensed_by/dispensed_at stamped
		var dispensedBy sql.NullInt64
		var dispensedAt sql.NullTime
		tx, _ := TxFromCtx(ctx)
		err = tx.QueryRowContext(ctx, `
			SELECT dispensed_by, dispensed_at
			FROM prescriptions
			WHERE id = $1
		`, upd.ID).Scan(&dispensedBy, &dispensedAt)
		require.NoError(t, err)
		require.True(t, dispensedBy.Valid)
		assert.EqualValues(t, adminID, dispensedBy.Int64)
		require.True(t, dispensedAt.Valid)
	})

	// Any further edits should be blocked
	runInTxAsUser(t, adminID, func(ctx context.Context) {
		p.Notes = entities.PtrTo("after-dispense-attempt")
		_, err := repo.UpdatePrescription(ctx, p)
		require.EqualError(t, err, "cannot update a dispensed prescription")
	})
}

func TestPrescription_Delete_RestoresStockAndRemovesRows(t *testing.T) {
	adminID := mustAdminID(t)
	id, vid := mustCreatePatientAndVisit(t)

	drug := mustCreateDrug(t, unique("Loratadine"))
	_, loc := mustCreateBatchAndLocation(t, drug.ID, "L-B1", 40, time.Now().AddDate(0, 4, 0), "Main")

	before := getLocationQty(t, loc)

	repo := NewPostgresPrescriptionRepository(db).(*postgresPrescriptionRepository)

	var pid int64
	runInTxAsUser(t, adminID, func(ctx context.Context) {
		p, err := repo.CreatePrescription(ctx, &entities.Prescription{
			PatientID: int64(id),
			VID:       vid,
			Notes:     entities.PtrTo("to-delete"),
			PrescribedDrugs: []entities.DrugPrescription{
				{
					DrugID:       drug.ID,
					Remarks:      nil,
					RequestedQty: 30,
					Batches: []entities.PrescriptionBatchItem{
						{BatchLocationId: loc, Quantity: 30},
					},
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, p)
		pid = p.ID
	})

	mid := getLocationQty(t, loc)
	assert.Equal(t, before-30, mid)

	// delete (no need for user GUC here, but keep pattern consistent)
	runInTxAsUser(t, adminID, func(ctx context.Context) {
		err := repo.DeletePrescription(ctx, pid)
		require.NoError(t, err)
	})

	after := getLocationQty(t, loc)
	assert.Equal(t, before, after)

	// deleting again → not found
	runInTxAsUser(t, adminID, func(ctx context.Context) {
		err := repo.DeletePrescription(ctx, pid)
		require.EqualError(t, err, "prescription not found")
	})
}
