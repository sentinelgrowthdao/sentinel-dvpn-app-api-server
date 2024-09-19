package controllers

import (
	"dvpn/internal/revenuecat"
	"dvpn/middleware"
	"dvpn/models"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"os"
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

	if payload.Address == "" || len(payload.Address) != 43 || payload.Address == os.Getenv("SENTINEL_FEE_GRANTER_WALLET_ADDRESS") {
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

func (wc WalletController) HandleRevenueCatWebhook(c *gin.Context) {

	auth := c.GetHeader("Authorization")
	if auth != os.Getenv("REVENUECAT_AUTH") {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid authorization header")
		return
	}

	var payload revenuecat.Payload
	if err := c.BindJSON(&payload); err != nil {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid request payload: "+err.Error())
		return
	}

	if payload.Event.Id == "" {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid event id")
		return
	}

	if payload.Event.AppUserId == "" || len(payload.Event.AppUserId) != 43 {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid wallet address")
		return
	}

	amount := 0
	denom := "udvpn"

	switch payload.Event.ProductId {
	case "sentinel_dvpn_5000":
		amount = 5000 * 1000000
	case "sentinel_dvpn_10000":
		amount = 10000 * 1000000
	case "sentinel_dvpn_15000":
		amount = 15000 * 1000000
	}

	if amount == 0 {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid product id")
		return
	}

	purchase := models.Purchase{
		EventId: payload.Event.Id,
		Address: strings.ToLower(payload.Event.AppUserId),
		Amount:  int64(amount),
		Denom:   denom,
	}

	tx := wc.DB.Create(&purchase)
	if tx.Error == nil {
		if strings.Contains(tx.Error.Error(), "duplicate key value violates unique constraint") == false {
			reason := "failed to create purchase: " + tx.Error.Error()
			middleware.RespondErr(c, middleware.APIErrorUnknown, reason)
			wc.Logger.Error(reason)
			return
		}
	}

	middleware.RespondOK(c, nil)
}
