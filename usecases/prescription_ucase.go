package usecases

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"sothea-backend/entities"
	"sothea-backend/repository/postgres"
	db "sothea-backend/repository/sqlc"
)

type PrescriptionUsecase struct {
	repo           *postgres.PostgresPrescriptionRepository
	pharmacy       *postgres.PostgresPharmacyRepository
	contextTimeout time.Duration
}

func NewPrescriptionUsecase(r *postgres.PostgresPrescriptionRepository, p *postgres.PostgresPharmacyRepository, timeout time.Duration) *PrescriptionUsecase {
	return &PrescriptionUsecase{
		repo:           r,
		pharmacy:       p,
		contextTimeout: timeout,
	}
}

// -----------------------------------------------------------------------------
// Basic CRUD
// -----------------------------------------------------------------------------

func (u *PrescriptionUsecase) CreatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreatePrescription(ctx, p)
}

func (u *PrescriptionUsecase) GetPrescriptionByID(ctx context.Context, id int64) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetPrescriptionByID(ctx, id)
}

func (u *PrescriptionUsecase) ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListPrescriptions(ctx, patientID, vid)
}

func (u *PrescriptionUsecase) UpdatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdatePrescription(ctx, p)
}

func (u *PrescriptionUsecase) DeletePrescription(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeletePrescription(ctx, id)
}

// -----------------------------------------------------------------------------
// Lines (one line = one presentation)
// -----------------------------------------------------------------------------

func (u *PrescriptionUsecase) AddLine(ctx context.Context, line *entities.PrescriptionLine) (*entities.PrescriptionLine, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if line.PrescriptionID == 0 || line.DrugID == 0 {
		return nil, errors.New("missing prescriptionId or drugId")
	}
	if line.DoseAmount <= 0 {
		return nil, errors.New("doseAmount must be > 0")
	}
	if line.Duration <= 0 {
		return nil, errors.New("duration must be > 0")
	}

	// DB trigger computes total_to_dispense based on schedule fields.
	return u.repo.AddLine(ctx, line)
}

func (u *PrescriptionUsecase) UpdateLine(ctx context.Context, line *entities.PrescriptionLine) (*entities.PrescriptionLine, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if line.ID == 0 {
		return nil, errors.New("missing line id")
	}
	if line.DoseAmount <= 0 {
		return nil, errors.New("doseAmount must be > 0")
	}
	if line.Duration <= 0 {
		return nil, errors.New("duration must be > 0")
	}

	// DB recomputes total_to_dispense; allocations are cleared in repo.
	return u.repo.UpdateLine(ctx, line)
}

func (u *PrescriptionUsecase) RemoveLine(ctx context.Context, lineID int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.RemoveLine(ctx, lineID)
}

// -----------------------------------------------------------------------------
// Packing allocations
// -----------------------------------------------------------------------------

func (u *PrescriptionUsecase) SetLineAllocations(ctx context.Context, lineID int64, allocs []db.PrescriptionBatchItem) ([]db.PrescriptionBatchItem, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if lineID == 0 {
		return nil, errors.New("missing lineId")
	}
	for i := range allocs {
		if allocs[i].Quantity <= 0 {
			return nil, fmt.Errorf("allocation %d has non-positive quantity", i)
		}
		allocs[i].LineID = lineID
	}

	// Repo replaces all allocations.
	// NOTE: DB triggers on prescription_batch_items will reserve/release stock per-row.
	out, err := u.repo.SetLineAllocations(ctx, lineID, allocs)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (u *PrescriptionUsecase) ListLineAllocations(ctx context.Context, lineID int64) ([]db.PrescriptionBatchItem, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Repo/DB enforces: sum(allocations) == total_to_dispense before allowing pack.
	return u.repo.ListLineAllocations(ctx, lineID)
}

func (u *PrescriptionUsecase) MarkLinePacked(ctx context.Context, lineID int64) (*entities.PrescriptionLine, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Repo/DB enforces: sum(allocations) == total_to_dispense before allowing pack.
	return u.repo.MarkLinePacked(ctx, lineID)
}

func (u *PrescriptionUsecase) UnpackLine(ctx context.Context, lineID int64) (*entities.PrescriptionLine, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UnpackLine(ctx, lineID)
}

// -----------------------------------------------------------------------------
// FEFO helper (optional convenience for the packing UI)
// -----------------------------------------------------------------------------

func (u *PrescriptionUsecase) SuggestFEFOAllocations(ctx context.Context, lineID int64) ([]db.PrescriptionBatchItem, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Load the line (needs PresentationID, TotalToDispense, DispenseUnit)
	line, err := u.repo.GetLine(ctx, lineID)
	if err != nil {
		return nil, err
	}

	// Get stock view for that presentation
	stock, err := u.pharmacy.GetDrugStock(ctx, line.DrugID)
	if err != nil {
		return nil, err
	}

	need := int(line.TotalToDispense)
	if need <= 0 {
		return nil, errors.New("line has zero totalToDispense")
	}

	// FEFO: earliest expiry first; stable-tie by ID
	type poolItem struct {
		BatchLocationID int64
		Qty             int
		Expiry          *time.Time
	}
	var pool []poolItem
	for _, b := range stock.Batches {
		for _, loc := range b.BatchLocations {
			pool = append(pool, poolItem{BatchLocationID: loc.ID, Qty: int(loc.Quantity), Expiry: b.DrugBatch.ExpiryDate})
		}
	}
	sort.SliceStable(pool, func(i, j int) bool {
		ie, je := pool[i].Expiry, pool[j].Expiry
		switch {
		case ie == nil && je != nil:
			return false
		case ie != nil && je == nil:
			return true
		case ie == nil && je == nil:
			return pool[i].BatchLocationID < pool[j].BatchLocationID
		default:
			if ie.Equal(*je) {
				return pool[i].BatchLocationID < pool[j].BatchLocationID
			}
			return ie.Before(*je)
		}
	})

	var out []db.PrescriptionBatchItem
	for need > 0 && len(pool) > 0 {
		head := &pool[0]
		if head.Qty == 0 {
			pool = pool[1:]
			continue
		}
		take := head.Qty
		if take > need {
			take = need
		}
		out = append(out, db.PrescriptionBatchItem{
			LineID:          lineID,
			BatchLocationID: head.BatchLocationID,
			Quantity:        int32(take),
		})
		need -= take
		pool[0].Qty -= take
		if pool[0].Qty == 0 {
			pool = pool[1:]
		}
	}
	if need > 0 {
		return out, fmt.Errorf("insufficient stock: short of %d %s", need, stock.Drug.Drug.DispenseUnit)
	}
	return out, nil
}

// -----------------------------------------------------------------------------
// Dispense (header-level): finalize Rx
// -----------------------------------------------------------------------------

func (u *PrescriptionUsecase) DispensePrescription(ctx context.Context, prescriptionID int64) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	// Repo performs one txn to validate all lines are packed and then stamps dispensed fields.
	// NOTE: Stock was already reserved/released by DB triggers on allocation changes.
	return u.repo.DispensePrescription(ctx, prescriptionID)
}
