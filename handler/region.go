package handler

import (
	"cellar-app/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListRegions handles GET /regions
func (h *Handler) ListRegions(c *gin.Context) {
	regions, err := h.Service.ListRegions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	response := make([]dto.RegionDTO, 0, len(regions))
	for _, region := range regions {
		response = append(response, dto.RegionFromModel(region))
	}
	c.JSON(http.StatusOK, response)
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
	c.JSON(http.StatusOK, dto.RegionFromModel(*region))
}

// CreateRegion handles POST /regions
func (h *Handler) CreateRegion(c *gin.Context) {
	var req dto.RegionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	regionModel := req.ToModel()
	err := h.Service.CreateRegion(&regionModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, dto.RegionFromModel(regionModel))
}

// UpdateRegion handles PUT /regions/:id
func (h *Handler) UpdateRegion(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	var req dto.RegionDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = id // Ensure the ID is set for the update
	regionModel := req.ToModel()
	err = h.Service.UpdateRegion(&regionModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.RegionFromModel(regionModel))
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
