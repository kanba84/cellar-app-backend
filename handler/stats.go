package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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

// GetVintageStats: ワインのビンテージ別の在庫構成比を取得
func (h *Handler) GetVintageStats(c *gin.Context) {
	vintageStats, err := h.Service.GetVintageStats()
	if err != nil {
		fmt.Println("[GetVintageStats] error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve vintage stats"})
		return
	}

	c.JSON(http.StatusOK, vintageStats)
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

// CreateDailySnapshot: 指定日付の日次スナップショットを生成（日次バッチ用）
// リクエストボディ: {\"date\": \"2026-04-26\"} （オプション、指定なしの場合は本日）
func (h *Handler) CreateDailySnapshot(c *gin.Context) {
	var req struct {
		Date *string `json:"date"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Println("[CreateDailySnapshot] bind error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	// スナップショット対象日を決定
	snapshotDate := time.Now()
	if req.Date != nil && *req.Date != "" {
		// リクエストで指定された日付を使用
		parsed, err := time.Parse("2006-01-02", *req.Date)
		if err != nil {
			fmt.Println("[CreateDailySnapshot] date parse error:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format (use YYYY-MM-DD)"})
			return
		}
		snapshotDate = parsed
	}

	if err := h.Service.CreateOrUpdateDailySnapshot(snapshotDate); err != nil {
		fmt.Println("[CreateDailySnapshot] service error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create snapshot"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "snapshot created successfully", "date": snapshotDate.Format("2006-01-02")})
}

// GetDailySnapshot: 指定日付の日次スナップショットを取得
// パラメータ: date (YYYY-MM-DD形式)
func (h *Handler) GetDailySnapshot(c *gin.Context) {
	dateStr := c.Param("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date parameter is required"})
		return
	}

	snapshotDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		fmt.Println("[GetDailySnapshot] date parse error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format (use YYYY-MM-DD)"})
		return
	}

	snapshot, err := h.Service.GetDailySnapshot(snapshotDate)
	if err != nil {
		fmt.Println("[GetDailySnapshot] service error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve snapshot"})
		return
	}

	if snapshot == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
		return
	}

	c.JSON(http.StatusOK, snapshot)
}
