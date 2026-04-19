package handler

import (
	"cellar-app/model"
	"cellar-app/service"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListBottles(c *gin.Context) {
	bottles, err := h.Service.ListBottles()
	if err != nil {
		fmt.Println("[ListBottles] error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve bottles"})
		return
	}
	c.JSON(http.StatusOK, bottles)
}

func (h *Handler) GetBottle(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fmt.Println("[GetBottle] invalid id:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	bottle, err := h.Service.GetBottle(id)
	if err != nil {
		fmt.Println("[GetBottle] error:", err)
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
		fmt.Println("[CreateBottle] bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	if err := validateCreateBottleRequest(bottle, false); err != nil {
		fmt.Println("[CreateBottle] validation error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input: " + err.Error()})
		return
	}
	// AddedAtが明示的に指定されていない場合は現在時刻を設定
	if bottle.AddedAt == nil {
		now := time.Now()
		bottle.AddedAt = &now
	}
	if err := h.Service.CreateBottle(&bottle); err != nil {
		fmt.Println("[CreateBottle] service error:", err)
		// 棚位置重複エラー
		if errors.Is(err, service.ErrPositionOccupied) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "POSITION_OCCUPIED",
				"message": "指定された棚位置は既に使用されています",
			})
			return
		}

		// その他エラー
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create bottle",
		})
		return
	}
	c.JSON(http.StatusCreated, bottle)
}

func validateCreateBottleRequest(bottle model.Bottle, ignoreId bool) error {
	if bottle.WineID == 0 && !ignoreId {
		return fmt.Errorf("wine_id is required")
	}
	if bottle.RowNumber == nil || *bottle.RowNumber <= 0 {
		return fmt.Errorf("row_number must be greater than 0")
	}
	if bottle.ColumnNumber == nil || *bottle.ColumnNumber <= 0 {
		return fmt.Errorf("column_number must be greater than 0")
	}
	return nil
}

func (h *Handler) UpdateBottle(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fmt.Println("[UpdateBottle] invalid id:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + err.Error()})
		return
	}
	id := uint(id64)
	var bottle model.Bottle
	if err := c.ShouldBindJSON(&bottle); err != nil {
		fmt.Println("[UpdateBottle] bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}
	bottle.ID = id
	if err := h.Service.UpdateBottle(&bottle); err != nil {
		fmt.Println("[UpdateBottle] service error:", err)
		// 棚位置重複エラー
		if errors.Is(err, service.ErrPositionOccupied) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "POSITION_OCCUPIED",
				"message": "指定された棚位置は既に使用されています",
			})
			return
		}

		// その他エラー
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create bottle",
		})
		return
	}
	c.JSON(http.StatusOK, bottle)
}

func (h *Handler) PatchBottle(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fmt.Println("[PatchBottle] invalid id:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id: " + err.Error()})
		return
	}
	id := uint(id64)
	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid input, %s", err.Error())})
		return
	}
	// AddedAtは更新不可のため、もしリクエストに含まれていたら削除
	delete(updates, "added_at")

	if err := h.Service.PatchBottle(id, updates); err != nil {
		fmt.Println("[PatchBottle] service error:", err)
		// 棚位置重複エラー
		if errors.Is(err, service.ErrPositionOccupied) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "POSITION_OCCUPIED",
				"message": "指定された棚位置は既に使用されています",
			})
			return
		}

		// その他エラー
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create bottle",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bottle patched successfully"})
}

func (h *Handler) DeleteBottle(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fmt.Println("[DeleteBottle] invalid id:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)
	if err := h.Service.DeleteBottle(id); err != nil {
		fmt.Println("[DeleteBottle] service error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete bottle"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "bottle deleted successfully"})
}
