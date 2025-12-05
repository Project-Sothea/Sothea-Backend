package controllers

import (
	"database/sql"
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

// Registers /pharmacy/* routes and applies JWT + Tx middlewares.
func NewPharmacyHandler(r gin.IRouter, uc entities.PharmacyUseCase, secretKey []byte, db *sql.DB) {
	h := &PharmacyHandler{Usecase: uc}

	grp := r.Group("/pharmacy")
	grp.Use(middleware.AuthRequired(secretKey))
	grp.Use(middleware.WithTx(db))

	// ---------------- DRUGS ----------------
	grp.GET("/drugs", h.ListDrugs) // ?q=parac
	grp.POST("/drugs", h.CreateDrug)
	grp.GET("/drugs/:drugId", h.GetDrugStock)
	grp.PATCH("/drugs/:drugId", h.UpdateDrug)
	grp.DELETE("/drugs/:drugId", h.DeleteDrug)

	// ---------------- BATCHES ---------------
	grp.GET("/drugs/:drugId/batches", h.ListBatches)
	grp.POST("/drugs/:drugId/batches", h.CreateBatch)

	grp.GET("/batches", h.ListAllBatches)

	grp.GET("/batches/:batchId", h.GetBatch)
	grp.PATCH("/batches/:batchId", h.UpdateBatch)
	grp.DELETE("/batches/:batchId", h.DeleteBatch)

	// ------------- LOCATIONS ----------------
	grp.GET("/batches/:batchId/locations", h.ListBatchLocations)
	grp.POST("/batches/:batchId/locations", h.CreateBatchLocation)

	// independent update/delete by locationId
	grp.PATCH("/locations/:locationId", h.UpdateBatchLocation)
	grp.DELETE("/locations/:locationId", h.DeleteBatchLocation)
}

// -----------------------------------------------------------------------------
//  DRUG endpoints
// -----------------------------------------------------------------------------

func (h *PharmacyHandler) ListDrugs(c *gin.Context) {
	ctx := c.Request.Context()
	var qPtr *string
	if q := c.Query("q"); q != "" {
		qPtr = &q
	}
	drugs, err := h.Usecase.ListDrugs(ctx, qPtr)
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
	id, err := strconv.ParseInt(c.Param("drugId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drugId"})
		return
	}
	ctx := c.Request.Context()
	drug, err := h.Usecase.GetDrug(ctx, id)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, drug)
}

func (h *PharmacyHandler) GetDrugStock(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("drugId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drugId"})
		return
	}
	ctx := c.Request.Context()
	stock, err := h.Usecase.GetDrugStock(ctx, id)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stock)
}

func (h *PharmacyHandler) UpdateDrug(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("drugId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drugId"})
		return
	}
	var d entities.Drug
	if err := c.ShouldBindJSON(&d); err != nil {
		handleBindErr(c, err)
		return
	}
	d.ID = id
	ctx := c.Request.Context()
	drug, err := h.Usecase.UpdateDrug(ctx, &d)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, drug)
}

func (h *PharmacyHandler) DeleteDrug(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("drugId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drugId"})
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
//  BATCH endpoints (scoped by drugId where creating/listing)
// -----------------------------------------------------------------------------

func (h *PharmacyHandler) ListBatches(c *gin.Context) {
	drugID, err := strconv.ParseInt(c.Param("drugId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drugId"})
		return
	}
	ctx := c.Request.Context()
	batches, err := h.Usecase.ListBatches(ctx, drugID)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, batches)
}

func (h *PharmacyHandler) CreateBatch(c *gin.Context) {
	drugID, err := strconv.ParseInt(c.Param("drugId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drugId"})
		return
	}
	var body struct {
		Batch     entities.DrugBatch           `json:"batch"`
		Locations []entities.DrugBatchLocation `json:"locations"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		handleBindErr(c, err)
		return
	}
	body.Batch.DrugID = drugID

	ctx := c.Request.Context()
	detail, err := h.Usecase.CreateBatch(ctx, &body.Batch, body.Locations)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

// ListAllBatches aggregates batches for all drugs.
// Route: GET /pharmacy/batches
func (h *PharmacyHandler) ListAllBatches(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all drugs
	drugs, err := h.Usecase.ListDrugs(ctx, nil)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}

	all := make([]entities.BatchDetail, 0, 128)

	for _, d := range drugs {
		batches, err := h.Usecase.ListBatches(ctx, d.ID)
		if err != nil {
			c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
			return
		}
		all = append(all, batches...)
	}

	c.JSON(http.StatusOK, all)
}

func (h *PharmacyHandler) GetBatch(c *gin.Context) {
	batchID, err := strconv.ParseInt(c.Param("batchId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batchId"})
		return
	}
	ctx := c.Request.Context()
	detail, err := h.Usecase.GetBatch(ctx, batchID)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *PharmacyHandler) UpdateBatch(c *gin.Context) {
	batchID, err := strconv.ParseInt(c.Param("batchId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batchId"})
		return
	}
	var b entities.DrugBatch
	if err := c.ShouldBindJSON(&b); err != nil {
		handleBindErr(c, err)
		return
	}
	b.ID = batchID

	ctx := c.Request.Context()
	detail, err := h.Usecase.UpdateBatch(ctx, &b)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *PharmacyHandler) DeleteBatch(c *gin.Context) {
	batchID, err := strconv.ParseInt(c.Param("batchId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batchId"})
		return
	}
	ctx := c.Request.Context()
	if err := h.Usecase.DeleteBatch(ctx, batchID); err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// -----------------------------------------------------------------------------
//  LOCATION endpoints
// -----------------------------------------------------------------------------

func (h *PharmacyHandler) ListBatchLocations(c *gin.Context) {
	batchID, err := strconv.ParseInt(c.Param("batchId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batchId"})
		return
	}
	ctx := c.Request.Context()
	locs, err := h.Usecase.ListBatchLocations(ctx, batchID)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, locs)
}

func (h *PharmacyHandler) CreateBatchLocation(c *gin.Context) {
	batchID, err := strconv.ParseInt(c.Param("batchId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batchId"})
		return
	}
	var loc entities.DrugBatchLocation
	if err := c.ShouldBindJSON(&loc); err != nil {
		handleBindErr(c, err)
		return
	}
	loc.BatchID = batchID

	ctx := c.Request.Context()
	created, err := h.Usecase.CreateBatchLocation(ctx, &loc)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, created)
}

func (h *PharmacyHandler) UpdateBatchLocation(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("locationId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid locationId"})
		return
	}
	var loc entities.DrugBatchLocation
	if err := c.ShouldBindJSON(&loc); err != nil {
		handleBindErr(c, err)
		return
	}
	loc.ID = locationID

	ctx := c.Request.Context()
	updated, err := h.Usecase.UpdateBatchLocation(ctx, &loc)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *PharmacyHandler) DeleteBatchLocation(c *gin.Context) {
	locationID, err := strconv.ParseInt(c.Param("locationId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid locationId"})
		return
	}
	ctx := c.Request.Context()
	if err := h.Usecase.DeleteBatchLocation(ctx, locationID); err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

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
