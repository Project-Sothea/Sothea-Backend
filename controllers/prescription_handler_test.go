// controllers/prescription_handler_test.go
package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"context"

	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/jieqiboh/sothea_backend/mocks"
)

func init() { gin.SetMode(gin.TestMode) }

// Test router without auth middleware (mirrors production routes, but fixes "/:id")
func newTestPrescriptionHandlerNoAuth(r *gin.Engine, uc entities.PrescriptionUseCase) {
	h := &PrescriptionHandler{Usecase: uc}
	g := r.Group("/prescriptions")
	{
		g.GET("", h.ListPrescriptions)
		g.POST("", h.CreatePrescription)
		g.GET("/:id", h.GetPrescription)       // <-- leading slash
		g.PATCH("/:id", h.UpdatePrescription)  // <-- leading slash
		g.DELETE("/:id", h.DeletePrescription) // <-- leading slash
	}
}

// Minimal JSON bodies (kept generic so they won't 400 unless JSON is malformed)
const validPrescriptionJSON = `{"notes":"Take once daily"}`
const malformedJSON = `{"notes":` // invalid JSON to force bind error

// -----------------------------------------------------------------------------
// ListPrescriptions
// -----------------------------------------------------------------------------

func TestListPrescriptions_Success_NoFilters(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("ListPrescriptions", mock.Anything, (*int64)(nil), (*int32)(nil)).
		Return([]*entities.Prescription{{ID: 1}, {ID: 2}}, nil)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListPrescriptions", mock.Anything, (*int64)(nil), (*int32)(nil))
}

func TestListPrescriptions_Success_WithFilters(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	var pid int64 = 42
	var vid int32 = 7
	uc.On("ListPrescriptions", mock.Anything, &pid, &vid).
		Return([]*entities.Prescription{{ID: 99}}, nil)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	url := "/prescriptions?patient_id=" + strconv.FormatInt(int64(pid), 10) + "&vid=" + strconv.FormatInt(int64(vid), 10)
	req, _ := http.NewRequest("GET", url, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListPrescriptions", mock.Anything,
		mock.MatchedBy(func(p *int64) bool { return p != nil && *p == pid }),
		mock.MatchedBy(func(p *int32) bool { return p != nil && *p == vid }),
	)
}

func TestListPrescriptions_BadPatientID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions?patient_id=nan", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListPrescriptions_BadVid(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions?vid=oops", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListPrescriptions_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("ListPrescriptions", mock.Anything, (*int64)(nil), (*int32)(nil)).
		Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// CreatePrescription
// -----------------------------------------------------------------------------

func TestCreatePrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("CreatePrescription", mock.Anything, mock.AnythingOfType("*entities.Prescription")).
		Return(&entities.Prescription{ID: 123}, nil)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions", bytes.NewBufferString(validPrescriptionJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "CreatePrescription", mock.Anything, mock.AnythingOfType("*entities.Prescription"))
}

func TestCreatePrescription_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePrescription_JSONBindError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions", bytes.NewBufferString(malformedJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePrescription_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("CreatePrescription", mock.Anything, mock.AnythingOfType("*entities.Prescription")).
		Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions", bytes.NewBufferString(validPrescriptionJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// GetPrescription
// -----------------------------------------------------------------------------

func TestGetPrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("GetPrescriptionByID", mock.Anything, int64(7)).
		Return(&entities.Prescription{ID: 7}, nil)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "GetPrescriptionByID", mock.Anything, int64(7))
}

func TestGetPrescription_BadID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/not-a-number", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPrescription_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("GetPrescriptionByID", mock.Anything, int64(3)).
		Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/3", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// UpdatePrescription
// -----------------------------------------------------------------------------

func TestUpdatePrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("UpdatePrescription", mock.Anything, mock.AnythingOfType("*entities.Prescription")).
		Return(&entities.Prescription{ID: 5}, nil)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	body := `{"notes":"Changed"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/5", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdatePrescription", mock.Anything, mock.MatchedBy(func(p *entities.Prescription) bool {
		return p != nil && p.ID == 5
	}))
}

func TestUpdatePrescription_BadID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/abc", bytes.NewBufferString(validPrescriptionJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePrescription_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/1", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePrescription_JSONBindError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/1", bytes.NewBufferString(malformedJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePrescription_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("UpdatePrescription", mock.Anything, mock.AnythingOfType("*entities.Prescription")).
		Return(nil, entities.ErrInternalServerError)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/1", bytes.NewBufferString(validPrescriptionJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// DeletePrescription
// -----------------------------------------------------------------------------

func TestDeletePrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("DeletePrescription", mock.Anything, int64(55)).Return(nil)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/prescriptions/55", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertCalled(t, "DeletePrescription", mock.Anything, int64(55))
}

func TestDeletePrescription_BadID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/prescriptions/notnum", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePrescription_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("DeletePrescription", mock.Anything, int64(9)).
		Return(entities.ErrInternalServerError)

	r := gin.Default()
	newTestPrescriptionHandlerNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/prescriptions/9", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// Compile-time guard that our mock satisfies the interface
var _ entities.PrescriptionUseCase = (*mocks.PrescriptionUseCase)(nil)

// Prevent unused import warnings for context in some editors
var _ = context.Background()
