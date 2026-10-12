package handler

import (
	"cellar-app/dto"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) ListAppellations(c *gin.Context) {
	appellations, err := h.Service.ListAppellations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve appellations"})
		return
	}
	response := make([]dto.AppellationDTO, 0, len(appellations))
	for _, appellation := range appellations {
		response = append(response, dto.AppellationFromModel(appellation))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetAppellation(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	appellation, err := h.Service.GetAppellation(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "appellation not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve appellation"})
		}
		return
	}
	c.JSON(http.StatusOK, dto.AppellationFromModel(*appellation))
}

func (h *Handler) CreateAppellation(c *gin.Context) {
	var request dto.AppellationDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	appellation := request.ToModel()
	if err := h.Service.CreateAppellation(&appellation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create appellation"})
		return
	}
	c.JSON(http.StatusCreated, dto.AppellationFromModel(appellation))
}

func (h *Handler) UpdateAppellation(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + err.Error()})
		return
	}
	id := uint(id64)
	var request dto.AppellationDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}
	appellation := request.ToModel()
	appellation.ID = id
	if err := h.Service.UpdateAppellation(&appellation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update appellation: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, dto.AppellationFromModel(appellation))
}

func (h *Handler) DeleteAppellation(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	if err := h.Service.DeleteAppellation(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete appellation"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "appellation deleted successfully"})
}
