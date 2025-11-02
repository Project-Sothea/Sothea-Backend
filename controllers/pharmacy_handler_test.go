// controllers/pharmacy_handler_test.go
package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/jieqiboh/sothea_backend/mocks"
)

func init() { gin.SetMode(gin.TestMode) }

// -----------------------------------------------------------------------------
// Router (no auth) mirroring production routes
// -----------------------------------------------------------------------------
func newTestPharmacyHandlerNoAuth(r *gin.Engine, uc entities.PharmacyUseCase) {
	h := &PharmacyHandler{Usecase: uc}
	grp := r.Group("/pharmacy")
	{
		// DRUGS
		grp.GET("/drugs", h.ListDrugs) // ?q=parac
		grp.POST("/drugs", h.CreateDrug)
		grp.GET("/drugs/:drugId", h.GetDrugWithPresentations)
		grp.PATCH("/drugs/:drugId", h.UpdateDrug)
		grp.DELETE("/drugs/:drugId", h.DeleteDrug)

		// PRESENTATIONS
		grp.GET("/drugs/:drugId/presentations", h.ListPresentationsForDrug)
		grp.POST("/drugs/:drugId/presentations", h.CreatePresentation)

		grp.GET("/presentations/:presentationId", h.GetPresentation)
		grp.PATCH("/presentations/:presentationId", h.UpdatePresentation)
		grp.DELETE("/presentations/:presentationId", h.DeletePresentation)

		// BATCHES (scoped by presentation)
		grp.GET("/presentations/:presentationId/batches", h.ListBatches)
		grp.POST("/presentations/:presentationId/batches", h.CreateBatch)

		grp.GET("/batches/:batchId", h.GetBatch)
		grp.PATCH("/batches/:batchId", h.UpdateBatch)
		grp.DELETE("/batches/:batchId", h.DeleteBatch)

		// LOCATIONS
		grp.GET("/batches/:batchId/locations", h.ListBatchLocations)
		grp.POST("/batches/:batchId/locations", h.CreateBatchLocation)
		grp.PATCH("/locations/:locationId", h.UpdateBatchLocation)
		grp.DELETE("/locations/:locationId", h.DeleteBatchLocation)
	}
}

// -----------------------------------------------------------------------------
// JSON helpers (match new entities)
// -----------------------------------------------------------------------------
const (

	// DRUGS
	validDrugCreateJSON = `{
		"genericName":"Paracetamol",
		"brandName":"Tylenol",
		"atcCode":"N02BE01",
		"notes":"pain relief",
		"isActive":true
	}`
	validDrugUpdateJSON = `{
		"genericName":"Paracetamol",
		"brandName":"Panadol",
		"notes":"updated",
		"isActive":true
	}`
	badTypeDrugJSON = `{"genericName":123}`

	// PRESENTATIONS
	validPresentationCreateJSON = `{
		"dosageFormCode":"TAB",
		"routeCode":"PO",
		"strengthNum":500,
		"strengthUnitNum":"mg",
		"dispenseUnit":"tab",
		"isFractionalAllowed":false,
		"barcode":"1234567890123",
		"notes":"500 mg tablet"
	}`
	validPresentationUpdateJSON = `{
		"notes":"label tweak","barcode":"9876543210000"
	}`

	// BATCHES (CreateBatch uses wrapper {batch, locations})
	validBatchCreateJSON = `{
		"batch": {
			"batchNumber": "B-001",
			"quantity": 100,
			"expiryDate": "2030-12-31T00:00:00Z",
			"supplier": "ACME"
		},
		"locations": [
			{"location":"Main","quantity":60},
			{"location":"Cabinet A","quantity":40}
		]
	}`
	validBatchUpdateJSON = `{
		"batchNumber":"B-001-UPDATED",
		"quantity":120,
		"supplier":"NewCo"
	}`
	badTypeBatchJSON = `{"batch":{"batchNumber":123,"quantity":"oops"}}`

	// LOCATIONS
	locCreateJSON_ConflictingBodyBatchID = `{"batchId":999,"location":"Main","quantity":10}`
	locUpdateJSON                        = `{"location":"Secondary","quantity":30}`
)

// -----------------------------------------------------------------------------
// DRUGS
// -----------------------------------------------------------------------------
func TestListDrugs_Success_NoQuery(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListDrugs", mock.Anything, (*string)(nil)).
		Return([]entities.Drug{
			{ID: 1, GenericName: "Paracetamol", IsActive: true},
			{ID: 2, GenericName: "Ibuprofen", IsActive: true},
		}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListDrugs", mock.Anything, (*string)(nil))
}

func TestListDrugs_Success_WithQuery(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListDrugs", mock.Anything, mock.MatchedBy(func(p *string) bool { return p != nil && *p == "parac" })).
		Return([]entities.Drug{{ID: 1, GenericName: "Paracetamol"}}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs?q=parac", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListDrugs_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListDrugs", mock.Anything, mock.Anything).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateDrug_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	created := &entities.Drug{ID: 10, GenericName: "Paracetamol", IsActive: true}
	uc.On("CreateDrug", mock.Anything, mock.AnythingOfType("*entities.Drug")).Return(created, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/drugs", bytes.NewBufferString(validDrugCreateJSON))
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
	req, _ := http.NewRequest("POST", "/pharmacy/drugs", bytes.NewBufferString(emptyJSON))
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

func TestGetDrugWithPresentations_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetDrugWithPresentations", mock.Anything, int64(7)).
		Return(&entities.DrugWithPresentations{
			Drug: entities.Drug{ID: 7, GenericName: "Aspirin"},
			Presentations: []entities.DrugPresentationView{
				{DrugPresentation: entities.DrugPresentation{ID: 101}, DrugName: "Aspirin"},
				{DrugPresentation: entities.DrugPresentation{ID: 102}, DrugName: "Aspirin"},
			},
		}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "GetDrugWithPresentations", mock.Anything, int64(7))
}

func TestGetDrugWithPresentations_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/not-a-number", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDrugWithPresentations_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetDrugWithPresentations", mock.Anything, int64(3)).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/3", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateDrug_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	updated := &entities.Drug{ID: 5, GenericName: "Paracetamol", IsActive: true}
	uc.On("UpdateDrug", mock.Anything, mock.MatchedBy(func(d *entities.Drug) bool {
		return d != nil && d.ID == 5 && d.GenericName == "Paracetamol" && d.IsActive
	})).Return(updated, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/drugs/5", bytes.NewBufferString(validDrugUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateDrug_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/drugs/abc", bytes.NewBufferString(validDrugUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateDrug_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/drugs/1", bytes.NewBufferString(emptyJSON))
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
// PRESENTATIONS
// -----------------------------------------------------------------------------
func TestListPresentationsForDrug_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetDrugWithPresentations", mock.Anything, int64(5)).
		Return(&entities.DrugWithPresentations{
			Drug: entities.Drug{ID: 5, GenericName: "PCM"},
			Presentations: []entities.DrugPresentationView{
				{DrugPresentation: entities.DrugPresentation{ID: 101}, DrugName: "PCM"},
				{DrugPresentation: entities.DrugPresentation{ID: 102}, DrugName: "PCM"},
			},
		}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/5/presentations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "GetDrugWithPresentations", mock.Anything, int64(5))
}

func TestListPresentationsForDrug_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/NaN/presentations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListPresentationsForDrug_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetDrugWithPresentations", mock.Anything, int64(77)).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/drugs/77/presentations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreatePresentation_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("CreatePresentation", mock.Anything, mock.MatchedBy(func(p *entities.DrugPresentation) bool {
		return p != nil && p.DrugID == 5 && p.DosageFormCode == "TAB" && p.RouteCode == "PO" && p.DispenseUnit == "tab"
	})).Return(&entities.DrugPresentationView{
		DrugPresentation: entities.DrugPresentation{ID: 123, DrugID: 5, DosageFormCode: "TAB", RouteCode: "PO", DispenseUnit: "tab"},
		DrugName:         "Paracetamol",
	}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/drugs/5/presentations", bytes.NewBufferString(validPresentationCreateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "CreatePresentation", mock.Anything, mock.AnythingOfType("*entities.DrugPresentation"))
}

func TestCreatePresentation_BadDrugID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/drugs/notnum/presentations", bytes.NewBufferString(validPresentationCreateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePresentation_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/drugs/5/presentations", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPresentation_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetPresentationStock", mock.Anything, int64(11)).
		Return(&entities.PresentationStock{
			Presentation: entities.DrugPresentationView{
				DrugPresentation: entities.DrugPresentation{ID: 11, DrugID: 2, DosageFormCode: "TAB"},
				DrugName:         "Ibuprofen",
				DisplayStrength:  "200 mg tab",
				DisplayRoute:     "PO",
				DisplayLabel:     "Ibuprofen 200 mg tablet (PO)",
			},
			TotalQty: 100,
		}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/11", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "GetPresentationStock", mock.Anything, int64(11))
}

func TestGetPresentation_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/notnum", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPresentation_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetPresentationStock", mock.Anything, int64(11)).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/11", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdatePresentation_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("UpdatePresentation", mock.Anything, mock.MatchedBy(func(p *entities.DrugPresentation) bool {
		return p != nil && p.ID == 22
	})).Return(&entities.DrugPresentationView{
		DrugPresentation: entities.DrugPresentation{ID: 22, Notes: strPtr("label tweak")},
	}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/presentations/22", bytes.NewBufferString(validPresentationUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdatePresentation", mock.Anything, mock.AnythingOfType("*entities.DrugPresentation"))
}

func TestUpdatePresentation_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/presentations/xx", bytes.NewBufferString(validPresentationUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePresentation_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/presentations/1", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePresentation_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("DeletePresentation", mock.Anything, int64(44)).Return(nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/presentations/44", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertCalled(t, "DeletePresentation", mock.Anything, int64(44))
}

func TestDeletePresentation_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/presentations/notnum", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPresentationStock_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetPresentationStock", mock.Anything, int64(77)).
		Return(&entities.PresentationStock{
			Presentation: entities.DrugPresentationView{
				DrugPresentation: entities.DrugPresentation{ID: 77},
				DrugName:         "Paracetamol",
			},
			TotalQty: 12,
		}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/77/stock", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "GetPresentationStock", mock.Anything, int64(77))
}

func TestGetPresentationStock_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/xx/stock", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPresentationStock_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetPresentationStock", mock.Anything, int64(77)).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/77/stock", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// BATCHES (scoped by presentation)
// -----------------------------------------------------------------------------
func TestListBatches_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListBatches", mock.Anything, int64(5)).
		Return([]entities.BatchDetail{
			{DrugBatch: entities.DrugBatch{ID: 1, PresentationID: 5, Quantity: 10}, DispenseUnit: "tab"},
			{DrugBatch: entities.DrugBatch{ID: 2, PresentationID: 5, Quantity: 15}, DispenseUnit: "tab"},
		}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/5/batches", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListBatches", mock.Anything, int64(5))
}

func TestListBatches_BadPresentationID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/NaN/batches", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListBatches_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListBatches", mock.Anything, int64(5)).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/presentations/5/batches", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateBatch_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("CreateBatch", mock.Anything,
		mock.MatchedBy(func(b *entities.DrugBatch) bool {
			return b != nil && b.PresentationID == 5 && b.BatchNumber == "B-001" && b.Quantity == 100
		}),
		mock.MatchedBy(func(locs []entities.DrugBatchLocation) bool {
			return len(locs) == 2 && locs[0].Location == "Main" && locs[0].Quantity == 60
		}),
	).Return(&entities.BatchDetail{
		DrugBatch:    entities.DrugBatch{ID: 123, PresentationID: 5, BatchNumber: "B-001", Quantity: 100},
		DispenseUnit: "tab",
		BatchLocations: []entities.DrugBatchLocation{
			{ID: 1, BatchID: 123, Location: "Main", Quantity: 60},
			{ID: 2, BatchID: 123, Location: "Cabinet A", Quantity: 40},
		},
	}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/presentations/5/batches", bytes.NewBufferString(validBatchCreateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "CreateBatch", mock.Anything, mock.Anything, mock.Anything)
}

func TestCreateBatch_BadPresentationID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/presentations/xx/batches", bytes.NewBufferString(validBatchCreateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBatch_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/presentations/5/batches", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetBatch_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetBatch", mock.Anything, int64(88)).
		Return(&entities.BatchDetail{DrugBatch: entities.DrugBatch{ID: 88, Quantity: 10}, DispenseUnit: "tab"}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches/88", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "GetBatch", mock.Anything, int64(88))
}

func TestGetBatch_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches/notnum", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetBatch_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("GetBatch", mock.Anything, int64(88)).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches/88", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateBatch_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("UpdateBatch", mock.Anything, mock.MatchedBy(func(b *entities.DrugBatch) bool {
		return b != nil && b.ID == 42 && b.BatchNumber == "B-001-UPDATED" && b.Quantity == 120
	})).Return(&entities.BatchDetail{DrugBatch: entities.DrugBatch{ID: 42, BatchNumber: "B-001-UPDATED", Quantity: 120}}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/batches/42", bytes.NewBufferString(validBatchUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdateBatch", mock.Anything, mock.AnythingOfType("*entities.DrugBatch"))
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
	req, _ := http.NewRequest("PATCH", "/pharmacy/batches/1", bytes.NewBufferString(emptyJSON))
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
// LOCATIONS
// -----------------------------------------------------------------------------
func TestListBatchLocations_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListBatchLocations", mock.Anything, int64(10)).
		Return([]entities.DrugBatchLocation{{ID: 1, BatchID: 10, Location: "Main", Quantity: 9}}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches/10/locations", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListBatchLocations", mock.Anything, int64(10))
}

func TestListBatchLocations_BadBatchID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches/xx/locations", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListBatchLocations_InternalError(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("ListBatchLocations", mock.Anything, int64(10)).Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/pharmacy/batches/10/locations", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreateBatchLocation_Success_PathOverridesBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("CreateBatchLocation", mock.Anything, mock.MatchedBy(func(loc *entities.DrugBatchLocation) bool {
		return loc != nil && loc.BatchID == 123 && loc.Location == "Main" && loc.Quantity == 10
	})).Return(&entities.DrugBatchLocation{ID: 888, BatchID: 123, Location: "Main", Quantity: 10}, nil)

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

func TestCreateBatchLocation_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/pharmacy/batches/1/locations", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "CreateBatchLocation", mock.Anything, mock.Anything)
}

func TestUpdateBatchLocation_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("UpdateBatchLocation", mock.Anything, mock.MatchedBy(func(loc *entities.DrugBatchLocation) bool {
		return loc != nil && loc.ID == 888 && loc.Location == "Secondary" && loc.Quantity == 30
	})).Return(&entities.DrugBatchLocation{ID: 888, Location: "Secondary", Quantity: 30}, nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/locations/888", bytes.NewBufferString(locUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdateBatchLocation", mock.Anything, mock.AnythingOfType("*entities.DrugBatchLocation"))
}

func TestUpdateBatchLocation_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/locations/xx", bytes.NewBufferString(locUpdateJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateBatchLocation_EmptyBody(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/pharmacy/locations/1", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteBatchLocation_Success(t *testing.T) {
	var uc mocks.PharmacyUseCase
	uc.On("DeleteBatchLocation", mock.Anything, int64(999)).Return(nil)

	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/locations/999", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertCalled(t, "DeleteBatchLocation", mock.Anything, int64(999))
}

func TestDeleteBatchLocation_BadID(t *testing.T) {
	var uc mocks.PharmacyUseCase
	r := gin.Default()
	newTestPharmacyHandlerNoAuth(r, &uc)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/pharmacy/locations/notnum", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
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
