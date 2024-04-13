package routers

import (
	"dvpn/controllers"
	"github.com/gin-gonic/gin"
)

type Router struct {
	HealthController *controllers.HealthController
	VPNController    *controllers.VPNController
	WalletController *controllers.WalletController
}

func (r Router) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", r.HealthController.Status)
	router.GET("/version", r.HealthController.Version)
	router.GET("/ip", r.VPNController.GetIPAddress)
	router.GET("/countries", r.VPNController.GetCountries)
	router.GET("/countries/:country_id/cities", r.VPNController.GetCities)
	router.GET("/countries/:country_id/cities/:city_id/servers", r.VPNController.GetServers)
	router.POST("/servers", r.VPNController.GetServersByIds)
	router.POST("/wallet", r.WalletController.RegisterWallet)
}
