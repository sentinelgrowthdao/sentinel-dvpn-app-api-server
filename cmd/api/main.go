package main

import (
	"dvpn/controllers"
	"dvpn/core"
	planwizardAPI "dvpn/internal/planwizard"
	sentinelAPI "dvpn/internal/sentinel"
	"dvpn/jobs"
	"dvpn/models"
	"dvpn/routers"
	"github.com/gin-gonic/gin"
	"github.com/go-co-op/gocron"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

func main() {
	godotenv.Load()

	db, err := core.InitDB()
	if err != nil {
		panic(err)
	}

	err = db.Debug().AutoMigrate(
		&models.Country{},
		&models.City{},
		&models.Server{},
		&models.Network{},
		&models.Wallet{},
	)
	if err != nil {
		panic(err)
	}

	err = core.PopulateDB(db)
	if err != nil {
		panic(err)
	}

	engine := gin.Default()
	err = engine.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	logger, err := core.NewLogger()
	if err != nil {
		panic(err)
	}

	planWizardPlanID, err := strconv.ParseInt(os.Getenv("PLANWIZARD_PLAN_ID"), 10, 64)
	if err != nil {
		panic(err)
	}

	planWizard := &planwizardAPI.PlanWizard{
		APIEndpoint: os.Getenv("PLANWIZARD_API_ENDPOINT"),
		PlanID:      planWizardPlanID,
	}

	gasBase, err := strconv.ParseInt(os.Getenv("SENTINEL_GAS_BASE"), 10, 64)
	if err != nil {
		panic(err)
	}

	sentinel := &sentinelAPI.Sentinel{
		APIEndpoint:              os.Getenv("SENTINEL_API_ENDPOINT"),
		RPCEndpoint:              os.Getenv("SENTINEL_RPC_ENDPOINT"),
		ProviderPlanBlockchainID: os.Getenv("SENTINEL_PROVIDER_PLAN_ID"),
		FeeGranterWalletAddress:  os.Getenv("SENTINEL_FEE_GRANTER_WALLET_ADDRESS"),
		FeeGranterMnemonic:       os.Getenv("SENTINEL_FEE_GRANTER_WALLET_MNEMONIC"),
		DefaultDenom:             os.Getenv("SENTINEL_DEFAULT_DENOM"),
		ChainID:                  os.Getenv("SENTINEL_CHAIN_ID"),
		GasPrice:                 os.Getenv("SENTINEL_GAS_PRICE"),
		GasBase:                  gasBase,
	}

	router := routers.Router{
		HealthController: &controllers.HealthController{
			DB:     db,
			Logger: logger.With("controller", "health"),
		},
		VPNController: &controllers.VPNController{
			DB:     db,
			Logger: logger.With("controller", "vpn"),
		},
		WalletController: &controllers.WalletController{
			DB:     db,
			Logger: logger.With("controller", "wallet"),
		},
	}

	logger.Info("Initializing jobs...")
	if os.Getenv("ENVIRONMENT") != "debug" {
		fetchNodesFromPlanWizard := jobs.FetchNodesFromPlanWizard{
			DB:         db,
			Logger:     logger,
			PlanWizard: planWizard,
		}

		fetchNodesFromPlanWizardScheduler := gocron.NewScheduler(time.UTC)
		fetchNodesFromPlanWizardScheduler.SetMaxConcurrentJobs(1, gocron.RescheduleMode)
		fetchNodesFromPlanWizardScheduler.Every(30).Minutes().Do(func() {
			fetchNodesFromPlanWizard.Run()
		})
		fetchNodesFromPlanWizardScheduler.StartAsync()

		enrollWallets := jobs.EnrollWallets{
			DB:       db,
			Logger:   logger,
			Sentinel: sentinel,
		}

		enrollWalletsScheduler := gocron.NewScheduler(time.UTC)
		enrollWalletsScheduler.SetMaxConcurrentJobs(1, gocron.RescheduleMode)
		enrollWalletsScheduler.Every(1).Seconds().Do(func() {
			enrollWallets.Run()
		})
		enrollWalletsScheduler.StartAsync()
	}

	logger.Info("Registering routes...")
	router.RegisterRoutes(engine)

	logger.Info("Launching API server...")
	engine.Run()
}
