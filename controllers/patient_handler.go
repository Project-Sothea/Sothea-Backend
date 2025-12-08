package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"sothea-backend/controllers/middleware"
	"sothea-backend/entities"
	db "sothea-backend/repository/sqlc"
	"sothea-backend/usecases"
	"sothea-backend/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// PatientHandler represent the httphandler for patient
type PatientHandler struct {
	Usecase *usecases.PatientUsecase
}

// NewPatientHandler will initialize the patients/ resources endpoint
func NewPatientHandler(r gin.IRouter, uc *usecases.PatientUsecase, secretKey []byte) {
	handler := &PatientHandler{
		Usecase: uc,
	}

	// Protected routes
	authorized := r.Group("/")
	authorized.Use(middleware.AuthRequired(secretKey))
	{
		authorized.GET("/patient/:id/:vid", handler.GetPatientVisit)
		authorized.GET("/patient/:id/photo", handler.GetPatientPhoto)
		authorized.POST("/patient", handler.CreatePatient)
		authorized.POST("/patient-with-visit", handler.CreatePatientWithVisit)
		authorized.PUT("/patient/:id", handler.UpdatePatient)
		authorized.DELETE("/patient/:id", handler.DeletePatient)
		authorized.POST("/patient/:id", handler.CreatePatientVisit)
		authorized.PATCH("/patient/:id/:vid", handler.UpdatePatientVisit)
		authorized.DELETE("/patient/:id/:vid", handler.DeletePatientVisit)
		authorized.GET("/patient-meta/:id", handler.GetPatientMeta)
		authorized.GET("/all-patient-visit-meta/:date", handler.GetAllPatientVisitMeta)
	}
}

func (p *PatientHandler) GetPatientVisit(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	vidP, err := strconv.Atoi(c.Param("vid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := int32(idP)
	vid := int32(vidP)
	patient, err := p.Usecase.GetPatientVisit(c.Request.Context(), id, vid)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, patient)
}

// GetPatientPhoto serves the raw image bytes for a patient's photo
func (p *PatientHandler) GetPatientPhoto(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := int32(idP)

	// Read from filesystem
	photoPath := util.PatientPhotoPath(id)
	data, err := os.ReadFile(photoPath)
	if os.IsNotExist(err) {
		c.Status(http.StatusNotFound)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read photo"})
		return
	}

	// Detect content type using first up to 512 bytes
	probeLen := len(data)
	if probeLen > 512 {
		probeLen = 512
	}
	mime := http.DetectContentType(data[:probeLen])
	c.Data(http.StatusOK, mime, data)
}

func (p *PatientHandler) CreatePatient(c *gin.Context) {
	ct := c.GetHeader("Content-Type")
	var patientProfile db.PatientDetail

	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		patientJSON := c.PostForm("patient_details")
		if patientJSON == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "patient_details JSON is required"})
			return
		}
		if err := json.Unmarshal([]byte(patientJSON), &patientProfile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient_details JSON"})
			return
		}
		photoBytes, present, err := readUploadedFile(c, "photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
			return
		}
		if present {
			if _, valErr := util.ValidateImageBytes(photoBytes); valErr != nil {
				if valErr.Error() == "file too large" {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": valErr.Error()})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Error()})
				}
				return
			}
			c.Set("uploadedPhoto", photoBytes)
		}
	case strings.HasPrefix(ct, "application/json"):
		if err := c.ShouldBindJSON(&patientProfile); err != nil {
			var validationErrs validator.ValidationErrors
			if errors.As(err, &validationErrs) {
				fieldErr := validationErrs[0]
				c.JSON(http.StatusBadRequest, gin.H{"error": fieldErr.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	default:
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "use application/json or multipart/form-data with field 'patient'"})
		return
	}

	id, err := p.Usecase.CreatePatient(c.Request.Context(), &patientProfile)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	if val, exists := c.Get("uploadedPhoto"); exists {
		if data, ok := val.([]byte); ok {
			if err := util.SavePatientPhoto(id, data); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store photo"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"id": id})
}

// CreatePatientWithVisit creates patient + first visit atomically and returns both ids.
func (p *PatientHandler) CreatePatientWithVisit(c *gin.Context) {
	ct := c.GetHeader("Content-Type")

	type request struct {
		PatientDetails db.PatientDetail `json:"patient_details"`
		Admin          db.Admin         `json:"admin"`
	}

	var patientProfile db.PatientDetail
	var admin db.Admin

	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		patientJSON := c.PostForm("patient_details")
		if patientJSON == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "patient_details JSON is required"})
			return
		}
		adminJSON := c.PostForm("admin")
		if adminJSON == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "admin JSON is required"})
			return
		}
		if err := json.Unmarshal([]byte(patientJSON), &patientProfile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient_details JSON"})
			return
		}
		if err := json.Unmarshal([]byte(adminJSON), &admin); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin JSON"})
			return
		}
		photoBytes, present, err := readUploadedFile(c, "photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
			return
		}
		if present {
			if _, valErr := util.ValidateImageBytes(photoBytes); valErr != nil {
				if valErr.Error() == "file too large" {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": valErr.Error()})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Error()})
				}
				return
			}
			c.Set("uploadedPhoto", photoBytes)
		}
	case strings.HasPrefix(ct, "application/json"):
		var req request
		if err := c.ShouldBindJSON(&req); err != nil {
			var validationErrs validator.ValidationErrors
			if errors.As(err, &validationErrs) {
				fieldErr := validationErrs[0]
				c.JSON(http.StatusBadRequest, gin.H{"error": fieldErr.Error()})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		patientProfile = req.PatientDetails
		admin = req.Admin
	default:
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "use application/json or multipart/form-data with fields 'patient_details' and 'admin'"})
		return
	}

	id, vid, err := p.Usecase.CreatePatientWithVisit(c.Request.Context(), &patientProfile, &admin)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	if val, exists := c.Get("uploadedPhoto"); exists {
		if data, ok := val.([]byte); ok {
			if err := util.SavePatientPhoto(id, data); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store photo"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "vid": vid})
}

func (p *PatientHandler) CreatePatientVisit(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id32 := int32(idP)
	var patientAdmin db.Admin
	if err := c.ShouldBindJSON(&patientAdmin); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			fieldErr := validationErrs[0]
			c.JSON(http.StatusBadRequest, gin.H{"error": fieldErr.Error()})
			return
		}
	}

	// Create visit first to obtain vid
	vid, err := p.Usecase.CreatePatientVisit(c.Request.Context(), id32, &patientAdmin)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"vid": vid})
}

func (p *PatientHandler) DeletePatientVisit(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	vidP, err := strconv.Atoi(c.Param("vid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id32 := int32(idP)
	vid32 := int32(vidP)

	err = p.Usecase.DeletePatientVisit(c.Request.Context(), id32, vid32)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (p *PatientHandler) UpdatePatientVisit(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	vidP, err := strconv.Atoi(c.Param("vid"))
	// Check if the id or vid is not a number
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id32 := int32(idP)
	vid32 := int32(vidP)

	// JSON PATCH only
	var patient entities.Patient
	if err := c.ShouldBindJSON(&patient); err != nil {
		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			fieldErr := validationErrs[0]
			c.JSON(http.StatusBadRequest, gin.H{"error": fieldErr.Error()})
			return
		}
	}

	if err := p.Usecase.UpdatePatientVisit(c.Request.Context(), id32, vid32, &patient); err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (p *PatientHandler) GetPatientMeta(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id32 := int32(idP)
	patientMeta, err := p.Usecase.GetPatientMeta(c.Request.Context(), id32)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, patientMeta)
}

func (p *PatientHandler) GetAllPatientVisitMeta(c *gin.Context) {
	dateStr := c.Param("date")

	var date time.Time
	var err error
	if dateStr == "default" {
		date = time.Time{}
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid date format: %s", dateStr)})
			return
		}
	}

	patientVisitMeta, err := p.Usecase.GetAllPatientVisitMeta(c.Request.Context(), date)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, patientVisitMeta)
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	case errors.Is(err, entities.ErrInternalServerError):
		return http.StatusInternalServerError
	case errors.Is(err, entities.ErrPatientNotFound):
		return http.StatusNotFound
	case errors.Is(err, entities.ErrPatientVisitNotFound):
		return http.StatusNotFound
	case errors.Is(err, entities.ErrMissingPatientData):
		return http.StatusBadRequest
	case errors.Is(err, entities.ErrMissingAdminCategory):
		return http.StatusBadRequest
	case errors.Is(err, entities.ErrAuthenticationFailed):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// UpdatePatient updates demographics for a patient (no visit fields here)
func (p *PatientHandler) UpdatePatient(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := int32(idP)
	var patientProfile db.PatientDetail
	ct := c.GetHeader("Content-Type")

	switch {
	case strings.HasPrefix(ct, "multipart/form-data"):
		patientJSON := c.PostForm("patient_details")
		if patientJSON == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "patient_details JSON is required"})
			return
		}
		if err := json.Unmarshal([]byte(patientJSON), &patientProfile); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient_details JSON"})
			return
		}
		photoBytes, present, err := readUploadedFile(c, "photo")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
			return
		}
		if present {
			if _, valErr := util.ValidateImageBytes(photoBytes); valErr != nil {
				if valErr.Error() == "file too large" {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": valErr.Error()})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Error()})
				}
				return
			}
			if err := util.SavePatientPhoto(id, photoBytes); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store photo"})
				return
			}
		}
	case strings.HasPrefix(ct, "application/json"):
		if err := c.ShouldBindJSON(&patientProfile); err != nil {
			var validationErrs validator.ValidationErrors
			if errors.As(err, &validationErrs) {
				fieldErr := validationErrs[0]
				c.JSON(http.StatusBadRequest, gin.H{"error": fieldErr.Error()})
				return
			} else if err == io.EOF {
				c.JSON(http.StatusBadRequest, gin.H{"error": "patient_details JSON is required"})
				return
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
	default:
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "use application/json or multipart/form-data with field 'patient_details'"})
		return
	}

	if err := p.Usecase.UpdatePatient(c.Request.Context(), id, &patientProfile); err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// DeletePatient removes a patient and all associated visits/data
func (p *PatientHandler) DeletePatient(c *gin.Context) {
	idP, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id := int32(idP)

	if err := p.Usecase.DeletePatient(c.Request.Context(), id); err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	if err := util.DeletePatientPhotoIfExists(id); err != nil {
		log.Printf("warning: failed to delete photo for patient %d: %v", id, err)
	}

	c.Status(http.StatusOK)
}

// readUploadedFile reads a single file field from a multipart/form-data request.
// It returns (data, present, error). When the field is missing, present=false and error=nil.
func readUploadedFile(c *gin.Context, field string) ([]byte, bool, error) {
	fh, err := c.FormFile(field)
	if err != nil || fh == nil {
		return nil, false, nil
	}
	f, openErr := fh.Open()
	if openErr != nil {
		return nil, true, openErr
	}
	defer f.Close()
	data, readErr := io.ReadAll(f)
	if readErr != nil {
		return nil, true, readErr
	}
	return data, true, nil
}
