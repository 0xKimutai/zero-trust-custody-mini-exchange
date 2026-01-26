package api

import (
	"mini-exchange/internal/auth"
	"mini-exchange/internal/deposit"
	"mini-exchange/internal/por"
	"mini-exchange/internal/withdrawal"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all the routes for the application
func SetupRouter(
	authService *auth.Service,
	depositHandler *deposit.Handler,
	withdrawalHandler *withdrawal.Handler,
	porHandler *por.Handler,
) *gin.Engine {
	r := gin.Default()

	// Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up",
			"system": "mini-custody-backend",
		})
	})

	authHandler := auth.NewHandler(authService)

	// Public Routes
	authGroup := r.Group("/api/v1/auth") // Authenticated routes are public but specific
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
	}

	// Admin / Simulation Routes
	admin := r.Group("/admin")
	{
		admin.POST("/deposit/webhook", depositHandler.SimulateWebhook)
		admin.POST("/withdrawal/process", withdrawalHandler.ProcessBatch)
		admin.POST("/por/generate", porHandler.Generate)
	}

	// Protected Routes
	// Apply Rate Limiting globally or per group. Applying here for API
	apiV1 := r.Group("/api/v1")
	apiV1.Use(RateLimitMiddleware())
	apiV1.Use(auth.AuthMiddleware(authService))
	{
		apiV1.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})
		apiV1.GET("/deposit/address", depositHandler.GetAddress)
		apiV1.POST("/withdraw", withdrawalHandler.RequestWithdrawal)
		apiV1.GET("/withdrawals", withdrawalHandler.GetHistory)
	}

	return r
}
