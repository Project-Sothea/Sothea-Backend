package usecases

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
)

// -----------------------------------------------------------------------------
// Struct + Constructor
// -----------------------------------------------------------------------------

type prescriptionUsecase struct {
	repo           entities.PrescriptionRepository
	pharmacy       entities.PharmacyRepository
	contextTimeout time.Duration
}

func NewPrescriptionUsecase(
	r entities.PrescriptionRepository,
	p entities.PharmacyRepository,
	timeout time.Duration,
) entities.PrescriptionUseCase {
	return &prescriptionUsecase{
		repo:           r,
		pharmacy:       p,
		contextTimeout: timeout,
	}
}

// -----------------------------------------------------------------------------
// Basic CRUD
// -----------------------------------------------------------------------------

func (u *prescriptionUsecase) CreatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreatePrescription(ctx, p)
}

func (u *prescriptionUsecase) GetPrescriptionByID(ctx context.Context, id int64) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetPrescriptionByID(ctx, id)
}

func (u *prescriptionUsecase) ListPrescriptions(ctx context.Context, patientID *int64, vid *int32) ([]*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListPrescriptions(ctx, patientID, vid)
}

func (u *prescriptionUsecase) UpdatePrescription(ctx context.Context, p *entities.Prescription) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdatePrescription(ctx, p)
}

func (u *prescriptionUsecase) DeletePrescription(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeletePrescription(ctx, id)
}

// -----------------------------------------------------------------------------
// Lines (one line = one presentation)
// -----------------------------------------------------------------------------

func (u *prescriptionUsecase) AddLine(ctx context.Context, line *entities.PrescriptionLine) (*entities.PrescriptionLine, error) {
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

func (u *prescriptionUsecase) UpdateLine(ctx context.Context, line *entities.PrescriptionLine) (*entities.PrescriptionLine, error) {
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

func (u *prescriptionUsecase) RemoveLine(ctx context.Context, lineID int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.RemoveLine(ctx, lineID)
}

// -----------------------------------------------------------------------------
// Packing allocations
// -----------------------------------------------------------------------------

func (u *prescriptionUsecase) SetLineAllocations(ctx context.Context, lineID int64, allocs []entities.LineAllocation) ([]entities.LineAllocation, error) {
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

func (u *prescriptionUsecase) ListLineAllocations(ctx context.Context, lineID int64) ([]entities.LineAllocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Repo/DB enforces: sum(allocations) == total_to_dispense before allowing pack.
	return u.repo.ListLineAllocations(ctx, lineID)
}

func (u *prescriptionUsecase) MarkLinePacked(ctx context.Context, lineID int64) (*entities.PrescriptionLine, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Repo/DB enforces: sum(allocations) == total_to_dispense before allowing pack.
	return u.repo.MarkLinePacked(ctx, lineID)
}

func (u *prescriptionUsecase) UnpackLine(ctx context.Context, lineID int64) (*entities.PrescriptionLine, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UnpackLine(ctx, lineID)
}

// -----------------------------------------------------------------------------
// FEFO helper (optional convenience for the packing UI)
// -----------------------------------------------------------------------------

func (u *prescriptionUsecase) SuggestFEFOAllocations(ctx context.Context, lineID int64) ([]entities.LineAllocation, error) {
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

	need := line.TotalToDispense
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
			pool = append(pool, poolItem{BatchLocationID: loc.ID, Qty: loc.Quantity, Expiry: b.ExpiryDate})
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

	out := []entities.LineAllocation{}
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
		out = append(out, entities.LineAllocation{
			LineID:          lineID,
			BatchLocationID: head.BatchLocationID,
			Quantity:        take,
		})
		need -= take
		pool[0].Qty -= take
		if pool[0].Qty == 0 {
			pool = pool[1:]
		}
	}
	if need > 0 {
		return out, fmt.Errorf("insufficient stock: short of %d %s", need, line.DispenseUnit)
	}
	return out, nil
}

// -----------------------------------------------------------------------------
// Dispense (header-level): finalize Rx
// -----------------------------------------------------------------------------

func (u *prescriptionUsecase) DispensePrescription(ctx context.Context, prescriptionID int64) (*entities.Prescription, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	// Repo performs one txn to validate all lines are packed and then stamps dispensed fields.
	// NOTE: Stock was already reserved/released by DB triggers on allocation changes.
	return u.repo.DispensePrescription(ctx, prescriptionID)
}
