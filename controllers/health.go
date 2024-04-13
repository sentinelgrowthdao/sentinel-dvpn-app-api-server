package controllers

import (
	"dvpn/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
)

type HealthController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (h HealthController) Status(c *gin.Context) {
	err := h.DB.Raw(`SELECT 1`).Row().Err()
	if err != nil {
		h.Logger.Errorf("Error checking database health: %v", err)
		middleware.RespondErr(c, middleware.APIErrorUnknown, "Error checking database health")
		return
	}

	middleware.RespondOK(c, nil)
}

func (h HealthController) Version(c *gin.Context) {
	middleware.RespondOK(c, gin.H{
		"ios":     os.Getenv("LAST_IOS_VERSION"),
		"android": os.Getenv("LAST_ANDROID_VERSION"),
	})
}
