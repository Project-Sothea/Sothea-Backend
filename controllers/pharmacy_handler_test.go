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

// -----------------------------------------------------------------------------
// Test router without auth middleware (matches production routes exactly)
// -----------------------------------------------------------------------------
func newTestPharmacyHandlerNoAuth(r *gin.Engine, uc entities.PharmacyUseCase) {
	h := &PharmacyHandler{Usecase: uc}
	g := r.Group("/pharmacy")
	{
		// DRUGS
		g.GET("/drugs", h.ListDrugs)
		g.POST("/drugs", h.CreateDrug)
		g.GET("/drugs/:id", h.GetDrug)
		g.PATCH("/drugs/:id", h.UpdateDrug)
		g.DELETE("/drugs/:id", h.DeleteDrug)

		// BATCHES
		g.GET("/batches", h.ListBatches)
		g.POST("/batches", h.CreateBatch)
		g.PATCH("/batches/:id", h.UpdateBatch)
		g.DELETE("/batches/:id", h.DeleteBatch)
	}
}

func init() { gin.SetMode(gin.TestMode) }

// Minimal JSON helpers (adjust field names/types to your entities if needed)
const validDrugJSON = `{"name":"Paracetamol","form":"tablet","strengthMg":500}`
const badTypeDrugJSON = `{"name":123}` // forces JSON bind/type error for ShouldBindJSON

// For DrugBatch, adjust keys to your struct tags (e.g., drugId, quantity, expiryDate)
const validBatchJSON = `{"drugId":1,"quantity":50,"expiryDate":"2025-12-31T00:00:00Z"}`
const badTypeBatchJSON = `{"drugId":"oops","quantity":50}`

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
	created := &entities.Drug{ID: 10, Name: "Paracetamol"}
	uc.
		On("CreateDrug", mock.Anything, mock.AnythingOfType("*entities.Drug")).
		Return(created, nil)

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
				ID:   7,
				Name: "Aspirin",
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
	updated := &entities.Drug{ID: 5, Name: "Paracetamol Updated"}
	// If you haven't fixed the handler bug yet, change "UpdateDrug" to "CreateDrug" below.
	uc.
		On("UpdateDrug", mock.Anything, mock.AnythingOfType("*entities.Drug")).
		Return(updated, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	body := `{"name":"Paracetamol Updated"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/drugs/5", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdateDrug", mock.Anything, mock.MatchedBy(func(d *entities.Drug) bool {
		return d != nil && d.ID == 5 && d.Name == "Paracetamol Updated"
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
	uc.On("ListBatches", mock.Anything, (*int64)(nil)).Return([]entities.DrugBatch{
		{ID: 1, DrugID: 5}, {ID: 2, DrugID: 6},
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
	uc.On("ListBatches", mock.Anything, &id).Return([]entities.DrugBatch{
		{ID: 1, DrugID: 5},
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
	uc.On("CreateBatch", mock.Anything, mock.AnythingOfType("*entities.DrugBatch")).Return(int64(123), nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/batches", bytes.NewBufferString(validBatchJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "CreateBatch", mock.Anything, mock.AnythingOfType("*entities.DrugBatch"))
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
	uc.On("UpdateBatch", mock.Anything, mock.AnythingOfType("*entities.DrugBatch")).Return(nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/batches/42", bytes.NewBufferString(validBatchJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdateBatch", mock.Anything, mock.MatchedBy(func(b *entities.DrugBatch) bool {
		return b != nil && b.ID == 42
	}))
}

func TestUpdateBatch_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/batches/xyz", bytes.NewBufferString(validBatchJSON))
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
// Optional: quick smoke for error mapping
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
