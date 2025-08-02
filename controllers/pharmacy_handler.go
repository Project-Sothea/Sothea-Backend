package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/jieqiboh/sothea_backend/controllers/middleware"
	"github.com/jieqiboh/sothea_backend/entities"
)

// -----------------------------------------------------------------------------
//  Handler struct + constructor
// -----------------------------------------------------------------------------

type PharmacyHandler struct {
	Usecase entities.PharmacyUseCase
}

// NewPharmacyHandler registers /pharmacy/* routes and applies JWT auth.
func NewPharmacyHandler(r *gin.Engine, uc entities.PharmacyUseCase, secretKey []byte) {
	h := &PharmacyHandler{Usecase: uc}

	// NewPharmacyHandler …
	grp := r.Group("/pharmacy")
	grp.Use(middleware.AuthRequired(secretKey))
	{
		// DRUG CATALOG
		grp.GET("/drugs", h.ListDrugs)
		grp.POST("/drugs", h.CreateDrug)
		grp.GET("/drugs/:id", h.GetDrug)
		grp.PATCH("/drugs/:id", h.UpdateDrug)
		grp.DELETE("/drugs/:id", h.DeleteDrug)

		// BATCHES
		grp.GET("/batches", h.ListBatches)
		grp.POST("/batches", h.CreateBatch)
		grp.PATCH("/batches/:id", h.UpdateBatch)
		grp.DELETE("/batches/:id", h.DeleteBatch)
	}
}

// -----------------------------------------------------------------------------
//  Drug endpoints
// -----------------------------------------------------------------------------

func (h *PharmacyHandler) ListDrugs(c *gin.Context) {
	ctx := c.Request.Context()

	drugs, err := h.Usecase.ListDrugs(ctx)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, drugs)
}

func (h *PharmacyHandler) CreateDrug(c *gin.Context) {
	var d entities.Drug
	if err := c.ShouldBindJSON(&d); err != nil {
		handleBindErr(c, err)
		return
	}

	ctx := c.Request.Context()
	drug, err := h.Usecase.CreateDrug(ctx, &d)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, drug)
}

func (h *PharmacyHandler) GetDrug(c *gin.Context) {
	// 1. Parse :id from path
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	detail, err := h.Usecase.GetDrug(ctx, id)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h *PharmacyHandler) UpdateDrug(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var d entities.Drug
	if err := c.ShouldBindJSON(&d); err != nil {
		handleBindErr(c, err)
		return
	}
	d.ID = id // ensure path param wins

	ctx := c.Request.Context()
	drug, err := h.Usecase.CreateDrug(ctx, &d)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, drug)
}

func (h *PharmacyHandler) DeleteDrug(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	if err := h.Usecase.DeleteDrug(ctx, id); err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
//  Batch endpoints
// -----------------------------------------------------------------------------

func (h *PharmacyHandler) ListBatches(c *gin.Context) {
	var drugIDPtr *int64
	if q := c.Query("drug_id"); q != "" {
		val, err := strconv.ParseInt(q, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drug_id"})
			return
		}
		drugIDPtr = &val
	}

	ctx := c.Request.Context()
	batches, err := h.Usecase.ListBatches(ctx, drugIDPtr)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, batches)
}

func (h *PharmacyHandler) CreateBatch(c *gin.Context) {
	var b entities.DrugBatch
	if err := c.ShouldBindJSON(&b); err != nil {
		handleBindErr(c, err)
		return
	}

	ctx := c.Request.Context()
	id, err := h.Usecase.CreateBatch(ctx, &b)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *PharmacyHandler) UpdateBatch(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var b entities.DrugBatch
	if err := c.ShouldBindJSON(&b); err != nil {
		handleBindErr(c, err)
		return
	}
	b.ID = id

	ctx := c.Request.Context()
	if err := h.Usecase.UpdateBatch(ctx, &b); err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *PharmacyHandler) DeleteBatch(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	ctx := c.Request.Context()
	if err := h.Usecase.DeleteBatch(ctx, id); err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
//  Helper functions (copy-style from PatientHandler)
// -----------------------------------------------------------------------------

func handleBindErr(c *gin.Context, err error) {
	if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": ve[0].Error()})
		return
	}
	if err.Error() == "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request body is empty"})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
}

func mapPhErr(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch err {
	case entities.ErrInternalServerError:
		return http.StatusInternalServerError
	case entities.ErrDrugNameTaken:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
