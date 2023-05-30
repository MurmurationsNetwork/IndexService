package app

import (
	"github.com/MurmurationsNetwork/MurmurationsServices/pkg/middleware/logger"
	"github.com/gin-gonic/gin"
)

func getMiddlewares() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		gin.Recovery(),
		logger.NewLogger(),
	}
}
