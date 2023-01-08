package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	ginSwagger "github.com/swaggo/gin-swagger"
	// docs is generated by Swag CLI, you have to import it.
	swaggerFiles "github.com/swaggo/files"
	_ "github.com/zsfarkas/chartinstaller/docs"

	"github.com/zsfarkas/chartinstaller/src/health"
	"github.com/zsfarkas/chartinstaller/src/releases"
)

// @title           Chart Installer API
// @version         1.0
// @description     This API allows to install helm charts from the configured chart museum.

// @BasePath  /api/v1

func main() {
	log.SetPrefix("[chartinstaller] ")
	log.Println("booting...")

	r := gin.Default()
	r.SetTrustedProxies(nil)

	v1 := r.Group("/api/v1")
	{
		releasesGroup := v1.Group("/releases")
		{
			c := releases.NewController()
			releasesGroup.GET("/config", c.GetConfig)
			releasesGroup.GET("", c.ListReleases)
			releasesGroup.GET(":name", c.StatusRelease)
			releasesGroup.PUT(":name", c.InstallOrUpgradeRelease)
			releasesGroup.DELETE(":name", c.UninstallRelease)
		}
		healthGroup := v1.Group("/health")
		{
			healthController := health.NewController()
			healthGroup.GET("", healthController.Health)
		}

	}
	log.Println("api initialized...")

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	log.Println("swagger initialized...")

	port, present := os.LookupEnv("PORT")
	if !present {
		port = "8080"
	}

	log.Printf("listening on port %s...", port)

	r.Run(":" + port)
}