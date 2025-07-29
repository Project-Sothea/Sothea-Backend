package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/jieqiboh/sothea_backend/controllers/middleware"
	"github.com/jieqiboh/sothea_backend/entities"
)

// -----------------------------------------------------------------------------
//  Handler struct + constructor
// -----------------------------------------------------------------------------

type PrescriptionHandler struct {
	Usecase entities.PrescriptionUseCase
}

func NewPrescriptionHandler(r *gin.Engine, uc entities.PrescriptionUseCase, secretKey []byte) {
	h := &PrescriptionHandler{Usecase: uc}

	grp := r.Group("/prescriptions")
	grp.Use(middleware.AuthRequired(secretKey))
	{
		grp.GET("", h.ListPrescriptions)
		grp.POST("", h.CreatePrescription)
		grp.GET(":id", h.GetPrescription)
		grp.PATCH(":id", h.UpdatePrescription)
		grp.DELETE(":id", h.DeletePrescription)
	}
}

// -----------------------------------------------------------------------------
//  CRUD endpoints
// -----------------------------------------------------------------------------

func (h *PrescriptionHandler) ListPrescriptions(c *gin.Context) {
	var patientIDPtr *int64
	var vidPtr *int32

	if q := c.Query("patient_id"); q != "" {
		val, err := strconv.ParseInt(q, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid patient_id"})
			return
		}
		patientIDPtr = &val
	}

	if q := c.Query("vid"); q != "" {
		val, err := strconv.ParseInt(q, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vid"})
			return
		}
		tmp := int32(val)
		vidPtr = &tmp
	}

	ctx := c.Request.Context()
	prescriptions, err := h.Usecase.ListPrescriptions(ctx, patientIDPtr, vidPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prescriptions)
}

func (h *PrescriptionHandler) CreatePrescription(c *gin.Context) {
	var p entities.Prescription
	if err := c.ShouldBindJSON(&p); err != nil {
		handleBindErr(c, err)
		return
	}

	ctx := c.Request.Context()
	prescription, err := h.Usecase.CreatePrescription(ctx, &p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prescription)
}

func (h *PrescriptionHandler) GetPrescription(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	fmt.Println("HEHEHEHHE")
	ctx := c.Request.Context()
	prescription, err := h.Usecase.GetPrescriptionByID(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prescription)
}

func (h *PrescriptionHandler) UpdatePrescription(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var p entities.Prescription
	if err := c.ShouldBindJSON(&p); err != nil {
		handleBindErr(c, err)
		return
	}
	p.ID = id

	ctx := c.Request.Context()
	prescription, err := h.Usecase.UpdatePrescription(ctx, &p)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, prescription)
}

func (h *PrescriptionHandler) DeletePrescription(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	if err := h.Usecase.DeletePrescription(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
