package handler

import (
	"cellar-app/model"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListBottles(c *gin.Context) {
	bottles, err := h.Service.ListBottles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve bottles"})
		return
	}
	c.JSON(http.StatusOK, bottles)
}

func (h *Handler) GetBottle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	bottle, err := h.Service.GetBottle(id)
	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "bottle not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve bottle"})
		}
		return
	}
	c.JSON(http.StatusOK, bottle)
}

func (h *Handler) CreateBottle(c *gin.Context) {
	var bottle model.Bottle
	if err := c.ShouldBindJSON(&bottle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if err := h.Service.CreateBottle(&bottle); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create bottle"})
		return
	}
	c.JSON(http.StatusCreated, bottle)
}

func (h *Handler) UpdateBottle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + err.Error()})
		return
	}
	var bottle model.Bottle
	if err := c.ShouldBindJSON(&bottle); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}
	bottle.ID = id
	if err := h.Service.UpdateBottle(&bottle); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update bottle: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, bottle)
}

func (h *Handler) PatchBottle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + err.Error()})
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}

	if err := h.Service.PatchBottle(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to patch bottle: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bottle patched successfully"})
}

func (h *Handler) DeleteBottle(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Service.DeleteBottle(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete bottle"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bottle deleted successfully"})
}
