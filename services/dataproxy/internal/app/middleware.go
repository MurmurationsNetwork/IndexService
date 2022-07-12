package app

import (
	"time"

	"github.com/MurmurationsNetwork/MurmurationsServices/common/middleware/limiter"
	"github.com/MurmurationsNetwork/MurmurationsServices/common/middleware/logger"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/dataproxy/config"
	corslib "github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		gin.Recovery(),
		limiter.NewRateLimitWithOptions(limiter.RateLimitOptions{
			Period: config.Conf.Server.PostRateLimitPeriod,
			Method: "POST",
		}),
		limiter.NewRateLimitWithOptions(limiter.RateLimitOptions{
			Period: config.Conf.Server.GetRateLimitPeriod,
			Method: "GET",
		}),
		logger.NewLogger(),
		cors(),
	}
}

func cors() gin.HandlerFunc {
	// CORS for all origins, allowing:
	// - GET and POST methods
	// - Origin, Authorization and Content-Type header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	return corslib.New(corslib.Config{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}