package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"cellar-app/model"

	"github.com/gin-gonic/gin"
)

// wine handlers
func (h *Handler) ListWines(c *gin.Context) {
	wines, err := h.Service.ListWines()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve wines"})
		return
	}
	c.JSON(http.StatusOK, wines)
}

func (h *Handler) GetWine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	wine, err := h.Service.GetWine(id)
	if err != nil {
		if err.Error() == "not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "wine not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve wine"})
		}
		return
	}
	c.JSON(http.StatusOK, wine)
}

func (h *Handler) CreateWine(c *gin.Context) {
	var wine model.Wine
	if err := c.ShouldBindJSON(&wine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if err := h.Service.CreateWine(&wine); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create wine"})
		return
	}
	c.JSON(http.StatusCreated, wine)
}

func (h *Handler) CreateWineWithBottle(c *gin.Context) {
	var req model.CreateWineWithBottleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	wine, bottle, err := h.Service.CreateWineWithBottle(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wine and bottle"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"wine": wine, "bottle": bottle})
}

func (h *Handler) DeleteWine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Service.DeleteWine(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete wine"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wine deleted successfully"})
}

func (h *Handler) UpdateWine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id" + err.Error()})
		return
	}
	var wine model.Wine
	if err := c.ShouldBindJSON(&wine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input" + err.Error()})
		return
	}
	wine.ID = id
	if err := h.Service.UpdateWine(&wine); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update wine" + err.Error()})
		return
	}
	c.JSON(http.StatusOK, wine)
}

func (h *Handler) PatchWine(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	// 部分更新用のデータを受け取る
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}

	// PatchWine 関数を呼び出して部分更新を実行
	if err := h.Service.PatchWine(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to patch wine"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "wine patched successfully"})
}
