package jobs

import (
	"dvpn/internal/sentinel"
	"dvpn/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EnrollWallets struct {
	DB       *gorm.DB
	Logger   *zap.SugaredLogger
	Sentinel *sentinel.Sentinel
}

func (job EnrollWallets) Run() {
	var wallets []models.Wallet

	tx := job.DB.Model(&models.Wallet{}).Order("id desc").Limit(1000).Where("is_fee_granted = FALSE").Find(&wallets)
	if tx.Error != nil {
		job.Logger.Error("failed to get Sentinel wallets from the DB: " + tx.Error.Error())
		return
	}

	chunks := job.formChunks(wallets, 100)
	for _, chunk := range chunks {

		var walletsForGrantingFee []string

		for _, wallet := range chunk {
			existingAllowances, err := job.fetchAllowances(wallet.Address)
			if err != nil {
				job.Logger.Errorf("failed to fetch existing grant fee allowances from Sentinel for wallet %s: "+err.Error(), wallet.Address)
				continue
			}

			isDeviceAlreadyGrantedWithFee := false

			if len(*existingAllowances) > 0 {
				for _, allowance := range *existingAllowances {
					if allowance.Grantee == wallet.Address && allowance.Granter == job.Sentinel.FeeGranterWalletAddress {
						isDeviceAlreadyGrantedWithFee = true
					}
				}
			}

			if isDeviceAlreadyGrantedWithFee == false {
				walletsForGrantingFee = append(walletsForGrantingFee, wallet.Address)
			}
		}

		if len(walletsForGrantingFee) > 0 {
			err := job.Sentinel.GrantFeeToWallet(walletsForGrantingFee)
			if err != nil {
				job.Logger.Error("failed to grant fee to existing Sentinel wallets: " + err.Error())
				continue
			}
		}

		var walletsToSave []models.Wallet

		for _, wallet := range chunk {
			wallet.IsFeeGranted = true
			walletsToSave = append(walletsToSave, wallet)
		}

		if len(walletsToSave) > 0 {
			tx = job.DB.Save(&walletsToSave)
			if tx.Error != nil {
				job.Logger.Error("failed to update existing wallets: " + tx.Error.Error())
				continue
			}
		}

	}
}

func (job EnrollWallets) formChunks(wallets []models.Wallet, chunkSize int) [][]models.Wallet {
	var chunks [][]models.Wallet
	for chunkSize < len(wallets) {
		wallets, chunks = wallets[chunkSize:], append(chunks, wallets[0:chunkSize:chunkSize])
	}
	chunks = append(chunks, wallets)
	return chunks
}

func (job EnrollWallets) fetchAllowances(walletAddress string) (*[]sentinel.SentinelAllowance, error) {
	var syncInProgress bool
	var limit int
	var offset int

	syncInProgress = true
	limit = 10000
	offset = 0

	var allowances []sentinel.SentinelAllowance

	for syncInProgress {
		n, err := job.Sentinel.FetchFeeGrantAllowances(walletAddress, limit, offset)
		if err != nil {
			return nil, err
		}

		if n == nil {
			syncInProgress = false
		} else {
			if len(*n) < limit {
				syncInProgress = false
			}

			allowances = append(allowances, *n...)
		}

		offset += limit
	}

	return &allowances, nil
}
