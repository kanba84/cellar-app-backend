package handler

import (
	"cellar-app/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListDesignationTypes(c *gin.Context) {
	designationTypes, err := h.Service.ListDesignationTypes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve designation types"})
		return
	}
	c.JSON(http.StatusOK, designationTypes)
}

func (h *Handler) GetDesignationType(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	designationType, err := h.Service.GetDesignationType(id)
	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "designation type not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve designation type"})
		}
		return
	}
	c.JSON(http.StatusOK, designationType)
}

func (h *Handler) CreateDesignationType(c *gin.Context) {
	var designationType model.DesignationType
	if err := c.ShouldBindJSON(&designationType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if err := h.Service.CreateDesignationType(&designationType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create designation type"})
		return
	}
	c.JSON(http.StatusCreated, designationType)
}

func (h *Handler) UpdateDesignationType(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + err.Error()})
		return
	}
	id := uint(id64)
	var designationType model.DesignationType
	if err := c.ShouldBindJSON(&designationType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}
	designationType.ID = id
	if err := h.Service.UpdateDesignationType(&designationType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update designation type: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, designationType)
}

func (h *Handler) DeleteDesignationType(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	if err := h.Service.DeleteDesignationType(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete designation type"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "designation type deleted successfully"})
}
