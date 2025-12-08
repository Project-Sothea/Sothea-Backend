package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"

	"sothea-backend/controllers/middleware"
	"sothea-backend/entities"
	"sothea-backend/repository/postgres"
	db "sothea-backend/repository/sqlc"
	"sothea-backend/usecases"
)

// PharmacyHandler wires HTTP routes to the pharmacy usecase.
type PharmacyHandler struct {
	Usecase *usecases.PharmacyUsecase
}

// Registers /pharmacy/* routes and applies JWT + Tx middlewares.
func NewPharmacyHandler(r gin.IRouter, uc *usecases.PharmacyUsecase, secretKey []byte, pool *pgxpool.Pool) {
	h := &PharmacyHandler{Usecase: uc}

	grp := r.Group("/pharmacy")
	grp.Use(middleware.AuthRequired(secretKey))
	grp.Use(middleware.WithTx(pool))

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

// ---------------- DRUGS -----------------

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
	var d db.Drug
	if err := c.ShouldBindJSON(&d); err != nil {
		handleBindErr(c, err)
		return
	}
	drug, err := h.Usecase.CreateDrug(c.Request.Context(), &d)
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
	drug, err := h.Usecase.GetDrug(c.Request.Context(), id)
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
	stock, err := h.Usecase.GetDrugStock(c.Request.Context(), id)
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
	var d db.Drug
	if err := c.ShouldBindJSON(&d); err != nil {
		handleBindErr(c, err)
		return
	}
	d.ID = id
	drug, err := h.Usecase.UpdateDrug(c.Request.Context(), &d)
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
	if err := h.Usecase.DeleteDrug(c.Request.Context(), id); err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ---------------- BATCHES -----------------

func (h *PharmacyHandler) ListBatches(c *gin.Context) {
	drugID, err := strconv.ParseInt(c.Param("drugId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid drugId"})
		return
	}
	batches, err := h.Usecase.ListBatches(c.Request.Context(), drugID)
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
		Batch     db.DrugBatch       `json:"batch"`
		Locations []db.BatchLocation `json:"locations"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		handleBindErr(c, err)
		return
	}
	body.Batch.DrugID = drugID

	detail, err := h.Usecase.CreateBatch(c.Request.Context(), &body.Batch, body.Locations)
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
	drugs, err := h.Usecase.ListDrugs(ctx, nil)
	if err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}

	all := make([]entities.BatchDetail, 0, 128)
	for _, d := range drugs {
		batches, err := h.Usecase.ListBatches(ctx, d.Drug.ID)
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
	detail, err := h.Usecase.GetBatch(c.Request.Context(), batchID)
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
	var body struct {
		Batch     db.DrugBatch       `json:"batch"`
		Locations []db.BatchLocation `json:"locations"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		handleBindErr(c, err)
		return
	}
	body.Batch.ID = batchID

	detail, err := h.Usecase.UpdateBatch(c.Request.Context(), &body.Batch, body.Locations)
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
	if err := h.Usecase.DeleteBatch(c.Request.Context(), batchID); err != nil {
		c.JSON(mapPhErr(err), gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ------------- LOCATIONS ----------------

func (h *PharmacyHandler) ListBatchLocations(c *gin.Context) {
	batchID, err := strconv.ParseInt(c.Param("batchId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid batchId"})
		return
	}
	locs, err := h.Usecase.ListBatchLocations(c.Request.Context(), batchID)
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
	var loc db.BatchLocation
	if err := c.ShouldBindJSON(&loc); err != nil {
		handleBindErr(c, err)
		return
	}
	loc.BatchID = batchID

	created, err := h.Usecase.CreateBatchLocation(c.Request.Context(), &loc)
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
	var loc db.BatchLocation
	if err := c.ShouldBindJSON(&loc); err != nil {
		handleBindErr(c, err)
		return
	}
	loc.ID = locationID

	updated, err := h.Usecase.UpdateBatchLocation(c.Request.Context(), &loc)
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
	if err := h.Usecase.DeleteBatchLocation(c.Request.Context(), locationID); err != nil {
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
	if _, ok := err.(*postgres.DuplicateBatchNumberError); ok {
		return http.StatusConflict
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
