package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetStats: すべての統計情報を取得
// クエリパラメータ: days (在庫推移の期間、デフォルト30日)
func (h *Handler) GetStats(c *gin.Context) {
	days := 30 // デフォルト値
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 {
			days = d
		}
	}

	stats, err := h.Service.GetStats(days)
	if err != nil {
		fmt.Println("[GetStats] error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetWineTypeStats: ワインタイプ別の在庫構成比を取得
func (h *Handler) GetWineTypeStats(c *gin.Context) {
	wineTypeStats, err := h.Service.GetWineTypeStats()
	if err != nil {
		fmt.Println("[GetWineTypeStats] error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve wine type stats"})
		return
	}

	c.JSON(http.StatusOK, wineTypeStats)
}

// GetCountryStats: 生産国別の在庫構成比を取得
func (h *Handler) GetCountryStats(c *gin.Context) {
	countryStats, err := h.Service.GetCountryStats()
	if err != nil {
		fmt.Println("[GetCountryStats] error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve country stats"})
		return
	}

	c.JSON(http.StatusOK, countryStats)
}

// GetInventoryTrend: 在庫数推移を取得
// クエリパラメータ: days (期間、デフォルト30日)
func (h *Handler) GetInventoryTrend(c *gin.Context) {
	days := 30 // デフォルト値
	if daysParam := c.Query("days"); daysParam != "" {
		if d, err := strconv.Atoi(daysParam); err == nil && d > 0 {
			days = d
		}
	}

	inventoryTrend, err := h.Service.GetInventoryTrend(days)
	if err != nil {
		fmt.Println("[GetInventoryTrend] error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve inventory trend"})
		return
	}

	c.JSON(http.StatusOK, inventoryTrend)
}
