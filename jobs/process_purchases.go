package jobs

import (
	"dvpn/internal/sentinel"
	"dvpn/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
)

type ProcessPurchases struct {
	DB       *gorm.DB
	Logger   *zap.SugaredLogger
	Sentinel *sentinel.Sentinel
}

func (job ProcessPurchases) Run() {
	var purchases []models.Purchase

	tx := job.DB.Model(&models.Purchase{}).Order("id desc").Limit(100).Where("is_redeemed = FALSE").Find(&purchases)
	if tx.Error != nil {
		job.Logger.Error("failed to get purchases from the DB: " + tx.Error.Error())
		return
	}

	chunks := job.formChunks(purchases, 10)
	for _, chunk := range chunks {

		var ids []uint
		var walletAddresses []string
		var amounts []string

		for _, purchase := range chunk {
			ids = append(ids, purchase.ID)
			walletAddresses = append(walletAddresses, purchase.Address)
			amounts = append(amounts, strconv.Itoa(int(purchase.Amount))+purchase.Denom)
		}

		err := job.Sentinel.SendTokensToWallet(walletAddresses, amounts)
		if err != nil {
			job.Logger.Error("failed to send tokens to wallets: " + err.Error())
			continue
		}

		err = job.DB.Model(&models.Purchase{}).Where("id IN ?", ids).Updates(map[string]interface{}{"is_redeemed": true}).Error
		if err != nil {
			job.Logger.Error("failed to update purchases: " + err.Error())
		}
	}
}

func (job ProcessPurchases) formChunks(purchases []models.Purchase, chunkSize int) [][]models.Purchase {
	var chunks [][]models.Purchase
	for chunkSize < len(purchases) {
		purchases, chunks = purchases[chunkSize:], append(chunks, purchases[0:chunkSize:chunkSize])
	}
	chunks = append(chunks, purchases)
	return chunks
}
