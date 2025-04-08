package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/config"
	"github.com/skakunma/go-musthave-diploma-tpl/internal/middleware"
)

func SetupRouter(server *gin.Engine, cfg *config.Config) {
	api := server.Group("/api/user")
	auth := api.Group("", middleware.AuthMiddleware(cfg))
	server.POST("/api/user/register", func(c *gin.Context) { RegisterHandler(c, cfg) })
	server.POST("/api/user/login", func(c *gin.Context) { Login(c, cfg) })
	auth.POST("/api/user/orders", func(c *gin.Context) { CreateOrder(c, cfg) })
	auth.GET("/api/user/orders", func(c *gin.Context) { GetOrders(c, cfg) })
	auth.GET("/api/user/balance", func(c *gin.Context) { GetBalance(c, cfg) })
	auth.POST("/api/user/balance/withdraw", func(c *gin.Context) { WithdrawBalance(c, cfg) })
	auth.GET("/api/user/withdrawals", func(c *gin.Context) { GetWithdrawals(c, cfg) })
}
