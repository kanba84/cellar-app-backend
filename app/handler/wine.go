package handler

import (
	"fmt"
	"log" // 追加
	"net/http"
	"strconv"

	"cellar-app/model"

	"github.com/gin-gonic/gin"
)

// 例: ListWines ハンドラの修正例
func (h *Handler) ListWines(c *gin.Context) {
	wines, err := h.Service.ListWines()
	if err != nil {
		log.Printf("ListWines error: %v", err) // 追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve wines"})
		return
	}
	c.JSON(http.StatusOK, wines)
}

func (h *Handler) GetWine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("GetWine invalid id error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	wine, err := h.Service.GetWine(id)
	if err != nil {
		if err.Error() == "not found" {
			log.Printf("GetWine not found error: %v", err) // 追加（任意）
			c.JSON(http.StatusNotFound, gin.H{"error": "wine not found"})
		} else {
			log.Printf("GetWine service error: %v", err) // 追加
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve wine"})
		}
		return
	}
	c.JSON(http.StatusOK, wine)
}

func (h *Handler) CreateWine(c *gin.Context) {
	var wine model.Wine
	if err := c.ShouldBindJSON(&wine); err != nil {
		log.Printf("CreateWine bind error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if err := validateCreateWineRequest(wine); err != nil {
		log.Printf("CreateWine validation error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if err := h.Service.CreateWine(&wine); err != nil {
		log.Printf("CreateWine service error: %v", err) // 追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wine"})
		return
	}
	c.JSON(http.StatusCreated, wine)
}

func validateCreateWineRequest(wine model.Wine) error {
	if wine.Name == "" {
		return fmt.Errorf("name is required")
	}
	if wine.WineTypeID == 0 {
		return fmt.Errorf("wine_type_id is required")
	}
	if wine.CountryID == 0 {
		return fmt.Errorf("country_id is required")
	}
	return nil
}

func (h *Handler) CreateWineWithBottle(c *gin.Context) {
	var req model.CreateWineWithBottleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("CreateWineWithBottle bind error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}
	if err := validateCreateWineWithBottleRequest(req); err != nil {
		log.Printf("CreateWineWithBottle validation error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}
	wine, bottle, err := h.Service.CreateWineWithBottle(c, req)
	if err != nil {
		log.Printf("CreateWineWithBottle service error: %v", err) // 追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wine and bottle"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wine": wine, "bottle": bottle})
}

func validateCreateWineWithBottleRequest(req model.CreateWineWithBottleRequest) error {
	if validateCreateWineRequest(req.Wine) != nil {
		return fmt.Errorf("invalid wine data")
	}
	if validateCreateBottleRequest(req.Bottle, true) != nil {
		return fmt.Errorf("invalid bottle data")
	}
	return nil
}

func (h *Handler) DeleteWine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("DeleteWine invalid id error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Service.DeleteWine(id); err != nil {
		log.Printf("DeleteWine service error: %v", err) // 追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete wine"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wine deleted successfully"})
}

func (h *Handler) UpdateWine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("UpdateWine invalid id error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id" + err.Error()})
		return
	}
	var wine model.Wine
	if err := c.ShouldBindJSON(&wine); err != nil {
		log.Printf("UpdateWine bind error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input" + err.Error()})
		return
	}
	wine.ID = id
	if err := h.Service.UpdateWine(&wine); err != nil {
		log.Printf("UpdateWine service error: %v", err) // 追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update wine" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, wine)
}

func (h *Handler) PatchWine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("PatchWine invalid id error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		log.Printf("PatchWine bind error: %v", err) // 追加
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}
	if err := h.Service.PatchWine(id, updates); err != nil {
		log.Printf("PatchWine service error: %v", err) // 追加
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to patch wine"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wine patched successfully"})
}
