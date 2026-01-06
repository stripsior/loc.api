package api

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stripsior/loc.api/internal/cache"
)

func SetupRouter(cacheTTL time.Duration) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	corsConfig := cors.Config{
		AllowAllOrigins:  false,
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"X-Cache"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	if len(corsConfig.AllowOrigins) == 0 || (len(corsConfig.AllowOrigins) == 1 && corsConfig.AllowOrigins[0] == "*") {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowOrigins = nil
	}

	router.Use(cors.New(corsConfig))

	resultCache := cache.NewCache(cacheTTL)

	router.Use(func(c *gin.Context) {
		c.Set("cache", resultCache)
		c.Next()
	})

	router.GET("/health", healthCheck)

	router.GET("/cache/stats", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"size": resultCache.Size(),
			"ttl":  cacheTTL.String(),
		})
	})

	api := router.Group("/api")
	{
		api.POST("/analyze", analyzeRepository)
	}

	return router
}

func getAllowedOrigins() []string {
	allowOrigin := os.Getenv("CORS_ALLOW_ORIGIN")
	if allowOrigin == "" {
		return []string{"*"}
	}

	origins := strings.Split(allowOrigin, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return origins
}

func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "ok",
		"service": "loc-counter-api",
	})
}
