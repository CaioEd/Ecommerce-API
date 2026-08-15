package router

import (
	"net/http"

	"github.com/caioersb/ecommerce-api/internal/handler"
	"github.com/gin-gonic/gin"
)

// New cria o roteador Gin e registra todas as rotas da aplicação.
func New(userHandler *handler.UserHandler) *gin.Engine {
	router := gin.Default()

	// Healthcheck simples.
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Versionamento da API.
	api := router.Group("/api/v1")
	{
		users := api.Group("/users")
		{
			users.POST("", userHandler.Create)
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.GetByID)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}
	}

	return router
}
