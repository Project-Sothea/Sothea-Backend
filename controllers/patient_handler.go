package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jieqiboh/sothea_backend/controllers/middleware"
	"github.com/jieqiboh/sothea_backend/entities"
	"github.com/jieqiboh/sothea_backend/util"
)

// PatientHandler represent the httphandler for patient
type PatientHandler struct {
	Usecase entities.PatientUseCase
}

// NewPatientHandler will initialize the patients/ resources endpoint
func NewPatientHandler(e *gin.Engine, us entities.PatientUseCase, secretKey []byte) {
	handler := &PatientHandler{
		Usecase: us,
	}

	// Protected routes
	authorized := e.Group("/")
	authorized.Use(middleware.AuthRequired(secretKey))
	{
		authorized.GET("/patient/:id/:vid", handler.GetPatientVisit)
		authorized.GET("/patient/:id/:vid/photo", handler.GetPatientPhoto)
		authorized.POST("/patient", handler.CreatePatient)
		authorized.POST("/patient/:id", handler.CreatePatientVisit)
		authorized.DELETE("/patient/:id/:vid", handler.DeletePatientVisit)
		authorized.PATCH("/patient/:id/:vid", handler.UpdatePatientVisit)
		authorized.GET("/patient-meta/:id", handler.GetPatientMeta)
		authorized.GET("/all-patient-visit-meta/:date", handler.GetAllPatientVisitMeta)
		authorized.GET("/export-db", handler.ExportDatabaseToCSV)
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
	ctx := c.Request.Context()

	// Get the patient by id
	patient, err := p.Usecase.GetPatientVisit(ctx, id, vid)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, patient)
}

// GetPatientPhoto serves the raw image bytes for a patient's visit photo
func (p *PatientHandler) GetPatientPhoto(c *gin.Context) {
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

	// Read from filesystem
	photoPath := util.PatientPhotoPath(id, vid)
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
	if probeLen > 512 { probeLen = 512 }
	mime := http.DetectContentType(data[:probeLen])
	c.Data(http.StatusOK, mime, data)
}


func (p *PatientHandler) CreatePatient(c *gin.Context) {
	ctx := c.Request.Context()
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "use multipart/form-data with fields 'admin' and optional 'photo'"})
		return
	}

	adminJSON := c.PostForm("admin")
	if adminJSON == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin JSON is required"})
		return
	}
	var patientAdmin entities.Admin
	if err := json.Unmarshal([]byte(adminJSON), &patientAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin JSON"})
		return
	}

	// Create the patient (this inserts first visit row with vid=1 via trigger)
	id, err := p.Usecase.CreatePatient(ctx, &patientAdmin)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	// For a new patient, first visit VID is always 1
	vid := int32(1)

	// Optional photo handling
	if data, present, err := readUploadedFile(c, "photo"); err != nil {
		_ = p.Usecase.DeletePatientVisit(ctx, id, vid)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
		return
	} else if present {
		if _, valErr := util.ValidateImageBytes(data); valErr != nil {
			_ = p.Usecase.DeletePatientVisit(ctx, id, vid)
			if valErr.Error() == "file too large" {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": valErr.Error()})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Error()})
			}
			return
		}
		if err := util.SavePatientPhoto(id, vid, data); err != nil {
			_ = p.Usecase.DeletePatientVisit(ctx, id, vid)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store photo"})
			return
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
	ctx := c.Request.Context()

	// Expect multipart/form-data only
	ct := c.GetHeader("Content-Type")
	if !strings.HasPrefix(ct, "multipart/form-data") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "use multipart/form-data with fields 'admin' and optional 'photo'"})
		return
	}

	adminJSON := c.PostForm("admin")
	if adminJSON == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin JSON is required"})
		return
	}
	var patientAdmin entities.Admin
	if err := json.Unmarshal([]byte(adminJSON), &patientAdmin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin JSON"})
		return
	}

	// Create visit first to obtain vid
	vid, err := p.Usecase.CreatePatientVisit(ctx, id32, &patientAdmin)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	// If a photo was included, store to filesystem
	if data, present, err := readUploadedFile(c, "photo"); err != nil {
		_ = p.Usecase.DeletePatientVisit(ctx, id32, int32(vid))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
		return
	} else if present {
		if _, valErr := util.ValidateImageBytes(data); valErr != nil {
			_ = p.Usecase.DeletePatientVisit(ctx, id32, int32(vid))
			if valErr.Error() == "file too large" {
				c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": valErr.Error()})
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Error()})
			}
			return
		}
		if err := util.SavePatientPhoto(id32, int32(vid), data); err != nil {
			_ = p.Usecase.DeletePatientVisit(ctx, id32, int32(vid))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store photo"})
			return
		}
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
	ctx := c.Request.Context()

	err = p.Usecase.DeletePatientVisit(ctx, id32, vid32)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}
	// Also delete the corresponding photo if present
	if delErr := util.DeletePatientPhotoIfExists(id32, vid32); delErr != nil {
		// Log and continue - don't fail the API for file deletion issues
		log.Printf("warning: failed to delete photo for patient %d visit %d: %v", id32, vid32, delErr)
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
	ctx := c.Request.Context()

	ct := c.GetHeader("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		// Allow multipart PATCH with fields: admin (JSON, optional) and photo (file, optional)
		var patient entities.Patient
		adminJSON := c.PostForm("admin")
		if adminJSON != "" {
			var admin entities.Admin
			if err := json.Unmarshal([]byte(adminJSON), &admin); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admin JSON"})
				return
			}
			patient.Admin = &admin
		}

		// Save photo if provided
		if data, present, err := readUploadedFile(c, "photo"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file"})
			return
		} else if present {
			if _, valErr := util.ValidateImageBytes(data); valErr != nil {
				if valErr.Error() == "file too large" {
					c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": valErr.Error()})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": valErr.Error()})
				}
				return
			}
			if err := util.SavePatientPhoto(id32, vid32, data); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store photo"})
				return
			}
		}

		// Apply DB updates only if we have fields to update
		if patient.Admin != nil || patient.PastMedicalHistory != nil || patient.SocialHistory != nil || patient.VitalStatistics != nil || patient.HeightAndWeight != nil || patient.VisualAcuity != nil || patient.FallRisk != nil || patient.Dental != nil || patient.Physiotherapy != nil || patient.DoctorsConsultation != nil {
			if err := p.Usecase.UpdatePatientVisit(ctx, id32, vid32, &patient); err != nil {
				c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
				return
			}
		}

		c.Status(http.StatusOK)
		return
	}

	// Fallback: JSON PATCH behavior (existing clients)
	var patient entities.Patient
	if err := c.ShouldBindJSON(&patient); err != nil {
		if validationErrs, ok := err.(validator.ValidationErrors); ok {
			fieldErr := validationErrs[0]
			c.JSON(http.StatusBadRequest, gin.H{"error": fieldErr.Error()})
			return
		} else if err.Error() == "EOF" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Request Body is empty!"})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	if err := p.Usecase.UpdatePatientVisit(ctx, id32, vid32, &patient); err != nil {
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
	ctx := c.Request.Context()

	patientMeta, err := p.Usecase.GetPatientMeta(ctx, id32)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, patientMeta)
}

func (p *PatientHandler) GetAllPatientVisitMeta(c *gin.Context) {
	dateStr := c.Param("date")

	ctx := c.Request.Context()

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

	patientVisitMeta, err := p.Usecase.GetAllPatientVisitMeta(ctx, date)
	if err != nil {
		c.JSON(getStatusCode(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, patientVisitMeta)
}

func (p *PatientHandler) ExportDatabaseToCSV(c *gin.Context) {
	ctx := c.Request.Context()
	filePath := util.MustGitPath("repository/tmp/output.csv")
	err := p.Usecase.ExportDatabaseToCSV(ctx)
	if err != nil {
		log.Printf("Failed to export data to CSV: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to export data"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/csv")
	// Set the content disposition header to force download
	c.Writer.Header().Set("Content-Disposition", "attachment")

	// Write the contents of the CSV file to the response
	c.FileAttachment(filePath, "output.csv")
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch err {
	case entities.ErrInternalServerError:
		return http.StatusInternalServerError
	case entities.ErrPatientNotFound:
		return http.StatusNotFound
	case entities.ErrPatientVisitNotFound:
		return http.StatusNotFound
	case entities.ErrMissingAdminCategory:
		return http.StatusBadRequest
	case entities.ErrAuthenticationFailed:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
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

