package handler

import (
	"cellar-app/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListAppellations(c *gin.Context) {
	appellations, err := h.Service.ListAppellations()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve appellations"})
		return
	}
	c.JSON(http.StatusOK, appellations)
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
		if err.Error() == "not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "appellation not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve appellation"})
		}
		return
	}
	c.JSON(http.StatusOK, appellation)
}

func (h *Handler) CreateAppellation(c *gin.Context) {
	var appellation model.Appellation
	if err := c.ShouldBindJSON(&appellation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if err := h.Service.CreateAppellation(&appellation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create appellation"})
		return
	}
	c.JSON(http.StatusCreated, appellation)
}

func (h *Handler) UpdateAppellation(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + err.Error()})
		return
	}
	id := uint(id64)
	var appellation model.Appellation
	if err := c.ShouldBindJSON(&appellation); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}
	appellation.ID = id
	if err := h.Service.UpdateAppellation(&appellation); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update appellation: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, appellation)
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
