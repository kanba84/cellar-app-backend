package main

import (
	"cellar-app/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(h *handler.Handler) *gin.Engine {
	r := gin.Default()
	// ルーティング設定
	r.GET("/wines", h.ListWines)
	r.GET("/wines/:id", h.GetWine)
	r.POST("/wines", h.CreateWine)
	r.POST("/wines/with_bottle", h.CreateWineWithBottle)
	r.PUT("/wines/:id", h.UpdateWine)
	r.PATCH("/wines/:id", h.PatchWine)
	r.DELETE("/wines/:id", h.DeleteWine)

	r.GET("/bottles", h.ListBottles)
	r.GET("/bottles/:id", h.GetBottle)
	r.POST("/bottles", h.CreateBottle)
	r.PUT("/bottles/:id", h.UpdateBottle)
	r.PATCH("/bottles/:id", h.PatchBottle)
	r.DELETE("/bottles/:id", h.DeleteBottle)

	r.GET("/countries", h.ListCountries)
	r.GET("/countries/:id", h.GetCountry)
	r.POST("/countries", h.CreateCountry)
	r.PUT("/countries/:id", h.UpdateCountry)
	r.DELETE("/countries/:id", h.DeleteCountry)

	r.GET("/regions", h.ListRegions)
	r.GET("/regions/:id", h.GetRegion)
	r.POST("/regions", h.CreateRegion)
	r.PUT("/regions/:id", h.UpdateRegion)
	//r.PATCH("/regions/:id", h.PatchRegion)
	r.DELETE("/regions/:id", h.DeleteRegion)

	r.GET("/wine_types", h.ListWineTypes)
	r.GET("/wine_types/:id", h.GetWineType)
	r.POST("/wine_types", h.CreateWineType)
	r.PUT("/wine_types/:id", h.UpdateWineType)
	r.DELETE("/wine_types/:id", h.DeleteWineType)

	r.GET("/appellations", h.ListAppellations)
	r.GET("/appellations/:id", h.GetAppellation)
	r.POST("/appellations", h.CreateAppellation)
	r.PUT("/appellations/:id", h.UpdateAppellation)
	r.DELETE("/appellations/:id", h.DeleteAppellation)

	r.GET("/designation_types", h.ListDesignationTypes)
	r.GET("/designation_types/:id", h.GetDesignationType)
	r.POST("/designation_types", h.CreateDesignationType)
	r.PUT("/designation_types/:id", h.UpdateDesignationType)
	r.DELETE("/designation_types/:id", h.DeleteDesignationType)

	r.POST("/label_image", h.UploadLabelImage) // For testing image upload
	return r
}
