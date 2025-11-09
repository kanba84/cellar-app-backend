package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UploadLabelImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		fmt.Println("[UploadLabelImage] form file error:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to get image from form data"})
		return
	}
	// 即時レスポンス用のURLを生成
	outputFileName := fmt.Sprintf("thumb_%s", file.Filename)
	url := fmt.Sprintf("https://cellar-app.local/labels/%s", outputFileName)

	// 先にレスポンスを返す
	c.JSON(http.StatusAccepted, gin.H{
		"thumb_url": url,
		"message":   "thumbnail processing started",
	})

	// 非同期で処理を実行
	go func() {
		if err := h.Service.UploadLabelImage(file, outputFileName); err != nil {
			fmt.Println("[UploadLabelImage] error:", err)
		}
	}()
}
