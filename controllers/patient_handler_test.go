package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	entities "github.com/jieqiboh/sothea_backend/entities"
	"github.com/jieqiboh/sothea_backend/mocks"

	"github.com/stretchr/testify/assert"
)

var admin = entities.Admin{
	FamilyGroup:   entities.PtrTo("S001"),
	RegDate:       entities.PtrTo(time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC)),
	Name:          entities.PtrTo("John Doe"),
	KhmerName:     entities.PtrTo("១២៣៤ ៥៦៧៨៩០ឥឲ"),
	Dob:           entities.PtrTo(time.Date(1994, time.January, 10, 0, 0, 0, 0, time.UTC)),
	Age:           entities.PtrTo(30),
	Gender:        entities.PtrTo("M"),
	Village:       entities.PtrTo("SO"),
	ContactNo:     entities.PtrTo("12345678"),
	Pregnant:      entities.PtrTo(false),
	DrugAllergies: entities.PtrTo("panadol"),
	SentToID:      entities.PtrTo(false),
}
var pastmedicalhistory = entities.PastMedicalHistory{
	Tuberculosis:               entities.PtrTo(true),
	Diabetes:                   entities.PtrTo(false),
	Hypertension:               entities.PtrTo(true),
	Hyperlipidemia:             entities.PtrTo(false),
	ChronicJointPains:          entities.PtrTo(false),
	ChronicMuscleAches:         entities.PtrTo(true),
	SexuallyTransmittedDisease: entities.PtrTo(true),
	SpecifiedSTDs:              entities.PtrTo("TRICHOMONAS"),
	Others:                     nil,
}
var socialhistory = entities.SocialHistory{
	PastSmokingHistory:    entities.PtrTo(true),
	NumberOfYears:         entities.PtrTo(int32(15)),
	CurrentSmokingHistory: entities.PtrTo(false),
	CigarettesPerDay:      nil,
	AlcoholHistory:        entities.PtrTo(true),
	HowRegular:            entities.PtrTo("A"),
}
var vitalstatistics = entities.VitalStatistics{
	Temperature:             entities.PtrTo(36.5),
	SpO2:                    entities.PtrTo(98.0),
	SystolicBP1:             entities.PtrTo(120.0),
	DiastolicBP1:            entities.PtrTo(80.0),
	SystolicBP2:             entities.PtrTo(122.0),
	DiastolicBP2:            entities.PtrTo(78.0),
	AverageSystolicBP:       entities.PtrTo(121.0),
	AverageDiastolicBP:      entities.PtrTo(79.0),
	HR1:                     entities.PtrTo(72.0),
	HR2:                     entities.PtrTo(71.0),
	AverageHR:               entities.PtrTo(71.5),
	RandomBloodGlucoseMmolL: entities.PtrTo(5.4),
}
var heightandweight = entities.HeightAndWeight{
	Height:      entities.PtrTo(170.0),
	Weight:      entities.PtrTo(70.0),
	BMI:         entities.PtrTo(24.2),
	BMIAnalysis: entities.PtrTo("normal weight"),
	PaedsHeight: entities.PtrTo(90.0),
	PaedsWeight: entities.PtrTo(80.0),
}
var visualacuity = entities.VisualAcuity{
	LEyeVision:             entities.PtrTo(int32(20)),
	REyeVision:             entities.PtrTo(int32(20)),
	AdditionalIntervention: entities.PtrTo("VISUAL FIELD TEST REQUIRED"),
}
var doctorsconsultation = entities.DoctorsConsultation{
	Well:              entities.PtrTo(true),
	Msk:               entities.PtrTo(false),
	Cvs:               entities.PtrTo(false),
	Respi:             entities.PtrTo(true),
	Gu:                entities.PtrTo(true),
	Git:               entities.PtrTo(false),
	Eye:               entities.PtrTo(true),
	Derm:              entities.PtrTo(false),
	Others:            entities.PtrTo("TRICHOMONAS VAGINALIS"),
	ConsultationNotes: entities.PtrTo("CHEST PAIN, SHORTNESS OF BREATH, COUGH"),
	Diagnosis:         entities.PtrTo("ACUTE BRONCHITIS"),
	Treatment:         entities.PtrTo("REST, HYDRATION, COUGH SYRUP"),
	ReferralNeeded:    entities.PtrTo(false),
	ReferralLoc:       nil,
	Remarks:           entities.PtrTo("MONITOR FOR RESOLUTION"),
}

var patient = entities.Patient{
	Admin:               &admin,
	PastMedicalHistory:  &pastmedicalhistory,
	SocialHistory:       &socialhistory,
	VitalStatistics:     &vitalstatistics,
	HeightAndWeight:     &heightandweight,
	VisualAcuity:        &visualacuity,
	DoctorsConsultation: &doctorsconsultation,
}

// NewPatientHandler will initialize the patients/ resources endpoint, excluding the auth middleware
func newTestPatientHandler(e *gin.Engine, us entities.PatientUseCase) {
	handler := &PatientHandler{
		Usecase: us,
	}

	e.GET("/patient/:id/:vid", handler.GetPatientVisit)
	e.POST("/patient", handler.CreatePatient)
	e.POST("/patient/:id", handler.CreatePatientVisit)
	e.DELETE("/patient/:id/:vid", handler.DeletePatientVisit)
	e.PATCH("/patient/:id/:vid", handler.UpdatePatientVisit)
	e.GET("/patient-meta/:id", handler.GetPatientMeta)
	e.GET("/all-patient-visit-meta/:date", handler.GetAllPatientVisitMeta)
	e.GET("/export-db", handler.ExportDatabaseToCSV)
}

// Success - 200 OK
// Bad Request (id or vid is not a number) - 400 Bad Request
// Patient Visit Not Found - 404 Not Found
func TestGetPatientVisit_Success(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := 1
	vid := 1
	mockUsecase.On("GetPatientVisit", context.Background(), int32(id), int32(vid)).Return(&patient, nil)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/patient/1/1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGetPatientVisit_BadRequest(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := 1
	wrongId := "hello"
	vid := 1
	wrongVid := "world"
	mockUsecase.On("GetPatientVisit", context.Background(), int32(id), wrongVid).Return(nil, entities.ErrPatientVisitNotFound)
	mockUsecase.On("GetPatientVisit", context.Background(), wrongId, int32(vid)).Return(nil, entities.ErrPatientVisitNotFound)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/patient/1/"+wrongVid, nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	req, _ = http.NewRequest("GET", "/patient/"+wrongId+"/1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestGetPatientVisit_NotFound(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	num := -1
	mockUsecase.On("GetPatientVisit", context.Background(), int32(num)).Return(nil, entities.ErrPatientNotFound)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/patient/-1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

// Success - 200 OK
// Empty Request Body - 400 Bad Request
// Invalid Parameters - 400 Bad Request
// Json Marshalling Error - 400 Bad Request
func TestCreatePatient_Success(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := int32(7)
	mockUsecase.On("CreatePatient", context.Background(), mocks.ValidPatient.Admin).Return(id, nil)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient", strings.NewReader(mocks.ValidPatientAdminJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCreatePatient_EmptyRequestBody_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient", strings.NewReader(""))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestCreatePatient_InvalidParameters_Failure(t *testing.T) {
	// When ShouldBindJSON fails - e.g. A required field is not present
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient", strings.NewReader(mocks.InvalidParametersAdminJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	req, _ = http.NewRequest("POST", "/patient", strings.NewReader("{}"))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestCreatePatient_JSONMarshallingError_Failure(t *testing.T) {
	// When marshalling into a JSON struct fails - e.g. data type mismatch between the JSON and the expected golang struct field type
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient", strings.NewReader(mocks.JSONMarshallingErrorAdminJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Success - 200 OK
// Patient Not Found - 404 Not Found
// Bad Request (id is not a number) - 400 Bad Request
// Empty Request Body - 400 Bad Request
// Invalid Parameters - 400 Bad Request
// Json Marshalling Error - 400 Bad Request
func TestCreatePatientVisit_Success(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := int32(1)
	vid := int32(1)
	mockUsecase.On("CreatePatientVisit", context.Background(), id, mocks.ValidPatient.Admin).Return(vid, nil)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient/"+strconv.Itoa(int(id)), strings.NewReader(mocks.ValidPatientAdminJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestCreatePatientVisit_BadRequest(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	wrongId := "hello"
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient/"+wrongId, nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestCreatePatientVisit_EmptyRequestBody_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := int32(1)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient/"+strconv.Itoa(int(id)), strings.NewReader(""))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestCreatePatientVisit_InvalidParameters_Failure(t *testing.T) {
	// When ShouldBindJSON fails - e.g. A required field is not present
	var mockUsecase mocks.PatientUseCase
	id := int32(1)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient/"+strconv.Itoa(int(id)), strings.NewReader(mocks.InvalidParametersAdminJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestCreatePatientVisit_JSONMarshallingError_Failure(t *testing.T) {
	// When marshalling into a JSON struct fails - e.g. data type mismatch between the JSON and the expected golang struct field type
	var mockUsecase mocks.PatientUseCase
	id := int32(1)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/patient/"+strconv.Itoa(int(id)), strings.NewReader(mocks.JSONMarshallingErrorAdminJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Success - 200 OK
// Bad Request (id is not a number) - 400 Bad Request
// Patient Visit Not Found - 404 Not Found
func TestDeletePatientVisit_Success(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := 1
	vid := 1
	mockUsecase.On("DeletePatientVisit", context.Background(), int32(id), int32(vid)).Return(nil)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/patient/"+strconv.Itoa(id)+"/"+strconv.Itoa(vid), nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestDeletePatientByID_BadRequest_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := 1
	wrongId := "hello"
	vid := 1
	wrongVid := "world"

	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/patient/"+wrongId+"/"+strconv.Itoa(vid), nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	req, _ = http.NewRequest("DELETE", "/patient/"+strconv.Itoa(id)+"/"+wrongVid, nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Success - 200 OK
// Patient Not Found - 404 Not Found
// Bad Request (id or vid is not a number) - 400 Bad Request
// Invalid Parameters - 400 Bad Request
// Json Marshalling Error - 400 Bad Request
// Empty Request Body - 400 Bad Request
func TestUpdatePatientVisit_Success(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := int32(1)
	vid := int32(1)
	mockUsecase.On("UpdatePatientVisit", context.Background(), id, vid, &mocks.MissingAdminPatient).Return(nil)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/patient/"+strconv.Itoa(int(id))+"/"+strconv.Itoa(int(vid)), strings.NewReader(mocks.MissingAdminPatientJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestUpdatePatientVisit_BadRequest_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := int32(1)
	wrongId := "hello"
	vid := int32(1)
	wrongVid := "world"
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/patient/"+wrongId+"/"+strconv.Itoa(int(vid)), strings.NewReader(mocks.MissingAdminPatientJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)

	req, _ = http.NewRequest("PATCH", "/patient/"+strconv.Itoa(int(id))+"/"+wrongVid, strings.NewReader(mocks.MissingAdminPatientJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestUpdatePatientVisit_NotFound_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	mockUsecase.On("UpdatePatientVisit", context.Background(), int32(-1), int32(1), &mocks.MissingAdminPatient).Return(entities.ErrPatientVisitNotFound)
	mockUsecase.On("UpdatePatientVisit", context.Background(), int32(1), int32(-1), &mocks.MissingAdminPatient).Return(entities.ErrPatientVisitNotFound)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/patient/-1/1", strings.NewReader(mocks.MissingAdminPatientJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)

	req, _ = http.NewRequest("PATCH", "/patient/1/-1", strings.NewReader(mocks.MissingAdminPatientJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestUpdatePatientVisit_InvalidParameters_Failure(t *testing.T) {
	// When ShouldBindJSON fails - e.g. A required field is not present
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/patient/1/1", strings.NewReader(mocks.InvalidParametersPatientJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestUpdatePatientVisit_JSONMarshallingError_Failure(t *testing.T) {
	// When marshalling into a JSON struct fails - e.g. data type mismatch between the JSON and the expected golang struct field type
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/patient/1/1", strings.NewReader(mocks.JSONMarshallingErrorPatientJson))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

func TestUpdatePatientVisit_EmptyRequestBody_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PATCH", "/patient/1/1", strings.NewReader(""))

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Success - 200 OK
// Patient Not Found - 404 Not Found
// Bad Request (id or vid is not a number) - 400 Bad Request
func TestGetPatientMeta_Success(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := 1
	mockUsecase.On("GetPatientMeta", context.Background(), int32(id)).Return(&mocks.ValidPatientMeta, nil)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/patient-meta/1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGetPatientMeta_NotFound_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	id := -1
	mockUsecase.On("GetPatientMeta", context.Background(), int32(id)).Return(nil, entities.ErrPatientNotFound)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/patient-meta/-1", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestGetPatientMeta_BadRequest_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/patient-meta/hello", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Success - 200 OK
// Bad Request (id is not a number) - 400 Bad Request
func TestGetAllPatientVisitMeta_Success(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	date := time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC)
	mockUsecase.On("GetAllPatientVisitMeta", context.Background(), date).Return(mocks.ValidPatientVisitMetaArray, nil)
	mockUsecase.On("GetAllPatientVisitMeta", context.Background(), time.Time{}).Return(mocks.ValidPatientVisitMetaArray, nil)
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/all-patient-visit-meta/2024-01-10", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	req, _ = http.NewRequest("GET", "/all-patient-visit-meta/default", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestGetAllPatientVisitMeta_BadRequest_Failure(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/all-patient-visit-meta/hello", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, 400, w.Code)
}

// Success - 200 OK
func TestExportDatabaseToCSV_Success(t *testing.T) {
	var mockUsecase mocks.PatientUseCase
	mockUsecase.On("ExportDatabaseToCSV", context.Background()).Return(nil)

	router := gin.Default()
	newTestPatientHandler(router, &mockUsecase)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/export-db", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	mockUsecase.AssertCalled(t, "ExportDatabaseToCSV", context.Background())
}
