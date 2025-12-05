package usecases

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/jieqiboh/sothea_backend/entities"
)

// -----------------------------------------------------------------------------
// Struct + Constructor
// -----------------------------------------------------------------------------

type pharmacyUsecase struct {
	repo           entities.PharmacyRepository
	contextTimeout time.Duration
}

func NewPharmacyUsecase(r entities.PharmacyRepository, timeout time.Duration) entities.PharmacyUseCase {
	return &pharmacyUsecase{
		repo:           r,
		contextTimeout: timeout,
	}
}

// -----------------------------------------------------------------------------
// Helpers (formatting & small validations)
// -----------------------------------------------------------------------------

func valOrZero(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
func valOrEmpty(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Minimal business validations to fail fast at usecase layer (DB has stricter checks)
func validatePresentationForCreate(p *entities.DrugPresentation) error {
	if p.DrugID == 0 || p.DosageFormCode == "" || p.RouteCode == "" || p.DispenseUnit == "" {
		return fmt.Errorf("missing required fields (drugId, dosageFormCode, routeCode, dispenseUnit)")
	}

	// Check if strength is unknown (all strength fields NULL)
	if p.StrengthNum == nil && p.StrengthUnitNum == nil {
		if p.StrengthDen != nil || p.StrengthUnitDen != nil {
			return fmt.Errorf("invalid strength configuration: strength_num/unit_num NULL but strength_den/unit_den not NULL")
		}
		// Unknown strength: any dispense_unit is allowed (mL, g, tab, cap, drop, bottle, etc.)
		// If bottle, piece_content is optional
		return nil
	}

	// Known strength cases
	if p.StrengthDen == nil {
		// Solids must have numerator filled
		if p.StrengthNum == nil || p.StrengthUnitNum == nil {
			return fmt.Errorf("solid presentation requires strength numerator and unit")
		}
	} else {

		// If dispensed as a piece (bottle), piece content is optional
		// If provided, piece_content_unit should match one of strength units
		if p.DispenseUnit == "bottle" && p.PieceContentAmount != nil && p.PieceContentUnit != nil {
			if p.StrengthUnitNum != nil && p.StrengthUnitDen != nil {
				if *p.PieceContentUnit != *p.StrengthUnitNum && *p.PieceContentUnit != *p.StrengthUnitDen {
					return fmt.Errorf("piece_content_unit must match one of strength units (strength_unit_num or strength_unit_den)")
				}
			}
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// PharmacyUseCase implementation
// -----------------------------------------------------------------------------

// ----------------- DRUGS -----------------

func (u *pharmacyUsecase) ListDrugs(ctx context.Context, q *string) ([]entities.Drug, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListDrugs(ctx, q)
}

func (u *pharmacyUsecase) GetDrugWithPresentations(ctx context.Context, drugID int64) (*entities.DrugWithPresentations, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	d, err := u.repo.GetDrug(ctx, drugID)
	if err != nil {
		return nil, err
	}
	ps, err := u.repo.ListPresentations(ctx, drugID)
	if err != nil {
		return nil, err
	}

	// Ensure labels are set (repo may already do this; harmless to recompute)
	out := make([]entities.DrugPresentationView, 0, len(ps))
	for _, p := range ps {
		out = append(out, p)
	}

	return &entities.DrugWithPresentations{
		Drug:          *d,
		Presentations: out,
	}, nil
}

func (u *pharmacyUsecase) CreateDrug(ctx context.Context, d *entities.Drug) (*entities.Drug, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateDrug(ctx, d)
}

func (u *pharmacyUsecase) UpdateDrug(ctx context.Context, d *entities.Drug) (*entities.Drug, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateDrug(ctx, d)
}

func (u *pharmacyUsecase) DeleteDrug(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteDrug(ctx, id)
}

// -------------- PRESENTATIONS --------------

func (u *pharmacyUsecase) GetPresentationStock(ctx context.Context, presentationID int64) (*entities.PresentationStock, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Load presentation (for labels)
	pv, err := u.repo.GetPresentation(ctx, presentationID)
	if err != nil {
		return nil, err
	}

	// Load batches + locations
	batches, err := u.repo.ListBatches(ctx, presentationID)
	if err != nil {
		return nil, err
	}

	// FEFO sort (expiry asc, then batch number)
	sort.SliceStable(batches, func(i, j int) bool {
		// nil expiry goes LAST
		var ti, tj time.Time
		if batches[i].ExpiryDate != nil {
			ti = *batches[i].ExpiryDate
		}
		if batches[j].ExpiryDate != nil {
			tj = *batches[j].ExpiryDate
		}
		if batches[i].ExpiryDate == nil && batches[j].ExpiryDate != nil {
			return false
		}
		if batches[i].ExpiryDate != nil && batches[j].ExpiryDate == nil {
			return true
		}
		if ti.Equal(tj) {
			return batches[i].BatchNumber < batches[j].BatchNumber
		}
		return ti.Before(tj)
	})

	total := 0
	for _, b := range batches {
		total += b.Quantity
	}

	return &entities.PresentationStock{
		Presentation: *pv,
		Batches:      batches,
		TotalQty:     total,
	}, nil
}

func (u *pharmacyUsecase) CreatePresentation(ctx context.Context, p *entities.DrugPresentation) (*entities.DrugPresentationView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := validatePresentationForCreate(p); err != nil {
		return nil, err
	}
	created, err := u.repo.CreatePresentation(ctx, p)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (u *pharmacyUsecase) UpdatePresentation(ctx context.Context, p *entities.DrugPresentation) (*entities.DrugPresentationView, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if err := validatePresentationForCreate(p); err != nil { // same checks okay for update
		return nil, err
	}
	updated, err := u.repo.UpdatePresentation(ctx, p)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func (u *pharmacyUsecase) DeletePresentation(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeletePresentation(ctx, id)
}

// ----------------- BATCHES -----------------

func (u *pharmacyUsecase) ListBatches(ctx context.Context, presentationID int64) ([]entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListBatches(ctx, presentationID)
}

func (u *pharmacyUsecase) GetBatch(ctx context.Context, batchID int64) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.GetBatch(ctx, batchID)
}

func (u *pharmacyUsecase) CreateBatch(ctx context.Context, b *entities.DrugBatch, locations []entities.DrugBatchLocation) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateBatch(ctx, b, locations)
}

func (u *pharmacyUsecase) UpdateBatch(ctx context.Context, b *entities.DrugBatch) (*entities.BatchDetail, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatch(ctx, b)
}

func (u *pharmacyUsecase) DeleteBatch(ctx context.Context, batchID int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatch(ctx, batchID)
}

// ------------- BATCH LOCATIONS -------------

func (u *pharmacyUsecase) ListBatchLocations(ctx context.Context, batchID int64) ([]entities.DrugBatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.ListBatchLocations(ctx, batchID)
}

func (u *pharmacyUsecase) CreateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.CreateBatchLocation(ctx, loc)
}

func (u *pharmacyUsecase) UpdateBatchLocation(ctx context.Context, loc *entities.DrugBatchLocation) (*entities.DrugBatchLocation, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.UpdateBatchLocation(ctx, loc)
}

func (u *pharmacyUsecase) DeleteBatchLocation(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	return u.repo.DeleteBatchLocation(ctx, id)
}
