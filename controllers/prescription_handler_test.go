// controllers/prescription_handler_test.go
package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
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
// Router mirroring production routes (no auth / no tx)
// -----------------------------------------------------------------------------
func newTestPrescriptionRouterNoAuth(r *gin.Engine, uc entities.PrescriptionUseCase) {
	h := &PrescriptionHandler{Usecase: uc}

	grp := r.Group("/prescriptions")
	{
		// Header CRUD
		grp.GET("", h.ListPrescriptions)
		grp.GET("/:id", h.GetPrescription)
		grp.POST("", h.CreatePrescription)
		grp.PATCH("/:id", h.UpdatePrescription)
		grp.DELETE("/:id", h.DeletePrescription)

		// Lines
		grp.POST("/:id/lines", h.AddLine)
		grp.PATCH("/lines/:lineId", h.UpdateLine)
		grp.DELETE("/lines/:lineId", h.RemoveLine)

		// Allocations
		grp.GET("/lines/:lineId/allocations", h.ListLineAllocations)
		grp.PUT("/lines/:lineId/allocations", h.SetLineAllocations)

		// Pack / Unpack
		grp.POST("/lines/:lineId/pack", h.MarkLinePacked)
		grp.POST("/lines/:lineId/unpack", h.UnpackLine)

		// Dispense
		grp.POST("/:id/dispense", h.DispensePrescription)
	}
}

// -----------------------------------------------------------------------------
// JSON fixtures
// -----------------------------------------------------------------------------
const (
	emptyJSON = ``

	// Header
	validCreatePrescriptionJSON = `{"patientId":1,"vid":2,"notes":"initial note"}`
	validUpdatePrescriptionJSON = `{"notes":"updated note"}`
	badTypePrescriptionJSON     = `{"patientId":"oops"}`
	// Lines (add/update)
	validLineJSON = `{
		"presentationId": 5,
		"remarks": "take with food",
		"doseAmount": 2,
		"doseUnit": "tablet",
		"scheduleKind": "day",
		"everyN": 1,
		"frequencyPerSchedule": 3,
		"duration": 5
	}`
	missingRequiredLineJSON = `{
		"presentationId": 5,
		"doseUnit": "tablet",
		"scheduleKind": "day",
		"everyN": 1,
		"frequencyPerSchedule": 3,
		"duration": 5
	}` // doseAmount missing → 400 due to binding tag

	badEnumLineJSON = `{
		"presentationId": 5,
		"doseAmount": 2,
		"doseUnit": "tablet",
		"scheduleKind": "fortnight",
		"everyN": 1,
		"frequencyPerSchedule": 3,
		"duration": 5
	}` // invalid scheduleKind → 400

	badTypeLineJSON = `{"presentationId":"oops"}`

	// Allocations
	validSetAllocationsJSON = `{
		"allocations": [
			{"batchLocationId": 11, "quantity": 3},
			{"batchLocationId": 12, "quantity": 2}
		]
	}`

	// Pack / Unpack
	validPackJSON   = `{"packedBy": 99}`
	badTypePackJSON = `{"packedBy":"oops"}`

	// Dispense
	validDispenseJSON   = `{"userId": 77}`
	badTypeDispenseJSON = `{"userId":"oops"}`
)

// -----------------------------------------------------------------------------
// Header CRUD
// -----------------------------------------------------------------------------
func TestListPrescriptions_Success_NoFilters(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("ListPrescriptions", mock.Anything, (*int64)(nil), (*int32)(nil)).
		Return([]*entities.Prescription{
			{ID: 1, PatientID: 1, VID: 2},
			{ID: 2, PatientID: 2, VID: 1},
		}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

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

	uc.On("ListPrescriptions",
		mock.Anything,
		mock.MatchedBy(func(p *int64) bool { return p != nil && *p == pid }),
		mock.MatchedBy(func(v *int32) bool { return v != nil && *v == vid }),
	).Return([]*entities.Prescription{{ID: 9, PatientID: pid, VID: vid}}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	url := "/prescriptions?patient_id=" + strconv.FormatInt(int64(pid), 10) + "&vid=" + strconv.FormatInt(int64(vid), 10)
	req, _ := http.NewRequest("GET", url, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListPrescriptions", mock.Anything, mock.Anything, mock.Anything)
}

func TestListPrescriptions_BadPatientID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions?patient_id=NaN", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListPrescriptions_BadVid(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions?vid=notnum", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListPrescriptions_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("ListPrescriptions", mock.Anything, (*int64)(nil), (*int32)(nil)).
		Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCreatePrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("CreatePrescription", mock.Anything, mock.AnythingOfType("*entities.Prescription")).
		Return(&entities.Prescription{ID: 100, PatientID: 1, VID: 2}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions", bytes.NewBufferString(validCreatePrescriptionJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "CreatePrescription", mock.Anything, mock.AnythingOfType("*entities.Prescription"))
}

func TestCreatePrescription_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreatePrescription_JSONBindError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions", bytes.NewBufferString(badTypePrescriptionJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("GetPrescriptionByID", mock.Anything, int64(7)).
		Return(&entities.Prescription{ID: 7, PatientID: 1, VID: 2}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/7", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "GetPrescriptionByID", mock.Anything, int64(7))
}

func TestGetPrescription_BadID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/not-a-number", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetPrescription_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("GetPrescriptionByID", mock.Anything, int64(3)).Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/3", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdatePrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("UpdatePrescription", mock.Anything, mock.MatchedBy(func(p *entities.Prescription) bool {
		return p != nil && p.ID == 5
	})).Return(&entities.Prescription{ID: 5, PatientID: 1, VID: 2, Notes: strPtr("updated note")}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/5", bytes.NewBufferString(validUpdatePrescriptionJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdatePrescription", mock.Anything, mock.AnythingOfType("*entities.Prescription"))
}

func TestUpdatePrescription_BadID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/abc", bytes.NewBufferString(validUpdatePrescriptionJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdatePrescription_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/1", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeletePrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("DeletePrescription", mock.Anything, int64(9)).Return(nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/prescriptions/9", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertCalled(t, "DeletePrescription", mock.Anything, int64(9))
}

func TestDeletePrescription_BadID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/prescriptions/notnum", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// -----------------------------------------------------------------------------
// Lines
// -----------------------------------------------------------------------------
func TestAddLine_Success_PathOverridesPrescripID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("AddLine", mock.Anything, mock.MatchedBy(func(l *entities.PrescriptionLine) bool {
		return l != nil && l.PrescriptionID == 123 && l.PresentationID == 5 && l.DoseAmount == 2
	})).Return(&entities.PrescriptionLine{ID: 777, PrescriptionID: 123}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/123/lines", bytes.NewBufferString(validLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "AddLine", mock.Anything, mock.AnythingOfType("*entities.PrescriptionLine"))
}

func TestAddLine_BadPrescriptionID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/nan/lines", bytes.NewBufferString(validLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddLine_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/1/lines", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "AddLine", mock.Anything, mock.Anything)
}

func TestAddLine_MissingRequired_Validation400(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	// Missing doseAmount (binding:"required")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/1/lines", bytes.NewBufferString(missingRequiredLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "AddLine", mock.Anything, mock.Anything)
}

func TestAddLine_BadEnum_Validation400(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	// scheduleKind not in hour/day/week/month
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/1/lines", bytes.NewBufferString(badEnumLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "AddLine", mock.Anything, mock.Anything)
}

func TestAddLine_BadType_Bind400(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/1/lines", bytes.NewBufferString(badTypeLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddLine_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("AddLine", mock.Anything, mock.AnythingOfType("*entities.PrescriptionLine")).
		Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/1/lines", bytes.NewBufferString(validLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateLine_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("UpdateLine", mock.Anything, mock.MatchedBy(func(l *entities.PrescriptionLine) bool {
		return l != nil && l.ID == 456 && l.PresentationID == 5 && l.DoseAmount == 2
	})).Return(&entities.PrescriptionLine{ID: 456}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/lines/456", bytes.NewBufferString(validLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UpdateLine", mock.Anything, mock.AnythingOfType("*entities.PrescriptionLine"))
}

func TestUpdateLine_BadLineID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/lines/xx", bytes.NewBufferString(validLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateLine_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/lines/1", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "UpdateLine", mock.Anything, mock.Anything)
}

func TestUpdateLine_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("UpdateLine", mock.Anything, mock.AnythingOfType("*entities.PrescriptionLine")).Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/prescriptions/lines/1", bytes.NewBufferString(validLineJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestRemoveLine_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("RemoveLine", mock.Anything, int64(999)).Return(nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/prescriptions/lines/999", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	uc.AssertCalled(t, "RemoveLine", mock.Anything, int64(999))
}

func TestRemoveLine_BadLineID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/prescriptions/lines/notnum", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRemoveLine_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("RemoveLine", mock.Anything, int64(1)).Return(assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/prescriptions/lines/1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// Allocations
// -----------------------------------------------------------------------------
func TestListLineAllocations_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("ListLineAllocations", mock.Anything, int64(33)).
		Return([]entities.LineAllocation{
			{LineID: 33, BatchLocationID: 11, Quantity: 3},
			{LineID: 33, BatchLocationID: 12, Quantity: 2},
		}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/lines/33/allocations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "ListLineAllocations", mock.Anything, int64(33))
}

func TestListLineAllocations_BadLineID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/lines/xx/allocations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestListLineAllocations_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("ListLineAllocations", mock.Anything, int64(33)).Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/prescriptions/lines/33/allocations", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestSetLineAllocations_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("SetLineAllocations",
		mock.Anything,
		int64(44),
		mock.MatchedBy(func(as []entities.LineAllocation) bool {
			if len(as) != 2 {
				return false
			}
			return as[0].BatchLocationID == 11 && as[0].Quantity == 3 &&
				as[1].BatchLocationID == 12 && as[1].Quantity == 2
		}),
	).Return([]entities.LineAllocation{
		{LineID: 44, BatchLocationID: 11, Quantity: 3},
		{LineID: 44, BatchLocationID: 12, Quantity: 2},
	}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/prescriptions/lines/44/allocations", bytes.NewBufferString(validSetAllocationsJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "SetLineAllocations", mock.Anything, int64(44), mock.Anything)
}

func TestSetLineAllocations_BadLineID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/prescriptions/lines/notnum/allocations", bytes.NewBufferString(validSetAllocationsJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSetLineAllocations_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/prescriptions/lines/1/allocations", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "SetLineAllocations", mock.Anything, mock.Anything, mock.Anything)
}

func TestSetLineAllocations_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("SetLineAllocations", mock.Anything, int64(44), mock.Anything).
		Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/prescriptions/lines/44/allocations", bytes.NewBufferString(validSetAllocationsJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// Pack / Unpack
// -----------------------------------------------------------------------------
func TestMarkLinePacked_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("MarkLinePacked", mock.Anything, int64(66), int64(99)).
		Return(&entities.PrescriptionLine{ID: 66, PackedBy: int64Ptr(99)}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/lines/66/pack", bytes.NewBufferString(validPackJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "MarkLinePacked", mock.Anything, int64(66), int64(99))
}

func TestMarkLinePacked_BadLineID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/lines/xx/pack", bytes.NewBufferString(validPackJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkLinePacked_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/lines/1/pack", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "MarkLinePacked", mock.Anything, mock.Anything, mock.Anything)
}

func TestMarkLinePacked_BadBodyType(t *testing.T) {
	var uc mocks.PrescriptionUseCase

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/lines/1/pack", bytes.NewBufferString(badTypePackJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkLinePacked_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("MarkLinePacked", mock.Anything, int64(66), int64(99)).
		Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/lines/66/pack", bytes.NewBufferString(validPackJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUnpackLine_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("UnpackLine", mock.Anything, int64(66)).
		Return(&entities.PrescriptionLine{ID: 66, PackedBy: nil}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/lines/66/unpack", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "UnpackLine", mock.Anything, int64(66))
}

func TestUnpackLine_BadLineID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/lines/xx/unpack", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUnpackLine_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("UnpackLine", mock.Anything, int64(66)).Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/lines/66/unpack", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// Dispense
// -----------------------------------------------------------------------------
func TestDispensePrescription_Success(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("DispensePrescription", mock.Anything, int64(500), int64(77)).
		Return(&entities.Prescription{ID: 500, DispensedBy: int64Ptr(77)}, nil)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/500/dispense", bytes.NewBufferString(validDispenseJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	uc.AssertCalled(t, "DispensePrescription", mock.Anything, int64(500), int64(77))
}

func TestDispensePrescription_BadID(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/notnum/dispense", bytes.NewBufferString(validDispenseJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDispensePrescription_EmptyBody(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/1/dispense", bytes.NewBufferString(emptyJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	uc.AssertNotCalled(t, "DispensePrescription", mock.Anything, mock.Anything, mock.Anything)
}

func TestDispensePrescription_BadBodyType(t *testing.T) {
	var uc mocks.PrescriptionUseCase

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/1/dispense", bytes.NewBufferString(badTypeDispenseJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDispensePrescription_InternalError(t *testing.T) {
	var uc mocks.PrescriptionUseCase
	uc.On("DispensePrescription", mock.Anything, int64(500), int64(77)).
		Return(nil, assert.AnError)

	r := gin.Default()
	newTestPrescriptionRouterNoAuth(r, &uc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/prescriptions/500/dispense", bytes.NewBufferString(validDispenseJSON))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------
func strPtr(s string) *string { return &s }
func int64Ptr(v int64) *int64 { return &v }

// Ensure mocks implement the interface
var _ = func() any {
	var _ entities.PrescriptionUseCase = (*mocks.PrescriptionUseCase)(nil)
	return nil
}()

// Prevent unused import errors when context is not matched literally.
var _ = context.Background()
