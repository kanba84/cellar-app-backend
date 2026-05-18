package handler

import (
	"cellar-app/model"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetWineLLMInfo: Wine ID から LLM を使用してワイン情報を取得します
// GET /wines/:id/llm-info
// 取得した情報はフロントエンドで表示され、ユーザー操作により別途保存されます
func (h *Handler) GetWineLLMInfo(c *gin.Context) {
	// Wine ID を取得
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		log.Printf("GetWineLLMInfo invalid id error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	id := uint(id64)

	// Wine を取得
	wineDTO, err := h.Service.GetWine(id)
	if err != nil {
		if err.Error() == "not found" {
			log.Printf("GetWineLLMInfo wine not found: id=%d", id)
			c.JSON(http.StatusNotFound, gin.H{"error": "wine not found"})
		} else {
			log.Printf("GetWineLLMInfo failed to get wine: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve wine"})
		}
		return
	}

	// WineDTO を Wine に変換（LLM 処理用）
	wine := &model.Wine{
		ID:      uint(wineDTO.ID),
		Name:    wineDTO.Name,
		Vintage: wineDTO.Vintage,
	}
	if wineDTO.Producer != nil {
		wine.Producer = wineDTO.Producer
	}

	// LLM からワイン情報を取得（DB保存なし）
	ctx := c.Request.Context()
	info, err := h.Service.FetchWineInfoOnly(ctx, wine)
	if err != nil {
		log.Printf("GetWineLLMInfo failed to fetch wine info: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch wine info"})
		return
	}

	c.JSON(http.StatusOK, info)
}
