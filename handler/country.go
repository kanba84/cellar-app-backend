package handler

import (
	"cellar-app/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListCountries handles GET /countries
func (h *Handler) ListCountries(c *gin.Context) {
	countries, err := h.Service.ListCountries()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list countries"})
		return
	}
	c.JSON(http.StatusOK, countries)
}

// GetCountry handles GET /countries/:id
func (h *Handler) GetCountry(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country ID"})
		return
	}
	id := uint(id64)
	country, err := h.Service.GetCountry(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Country not found"})
		return
	}
	c.JSON(http.StatusOK, country)
}

// CreateCountry handles POST /countries
func (h *Handler) CreateCountry(c *gin.Context) {
	var country model.Country
	if err := c.ShouldBindJSON(&country); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.Service.CreateCountry(&country); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create country"})
		return
	}
	c.JSON(http.StatusCreated, country)
}

// UpdateCountry handles PUT /countries/:id
func (h *Handler) UpdateCountry(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country ID"})
		return
	}
	id := uint(id64)
	var country model.Country
	if err := c.ShouldBindJSON(&country); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	country.ID = id
	if err := h.Service.UpdateCountry(&country); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update country"})
		return
	}
	c.JSON(http.StatusOK, country)
}

// DeleteCountry handles DELETE /countries/:id
func (h *Handler) DeleteCountry(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid country ID"})
		return
	}
	id := uint(id64)
	if err := h.Service.DeleteCountry(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete country"})
		return
	}
	c.Status(http.StatusNoContent)
}
