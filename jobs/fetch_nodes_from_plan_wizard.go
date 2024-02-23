package jobs

import (
	"dvpn/internal/planwizard"
	"dvpn/models"
	"errors"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

type FetchNodesFromPlanWizard struct {
	DB         *gorm.DB
	Logger     *zap.SugaredLogger
	PlanWizard *planwizard.PlanWizard
}

func (job FetchNodesFromPlanWizard) Run() {
	job.Logger.Infof("fetching nodes from Plan Wizard API")

	nodes, err := job.fetchNodes()
	if err != nil {
		job.Logger.Error("failed to fetch nodes from Plan Wizard API: " + err.Error())
		return
	}

	job.Logger.Infof("fetched %d nodes from Plan Wizard API", len(*nodes))

	revision := time.Now().Unix()

	for _, node := range *nodes {
		protocol, err := job.parseNodeProtocol(&node)
		if err != nil {
			job.Logger.Errorf("failed to determine protocol for %s: %s", node.Address, err)
			continue
		}

		configuration := datatypes.NewJSONType(job.parseNodeConfiguration(&node))
		currentLoad := job.parseCurrentLoad(&node)
		countryId, err := job.parseCountryId(&node)
		if err != nil {
			job.Logger.Errorf("failed to determine country id for %s: %s", node.Address, err)
			continue
		}

		cityId, err := job.parseCityId(&node, countryId)
		if err != nil {
			job.Logger.Errorf("failed to determine city id in country %d for %s: %s", countryId, node.Address, err)
			continue
		}

		var server models.Server
		tx := job.DB.First(&server, "address = ?", node.Address)
		if tx.Error == nil {
			server.Name = *node.Moniker
			server.Address = node.Address
			server.CountryID = countryId
			server.CityID = cityId
			server.Protocol = *protocol
			server.Configuration = configuration
			server.CurrentLoad = currentLoad
			server.IsActive = true
			server.Revision = revision

			tx = job.DB.Save(&server)
			if tx.Error != nil {
				job.Logger.Errorf("failed to update server %s in the DB: %s", node.Address, tx.Error)
			} else {
				job.Logger.Infof("updated DB record for server %s", node.Address)
			}
		} else {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				server = models.Server{
					CountryID:     countryId,
					CityID:        cityId,
					Name:          *node.Moniker,
					Address:       node.Address,
					IsActive:      true,
					IsBanned:      false,
					CurrentLoad:   currentLoad,
					Protocol:      *protocol,
					Configuration: configuration,
					Revision:      revision,
				}

				tx = job.DB.Create(&server)
				if tx.Error != nil {
					job.Logger.Errorf("failed to create server %s in the DB: %s", node.Address, tx.Error)
				} else {
					job.Logger.Infof("created DB record for server %s", node.Address)
				}
			} else {
				job.Logger.Errorf("failed to fetch server %s from database: %s", node.Address, tx.Error)
			}
		}
	}

	tx := job.DB.Model(&models.Server{}).Where("revision != ?", revision).Update("is_active", false)
	if tx.Error != nil {
		job.Logger.Errorf("failed to deactivate inactive servers: %s", tx.Error)
	} else {
		job.Logger.Infof("deactivated %d inactive servers", tx.RowsAffected)
	}
}

func (job FetchNodesFromPlanWizard) fetchNodes() (*[]planwizard.Node, error) {
	var syncInProgress bool
	var limit int
	var offset int

	syncInProgress = true
	limit = 15000
	offset = 0

	var nodes []planwizard.Node

	for syncInProgress {
		n, err := job.PlanWizard.FetchPlanNodes(limit, offset)
		if err != nil {
			return nil, err
		}

		if n == nil || len(*n) < limit {
			syncInProgress = false
		}

		nodes = append(nodes, *n...)
		offset += limit
	}

	return &nodes, nil
}

func (job FetchNodesFromPlanWizard) parseNodePrices(node *planwizard.Node) (int64, int64) {
	var pricePerGB int64
	for _, gigabytePrice := range node.GigabytePrices {
		if gigabytePrice.Denom == "udvpn" {
			pricePerGB = gigabytePrice.Amount
		}
	}

	var pricePerHour int64
	for _, hourlyPrice := range node.HourlyPrices {
		if hourlyPrice.Denom == "udvpn" {
			pricePerHour = hourlyPrice.Amount
		}
	}

	return pricePerGB, pricePerHour
}

func (job FetchNodesFromPlanWizard) parseNodeProtocol(node *planwizard.Node) (*models.ServerProtocol, error) {
	var protocol models.ServerProtocol

	if *node.Type == 1 {
		protocol = models.ServerProtocolWireGuard
		return &protocol, nil
	}

	if *node.Type == 2 {
		protocol = models.ServerProtocolV2Ray
		return &protocol, nil
	}

	return nil, errors.New("unknown protocol")
}

func (job FetchNodesFromPlanWizard) parseNodeConfiguration(node *planwizard.Node) models.ServerConfiguration {
	pricePerGB, pricePerHour := job.parseNodePrices(node)

	return models.ServerConfiguration{
		RemoteURL:         node.RemoteUrl,
		BandwidthDownload: *node.BandwidthDownload,
		BandwidthUpload:   *node.BandwidthUpload,
		LocationCity:      *node.LocationCity,
		LocationCountry:   *node.LocationCountry,
		LocationLat:       *node.LocationLat,
		LocationLon:       *node.LocationLon,
		PricePerGB:        pricePerGB,
		PricePerHour:      pricePerHour,
		Version:           *node.Version,
	}
}

func (job FetchNodesFromPlanWizard) parseCurrentLoad(node *planwizard.Node) float64 {
	currentLoad := float64(*node.Peers) / float64(*node.MaxPeers)
	if currentLoad > 1 {
		currentLoad = 1
	}

	return currentLoad
}

func (job FetchNodesFromPlanWizard) parseCountryId(node *planwizard.Node) (uint, error) {
	countryName := *node.LocationCountry

	var country models.Country
	tx := job.DB.First(&country, "name = ?", countryName)
	if tx.Error != nil {
		return 0, tx.Error
	}

	return country.ID, nil
}

func (job FetchNodesFromPlanWizard) parseCityId(node *planwizard.Node, countryId uint) (uint, error) {
	cityName := *node.LocationCity

	var city models.City
	tx := job.DB.First(&city, "name = ? AND country_id = ?", cityName, countryId)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			city = models.City{
				Name:      *node.LocationCity,
				CountryID: countryId,
			}
			tx = job.DB.Create(&city)
			if tx.Error != nil {
				job.Logger.Errorf("Error creating city %s: %v", city.Name, tx.Error)
				return 0, tx.Error
			}

			return city.ID, nil
		}
		return 0, tx.Error
	}

	return city.ID, nil
}
