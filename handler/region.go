package handler

import (
	"net/http"
	"strconv"

	"cellar-app/model"

	"github.com/gin-gonic/gin"
)

// ListRegions handles GET /regions
func (h *Handler) ListRegions(c *gin.Context) {
	regions, err := h.Service.ListRegions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, regions)
}

// GetRegion handles GET /regions/:id
func (h *Handler) GetRegion(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	region, err := h.Service.GetRegion(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, region)
}

// CreateRegion handles POST /regions
func (h *Handler) CreateRegion(c *gin.Context) {
	var req model.Region
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := h.Service.CreateRegion(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, req)
}

// UpdateRegion handles PUT /regions/:id
func (h *Handler) UpdateRegion(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	var req model.Region
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id // Ensure the ID is set for the update
	err = h.Service.UpdateRegion(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, req)
}

// DeleteRegion handles DELETE /regions/:id
func (h *Handler) DeleteRegion(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	if err := h.Service.DeleteRegion(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
