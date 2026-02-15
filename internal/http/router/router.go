package router

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"green-api/internal/config"
	"green-api/internal/docs"
	"green-api/internal/http/handler"
	"green-api/internal/middleware"
	"green-api/internal/service"
)

func New(cfg config.Config, logger *zap.Logger, service *service.Service) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestID())
	engine.Use(middleware.RequestLogger(logger))

	engine.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-Request-Id"},
		ExposeHeaders:    []string{"X-Request-Id"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	engine.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(200, "application/yaml; charset=utf-8", docs.OpenAPIYAML)
	})
	engine.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/docs/index.html")
	})
	engine.GET("/docs/index.html", func(c *gin.Context) {
		c.Data(200, "text/html; charset=utf-8", []byte(docs.SwaggerHTML("/openapi.yaml")))
	})

	api := engine.Group("/api/v1")
	h := handler.NewGreenAPIHandler(service)
	h.RegisterRoutes(api)

	return engine
}
