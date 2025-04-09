package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/exprof512/content-generator/pkg/deepseek"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/patrickmn/go-cache" // Импортируем пакет go-cache
)

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Инициализация клиента DeepSeek
	client := deepseek.NewClient(os.Getenv("DEEPSEEK_API_KEY"))

	// Инициализация кэша
	cacheInstance := cache.New(5*time.Minute, 10*time.Minute) // Кэш живет 5 минут, очистка каждые 10 минут

	// Настройка роутера Gin
	router := gin.Default()

	// Обработка фидбека
	router.POST("/feedback", func(c *gin.Context) {
		var feedback struct {
			Score   int    `json:"score"`
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&feedback); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		// Сохраняем в базу данных
		c.JSON(200, gin.H{"status": "success"})
	})

	// Маршрут для генерации контента
	router.POST("/generate", func(c *gin.Context) {
		var request struct {
			Prompt string `json:"prompt" binding:"required"`
		}

		// Валидация запроса
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		// Логирование запроса
		log.Printf("Request received: %s", request.Prompt)

		// Проверка кэша
		if cachedResult, found := cacheInstance.Get(request.Prompt); found {
			log.Printf("Cache hit for prompt: %s", request.Prompt)
			c.JSON(http.StatusOK, gin.H{"content": cachedResult})
			return
		}

		// Генерация контента через API
		result, err := client.Generate(request.Prompt)
		if err != nil {
			// Логируем подробную ошибку для себя
			log.Printf("API generation error: %v", err)
			// Отправляем пользователю общее сообщение об ошибке
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при генерации контента. Попробуйте позже."})
			return
		}

		// Сохранение результата в кэш (используем глобальную константу DefaultExpiration)
		cacheInstance.Set(request.Prompt, result, cache.DefaultExpiration)
		log.Printf("Generated and cached result for prompt: %s", request.Prompt)

		// Отправка результата клиенту
		c.JSON(http.StatusOK, gin.H{"content": result})
	})

	// Запуск сервера
	fmt.Println("Server running on :8080")
	router.Run(":8080")
}
