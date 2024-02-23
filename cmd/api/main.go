package main

import (
	"dvpn/controllers"
	"dvpn/core"
	planwizardAPI "dvpn/internal/planwizard"
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

	router := routers.Router{
		HealthController: &controllers.HealthController{
			DB:     db,
			Logger: logger.With("controller", "health"),
		},
		VPNController: &controllers.VPNController{
			DB:     db,
			Logger: logger.With("controller", "vpn"),
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
	}

	logger.Info("Registering routes...")
	router.RegisterRoutes(engine)

	logger.Info("Launching API server...")
	engine.Run()
}
