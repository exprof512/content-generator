package main

import (
	"log"

	"github.com/exprof512/content-generator/internal/db"
	"github.com/exprof512/content-generator/internal/logger"
	"github.com/exprof512/content-generator/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Failed to load .env")
	}

	logger.InitLogger()
	db.InitPostgres()
	db.InitRedis()

	router := gin.Default()
	routes.RegisterRoutes(router)

	logger.Log.Info("Server running on :8080")
	if err := router.Run(":8080"); err != nil {
		logger.Log.Fatal("Failed to run server: ", err)
	}
}
