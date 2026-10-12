package handler

import (
	"cellar-app/dto"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListWineTypes(c *gin.Context) {
	wineTypes, err := h.Service.ListWineTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list wine types"})
		return
	}
	response := make([]dto.WineTypeDTO, 0, len(wineTypes))
	for _, wineType := range wineTypes {
		response = append(response, dto.WineTypeFromModel(wineType))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetWineType(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	wineType, err := h.Service.GetWineType(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wine type not found"})
		return
	}
	c.JSON(http.StatusOK, dto.WineTypeFromModel(*wineType))
}

func (h *Handler) CreateWineType(c *gin.Context) {
	var req dto.WineTypeDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	wineTypeModel := req.ToModel()
	err := h.Service.CreateWineType(&wineTypeModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wine type"})
		return
	}
	c.JSON(http.StatusCreated, dto.WineTypeFromModel(wineTypeModel))
}

func (h *Handler) UpdateWineType(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	var req dto.WineTypeDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	req.ID = id // Ensure the ID is set for the update
	wineTypeModel := req.ToModel()
	err = h.Service.UpdateWineType(&wineTypeModel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update wine type"})
		return
	}
	c.JSON(http.StatusOK, dto.WineTypeFromModel(wineTypeModel))
}

func (h *Handler) DeleteWineType(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	if err := h.Service.DeleteWineType(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete wine type"})
		return
	}
	c.Status(http.StatusNoContent)
}
