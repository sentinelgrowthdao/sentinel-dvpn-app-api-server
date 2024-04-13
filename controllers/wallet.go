package controllers

import (
	"dvpn/middleware"
	"dvpn/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strings"
)

type WalletController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (wc WalletController) RegisterWallet(c *gin.Context) {
	type requestPayload struct {
		Address string `json:"address"`
	}

	var payload requestPayload
	if err := c.BindJSON(&payload); err != nil {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid request payload: "+err.Error())
		return
	}

	if payload.Address == "" || len(payload.Address) != 43 {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid wallet address")
		return
	}

	wallet := models.Wallet{
		Address:      strings.ToLower(payload.Address),
		IsFeeGranted: false,
	}

	tx := wc.DB.Create(&wallet)
	if tx.Error != nil {
		if strings.Contains(tx.Error.Error(), "duplicate key value violates unique constraint") == false {
			reason := "failed to create wallet: " + tx.Error.Error()
			middleware.RespondErr(c, middleware.APIErrorUnknown, reason)
			wc.Logger.Error(reason)
			return
		}
	}

	middleware.RespondOK(c, nil)
}
