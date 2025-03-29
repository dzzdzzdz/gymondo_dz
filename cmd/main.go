package main

import (
	"gymondo_dz/pkg/handlers"
	"gymondo_dz/pkg/repositories"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	productRepo := repositories.NewProductRepository()
	subscriptionRepo := repositories.NewSubscriptionRepository()

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

	port := ":8080"
	log.Printf("Starting server on %s", port)
	if err := router.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
