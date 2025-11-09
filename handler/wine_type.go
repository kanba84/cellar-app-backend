package handler

import (
	"cellar-app/model"
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
	c.JSON(http.StatusOK, wineTypes)
}

func (h *Handler) GetWineType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	wineType, err := h.Service.GetWineType(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "wine type not found"})
		return
	}
	c.JSON(http.StatusOK, wineType)
}

func (h *Handler) CreateWineType(c *gin.Context) {
	var req model.WineType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	err := h.Service.CreateWineType(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wine type"})
		return
	}
	c.JSON(http.StatusCreated, req)
}

func (h *Handler) UpdateWineType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req model.WineType
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	req.ID = id // Ensure the ID is set for the update
	err = h.Service.UpdateWineType(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update wine type"})
		return
	}
	c.JSON(http.StatusOK, req)
}

func (h *Handler) DeleteWineType(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Service.DeleteWineType(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete wine type"})
		return
	}
	c.Status(http.StatusNoContent)
}
