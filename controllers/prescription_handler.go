package controllers

import (
	"database/sql"
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

func NewPrescriptionHandler(r gin.IRouter, uc entities.PrescriptionUseCase, secretKey []byte, db *sql.DB) {
	h := &PrescriptionHandler{Usecase: uc}

	grp := r.Group("/prescriptions")
	grp.Use(middleware.AuthRequired(secretKey))
	grp.Use(middleware.WithTx(db))

	// Header CRUD
	grp.GET("", h.ListPrescriptions)
	grp.GET("/:id", h.GetPrescription)
	grp.POST("", h.CreatePrescription)
	grp.PATCH("/:id", h.UpdatePrescription)
	grp.DELETE("/:id", h.DeletePrescription)

	// Lines (one presentation per line)
	grp.POST("/:id/lines", h.AddLine)         // prescription_id in path
	grp.PATCH("/lines/:lineId", h.UpdateLine) // generic updater by lineId
	grp.DELETE("/lines/:lineId", h.RemoveLine)

	// Allocations (reserve/return handled by DB triggers)
	grp.GET("/lines/:lineId/allocations", h.ListLineAllocations)
	grp.PUT("/lines/:lineId/allocations", h.SetLineAllocations) // replace-all

	// Pack / Unpack
	grp.POST("/lines/:lineId/pack", h.MarkLinePacked)
	grp.POST("/lines/:lineId/unpack", h.UnpackLine)

	// Dispense
	grp.POST("/:id/dispense", h.DispensePrescription)
}

// -----------------------------------------------------------------------------
//  CRUD endpoints (header)
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

// -----------------------------------------------------------------------------
//  Lines
// -----------------------------------------------------------------------------

func (h *PrescriptionHandler) AddLine(c *gin.Context) {
	prescriptionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prescription id"})
		return
	}

	var req entities.AddLineReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindErr(c, err)
		return
	}

	line := entities.PrescriptionLine{
		PrescriptionID:       prescriptionID,
		PresentationID:       req.PresentationID,
		Remarks:              req.Remarks,
		DoseAmount:           req.DoseAmount,
		DoseUnit:             req.DoseUnit,
		ScheduleKind:         req.ScheduleKind,
		EveryN:               req.EveryN,
		FrequencyPerSchedule: req.FrequencyPerSchedule,
		Duration:             req.Duration,
		DurationUnit:         req.DurationUnit,
	}

	ctx := c.Request.Context()
	created, err := h.Usecase.AddLine(ctx, &line)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

func (h *PrescriptionHandler) UpdateLine(c *gin.Context) {
	lineID, err := strconv.ParseInt(c.Param("lineId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line id"})
		return
	}

	var req entities.AddLineReq // same payload shape as AddLine
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindErr(c, err)
		return
	}

	line := entities.PrescriptionLine{
		ID:                   lineID,
		PresentationID:       req.PresentationID,
		Remarks:              req.Remarks,
		DoseAmount:           req.DoseAmount,
		DoseUnit:             req.DoseUnit,
		ScheduleKind:         req.ScheduleKind,
		EveryN:               req.EveryN,
		FrequencyPerSchedule: req.FrequencyPerSchedule,
		Duration:             req.Duration,
		DurationUnit:         req.DurationUnit,
	}

	ctx := c.Request.Context()
	updated, err := h.Usecase.UpdateLine(ctx, &line)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *PrescriptionHandler) RemoveLine(c *gin.Context) {
	lineID, err := strconv.ParseInt(c.Param("lineId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line id"})
		return
	}
	ctx := c.Request.Context()
	if err := h.Usecase.RemoveLine(ctx, lineID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
//  Allocations (replace-all)
// -----------------------------------------------------------------------------

func (h *PrescriptionHandler) ListLineAllocations(c *gin.Context) {
	lineID, err := strconv.ParseInt(c.Param("lineId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line id"})
		return
	}
	ctx := c.Request.Context()
	allocs, err := h.Usecase.ListLineAllocations(ctx, lineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, allocs)
}

func (h *PrescriptionHandler) SetLineAllocations(c *gin.Context) {
	lineID, err := strconv.ParseInt(c.Param("lineId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line id"})
		return
	}
	var req entities.SetAllocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		handleBindErr(c, err)
		return
	}
	allocs := make([]entities.LineAllocation, 0, len(req.Allocations))
	for _, a := range req.Allocations {
		allocs = append(allocs, entities.LineAllocation{
			LineID:          lineID, // will be set again in UC/repo but harmless
			BatchLocationID: a.BatchLocationID,
			Quantity:        a.Quantity,
		})
	}

	ctx := c.Request.Context()
	out, err := h.Usecase.SetLineAllocations(ctx, lineID, allocs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// -----------------------------------------------------------------------------
//  Pack / Unpack
// -----------------------------------------------------------------------------

func (h *PrescriptionHandler) MarkLinePacked(c *gin.Context) {
	lineID, err := strconv.ParseInt(c.Param("lineId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line id"})
		return
	}
	ctx := c.Request.Context()
	line, err := h.Usecase.MarkLinePacked(ctx, lineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, line)
}

func (h *PrescriptionHandler) UnpackLine(c *gin.Context) {
	lineID, err := strconv.ParseInt(c.Param("lineId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid line id"})
		return
	}
	ctx := c.Request.Context()
	line, err := h.Usecase.UnpackLine(ctx, lineID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, line)
}

// -----------------------------------------------------------------------------
//  Dispense (no stock mutation here; triggers handled it at allocation time)
// -----------------------------------------------------------------------------

func (h *PrescriptionHandler) DispensePrescription(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	p, err := h.Usecase.DispensePrescription(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, p)
}
