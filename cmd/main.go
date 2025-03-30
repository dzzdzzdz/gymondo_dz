package main

import (
	"gymondo_dz/pkg/database"
	"gymondo_dz/pkg/handlers"
	"gymondo_dz/pkg/repositories"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.NewPostgresConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := database.AutoMigrate(db, false); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	productRepo := repositories.NewProductRepository(db)
	subscriptionRepo := repositories.NewSubscriptionRepository(db)

	productHandler := handlers.NewProductHandler(productRepo)
	subscriptionHandler := handlers.NewSubscriptionHandler(subscriptionRepo, productRepo)

	router := gin.Default()

	productRoutes := router.Group("/products")
	{
		productRoutes.GET("", productHandler.GetProducts)
		productRoutes.GET("/:id", productHandler.GetProduct)
	}

	subscriptionRoutes := router.Group("/subscriptions")
	{
		subscriptionRoutes.POST("/:product_id", subscriptionHandler.CreateSubscription)
		subscriptionRoutes.GET("/:id", subscriptionHandler.GetSubscription)
		subscriptionRoutes.PATCH("/:id/pause", subscriptionHandler.PauseSubscription)
		subscriptionRoutes.PATCH("/:id/unpause", subscriptionHandler.UnpauseSubscription)
		subscriptionRoutes.DELETE("/:id", subscriptionHandler.CancelSubscription)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
