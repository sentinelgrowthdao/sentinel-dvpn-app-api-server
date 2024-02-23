package planwizard

import (
	"time"
)

type NodePrice struct {
	Denom  string `json:"denom"`
	Amount int64  `json:"amount"`
}

type Node struct {
	ID uint `json:"id"`

	IsActive bool  `json:"is_active"`
	Revision int64 `json:"revision"`

	IsNodeStatusFetched bool       `json:"is_node_status_fetched"`
	LastNodeStatusFetch *time.Time `json:"last_node_status_fetch"`

	IsNetworkInfoFetched bool       `json:"is_network_info_fetched"`
	LastNetworkInfoFetch *time.Time `json:"last_network_info_fetch"`

	IsHealthChecked bool       `json:"is_health_checked"`
	LastHealthCheck *time.Time `json:"last_health_check"`

	Address        string      `json:"address"`
	RemoteUrl      string      `json:"remote_url"`
	Status         int64       `json:"status"`
	StatusAt       time.Time   `json:"status_at"`
	InactiveAt     time.Time   `json:"inactive_at"`
	GigabytePrices []NodePrice `json:"gigabyte_prices"`
	HourlyPrices   []NodePrice `json:"hourly_prices"`

	Moniker                *string  `json:"moniker"`
	BandwidthUpload        *int64   `json:"bandwidth_upload"`
	BandwidthDownload      *int64   `json:"bandwidth_download"`
	IsHandshakeEnabled     *bool    `json:"is_handshake_enabled"`
	HandshakePeers         *int64   `json:"handshake_peers"`
	IntervalSetSessions    *int64   `json:"interval_set_sessions"`
	IntervalUpdateSessions *int64   `json:"interval_update_sessions"`
	IntervalUpdateStatus   *int64   `json:"interval_update_status"`
	LocationCity           *string  `json:"location_city"`
	LocationCountry        *string  `json:"location_country"`
	LocationLat            *float64 `json:"location_lat"`
	LocationLon            *float64 `json:"location_lon"`
	Operator               *string  `json:"operator"`
	Peers                  *int64   `json:"peers"`
	MaxPeers               *int64   `json:"max_peers"`
	Type                   *int64   `json:"type"`
	Version                *string  `json:"version"`

	ASN           *string `json:"asn"`
	IsResidential *bool   `json:"is_residential"`

	IsHealthy *bool `json:"is_healthy"`
}
