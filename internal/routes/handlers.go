package routes

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/exprof512/content-generator/internal/db"
	"github.com/exprof512/content-generator/internal/logger"
	"github.com/exprof512/content-generator/internal/models"
	"github.com/exprof512/content-generator/pkg/deepseek"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	client := deepseek.NewClient(os.Getenv("DEEPSEEK_API_KEY"))

	r.POST("/generate", func(c *gin.Context) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Log.Warn("Invalid prompt")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid prompt"})
			return
		}

		val, err := db.Redis.Get(c, req.Prompt).Result()
		if err == nil {
			c.JSON(http.StatusOK, gin.H{"content": val})
			return
		}

		if os.Getenv("MOCK_MODE") == "true" {
			logger.Log.WithField("prompt", req.Prompt).Info("MOCK MODE: Возвращается мок-контент")

			mockContent := fmt.Sprintf(`[Моковые данные] Пример контента для: "%s". 
				Lorem ipsum dolor sit amet, consectetur adipiscing elit. 
				Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.`, req.Prompt)

			db.Redis.Set(c, req.Prompt, mockContent, 60*time.Second)

			c.JSON(http.StatusOK, gin.H{"content": mockContent})
			return
		}

		content, err := client.Generate(req.Prompt)
		if err != nil {
			logger.Log.Error("DeepSeek error: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации"})
			return
		}

		db.Redis.Set(c, req.Prompt, content, 60*time.Second)
		c.JSON(http.StatusOK, gin.H{"content": content})
	})

	r.POST("/feedback", func(c *gin.Context) {
		var f models.Feedback
		if err := c.ShouldBindJSON(&f); err != nil {
			logger.Log.Warn("Invalid feedback")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid feedback"})
			return
		}
		_, err := db.Postgres.Exec("INSERT INTO feedback (score, content) VALUES ($1, $2)", f.Score, f.Content)
		if err != nil {
			logger.Log.Error("DB insert failed: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при сохранении"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})
}
