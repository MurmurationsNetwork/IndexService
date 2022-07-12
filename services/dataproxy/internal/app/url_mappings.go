package app

import (
	"github.com/MurmurationsNetwork/MurmurationsServices/services/dataproxy/internal/controller/http"
	"github.com/MurmurationsNetwork/MurmurationsServices/services/dataproxy/internal/repository/db"
	"github.com/gin-gonic/gin"
)

func mapUrls(router *gin.Engine) {
	pingHandler := http.NewPingHandler()
	mappingsHandler := http.NewMappingsHandler(db.NewMappingRepository())
	profilesHandler := http.NewProfilesHandler(db.NewProfileRepository())

	v1 := router.Group("/v1")
	{
		v1.GET("/ping", pingHandler.Ping)
		v1.POST("/mappings", mappingsHandler.Create)
		v1.GET("/profiles/:profileId", profilesHandler.Get)
	}
}