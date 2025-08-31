// controllers/pharmacy_handler_test.go
package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/jieqiboh/sothea_backend/mocks"
)

func init() { gin.SetMode(gin.TestMode) }

// -----------------------------------------------------------------------------
// Router (no auth) mirroring production routes (with corrected locations paths)
// -----------------------------------------------------------------------------
func newTestPharmacyHandlerNoAuth(r *gin.Engine, uc entities.PharmacyUseCase) {
	h := &PharmacyHandler{Usecase: uc}
	grp := r.Group("/pharmacy")
	{
		// DRUG CATALOG
		grp.GET("/drugs", h.ListDrugs)
		grp.POST("/drugs", h.CreateDrug)
		grp.GET("/drugs/:drugId", h.GetDrug)
		grp.PATCH("/drugs/:drugId", h.UpdateDrug)
		grp.DELETE("/drugs/:drugId", h.DeleteDrug)

		// BATCHES
		grp.GET("/batches", h.ListBatches)
		grp.POST("/batches", h.CreateBatch)
		grp.PATCH("/batches/:batchId", h.UpdateBatch)
		grp.DELETE("/batches/:batchId", h.DeleteBatch)

		// BATCH LOCATIONS
		grp.POST("/batches/:batchId/locations", h.CreateBatchLocation)
		grp.PATCH("/batches/:batchId/locations/:locationId", h.UpdateBatchLocation)
		grp.DELETE("/batches/:batchId/locations/:locationId", h.DeleteBatchLocation)
	}
}

// -----------------------------------------------------------------------------
// JSON helpers that match your current entity tags
// -----------------------------------------------------------------------------
const validDrugJSON = `{"name":"Paracetamol","unit":"tablet","defaultSize":1,"notes":"pain relief"}`
const badTypeDrugJSON = `{"name":123}`

const validBatchCreateJSON = `
{
  "drugId": 1,
  "batchNumber": "B-001",
  "expiryDate": "2025-12-31T00:00:00Z",
  "supplier": "ACME",
  "batchLocations": [
    {"location": "Main", "quantity": 10},
    {"location": "Cabinet A", "quantity": 5}
  ]
}`

const validBatchUpdateJSON = `
{
  "drugId": 1,
  "batchNumber": "B-001-UPDATED",
  "expiryDate": "2026-01-01T00:00:00Z",
  "supplier": "NewCo"
}`

const badTypeBatchJSON = `{"drugId":"oops","batchNumber":"B-001"}`

const locCreateJSON_ConflictingBodyBatchID = `{"batchId":999,"location":"Main","quantity":10}`
const locUpdateJSON = `{"batchId":999,"location":"Main","quantity":30}`

// -----------------------------------------------------------------------------
// DRUGS
// -----------------------------------------------------------------------------

func TestListDrugs_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListDrugs", mock.Anything).Return([]entities.Drug{
		{ID: 1, Name: "Paracetamol"},
		{ID: 2, Name: "Ibuprofen"},
	}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListDrugs", mock.Anything)
}

func TestListDrugs_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListDrugs", mock.Anything).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateDrug_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	created := &entities.Drug{ID: 10, Name: "Paracetamol", Unit: "tablet"}
	uc.On("CreateDrug", mock.Anything, mock.AnythingOfType("*entities.Drug")).Return(created, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/drugs", bytes.NewBufferString(validDrugJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "CreateDrug", mock.Anything, mock.AnythingOfType("*entities.Drug"))
}

func TestCreateDrug_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/drugs", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateDrug_JSONBindError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/drugs", bytes.NewBufferString(badTypeDrugJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDrug_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetDrug", mock.Anything, int64(7)).
		Return(&entities.DrugDetail{
			Drug: entities.Drug{
				ID: 7, Name: "Aspirin",
			},
		}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "GetDrug", mock.Anything, int64(7))
}

func TestGetDrug_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/not-a-number", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDrug_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetDrug", mock.Anything, int64(3)).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/3", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateDrug_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	updated := &entities.Drug{ID: 5, Name: "Paracetamol Updated", Unit: "tablet"}
	uc.On("UpdateDrug", mock.Anything, mock.AnythingOfType("*entities.Drug")).Return(updated, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	body := `{"name":"Paracetamol Updated","unit":"tablet"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/drugs/5", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdateDrug", mock.Anything, mock.MatchedBy(func(d *entities.Drug) bool {
		return d != nil && d.ID == 5 && d.Name == "Paracetamol Updated" && d.Unit == "tablet"
	}))
}

func TestUpdateDrug_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/drugs/abc", bytes.NewBufferString(validDrugJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateDrug_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/drugs/1", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteDrug_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("DeleteDrug", mock.Anything, int64(9)).Return(nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/drugs/9", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertCalled(t, "DeleteDrug", mock.Anything, int64(9))
}

func TestDeleteDrug_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/drugs/notnum", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// -----------------------------------------------------------------------------
// BATCHES
// -----------------------------------------------------------------------------

func TestListBatches_Success_NoFilter(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListBatches", mock.Anything, (*int64)(nil)).Return([]entities.BatchDetail{
		{DrugBatch: entities.DrugBatch{ID: 1, DrugID: 5}},
		{DrugBatch: entities.DrugBatch{ID: 2, DrugID: 6}},
	}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListBatches", mock.Anything, (*int64)(nil))
}

func TestListBatches_Success_WithFilter(t *testing.T) {
	var uc mocks.PharmacyUseCase
	var id int64 = 5
	uc.On("ListBatches", mock.Anything, &id).Return([]entities.BatchDetail{
		{DrugBatch: entities.DrugBatch{ID: 1, DrugID: 5}},
	}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches?drug_id="+strconv.FormatInt(id, 10), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListBatches", mock.Anything, mock.MatchedBy(func(p *int64) bool { return p != nil && *p == 5 }))
}

func TestListBatches_BadQueryParam(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches?drug_id=nan", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBatch_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("CreateBatch", mock.Anything, mock.AnythingOfType("*entities.BatchDetail")).Return(
		&entities.BatchDetail{
			DrugBatch: entities.DrugBatch{
				ID:          123,
				DrugID:      7,
				BatchNumber: "BN-001",
			},
		},
		nil,
	)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/batches", bytes.NewBufferString(validBatchCreateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "CreateBatch", mock.Anything, mock.AnythingOfType("*entities.BatchDetail"))
}

func TestCreateBatch_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/batches", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateBatch_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	// Handler returns whatever usecase returns; assume *BatchDetail
	uc.On("UpdateBatch", mock.Anything, mock.AnythingOfType("*entities.DrugBatch")).
		Return(&entities.BatchDetail{DrugBatch: entities.DrugBatch{ID: 42, BatchNumber: "B-001-UPDATED"}}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/batches/42", bytes.NewBufferString(validBatchUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdateBatch", mock.Anything, mock.MatchedBy(func(b *entities.DrugBatch) bool {
		return b != nil && b.ID == 42 && b.BatchNumber == "B-001-UPDATED"
	}))
}

func TestUpdateBatch_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/batches/xyz", bytes.NewBufferString(validBatchUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateBatch_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/batches/1", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteBatch_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("DeleteBatch", mock.Anything, int64(55)).Return(nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/batches/55", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertCalled(t, "DeleteBatch", mock.Anything, int64(55))
}

func TestDeleteBatch_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/batches/bad", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// -----------------------------------------------------------------------------
// BATCH LOCATIONS
// -----------------------------------------------------------------------------

func TestCreateBatchLocation_Success_PathOverridesBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	// Expect BatchID = 123 from path, not 999 from body
	uc.On("CreateBatchLocation", mock.Anything, mock.MatchedBy(func(loc *entities.DrugBatchLocation) bool {
		return loc != nil && loc.BatchID == 123 && loc.Location == "Main" && loc.Quantity == 10
	})).Return(&entities.DrugBatchLocation{ID: 888, BatchID: 321, Location: "Main", Quantity: 30}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/batches/123/locations", bytes.NewBufferString(locCreateJSON_ConflictingBodyBatchID))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "CreateBatchLocation", mock.Anything, mock.AnythingOfType("*entities.DrugBatchLocation"))
}

func TestCreateBatchLocation_BadBatchID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/batches/notnum/locations", bytes.NewBufferString(locCreateJSON_ConflictingBodyBatchID))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateBatchLocation_Success_PathOverridesBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	// Expect BatchID from path = 321 and id from path = 888
	uc.On("UpdateBatchLocation", mock.Anything, mock.MatchedBy(func(loc *entities.DrugBatchLocation) bool {
		return loc != nil && loc.ID == 888 && loc.BatchID == 321 && loc.Location == "Main" && loc.Quantity == 30
	})).Return(&entities.DrugBatchLocation{ID: 888, BatchID: 321, Location: "Main", Quantity: 30}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/batches/321/locations/888", bytes.NewBufferString(locUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdateBatchLocation", mock.Anything, mock.AnythingOfType("*entities.DrugBatchLocation"))
}

func TestUpdateBatchLocation_BadIDs(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	// bad batchId
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("PATCH", "/pharmacy/batches/xx/locations/1", bytes.NewBufferString(locUpdateJSON))
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	// bad id
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("PATCH", "/pharmacy/batches/1/locations/yy", bytes.NewBufferString(locUpdateJSON))
	req2.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

func TestDeleteBatchLocation_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("DeleteBatchLocation", mock.Anything, int64(999)).Return(nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/batches/100/locations/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertCalled(t, "DeleteBatchLocation", mock.Anything, int64(999))
}

func TestDeleteBatchLocation_BadIDs(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	// bad batchId
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("DELETE", "/pharmacy/batches/xx/locations/1", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusBadRequest, w1.Code)

	// bad id
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("DELETE", "/pharmacy/batches/1/locations/yy", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusBadRequest, w2.Code)
}

// -----------------------------------------------------------------------------
// mapPhErr smoke
// -----------------------------------------------------------------------------
func TestMapPhErr_InternalAndConflict(t *testing.T) {
	assert.Equal(t, http.StatusInternalServerError, mapPhErr(entities.ErrInternalServerError))
	assert.Equal(t, http.StatusConflict, mapPhErr(entities.ErrDrugNameTaken))
	assert.Equal(t, http.StatusInternalServerError, mapPhErr(assert.AnError)) // default
}

// Tiny compilation guard to ensure the interface matches our mock expectations.
var _ = func() any {
	var _ entities.PharmacyUseCase = (*mocks.PharmacyUseCase)(nil)
	return nil
}()

// Prevent unused import errors when context is not matched literally.
var _ = context.Background()
