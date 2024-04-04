package controllers

import (
	"dvpn/middleware"
	"dvpn/models"
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type VPNController struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}

func (vc VPNController) GetIPAddress(c *gin.Context) {
	type result struct {
		Ip        string  `json:"ip"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	var ipAddr string

	realIp := c.GetHeader("CF-Connecting-IP")
	if realIp != "" {
		ipAddr = realIp
	} else {
		forwardedIp := c.GetHeader("X-Forwarded-For")
		if forwardedIp != "" {
			parts := strings.Split(forwardedIp, ",")
			if len(parts) > 0 {
				ipAddr = strings.TrimSpace(parts[0])
			}
		} else {
			remoteIp := c.RemoteIP()
			if remoteIp != "" {
				ipAddr = remoteIp
			} else {
				middleware.RespondErr(c, middleware.APIErrorUnknown, "failed to get IP address")
				return
			}
		}
	}

	var network models.Network
	tx := vc.DB.First(&network, "network >> inet '"+ipAddr+"'")
	if tx.Error != nil {
		middleware.RespondErr(c, middleware.APIErrorUnknown, "failed to find matching IP range for "+ipAddr+": "+tx.Error.Error())
		return
	}

	resultObject := result{
		Ip:        ipAddr,
		Latitude:  network.Latitude,
		Longitude: network.Longitude,
	}

	middleware.RespondOK(c, resultObject)
}

func (vc VPNController) GetCountries(c *gin.Context) {
	var countries []models.Country
	var tx *gorm.DB

	protocol := c.Query("protocol")
	if protocol != "" && protocol != "ALL" {
		tx = vc.DB.Raw("SELECT c.id, c.created_at, c.updated_at, c.name, c.code, COUNT(s.id) as servers_available FROM countries AS c INNER JOIN servers AS s ON s.country_id = c.id WHERE s.is_active = true AND s.is_banned = false AND s.protocol = ? GROUP BY c.id ORDER BY c.name", protocol).Scan(&countries)
	} else {
		tx = vc.DB.Raw("SELECT c.id, c.created_at, c.updated_at, c.name, c.code, COUNT(s.id) as servers_available FROM countries AS c INNER JOIN servers AS s ON s.country_id = c.id WHERE s.is_active = true AND s.is_banned = false GROUP BY c.id ORDER BY c.name").Scan(&countries)
	}

	if tx.Error != nil {
		reason := "failed to get countries: " + tx.Error.Error()
		middleware.RespondErr(c, middleware.APIErrorUnknown, reason)
		vc.Logger.Error(reason)
		return
	}

	middleware.RespondOK(c, countries)
}

func (vc VPNController) GetCities(c *gin.Context) {
	countryId, err := strconv.ParseUint(c.Params.ByName("country_id"), 10, 64)
	if err != nil {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid country id: "+err.Error())
		return
	}

	var cities []models.City
	var tx *gorm.DB

	protocol := c.Query("protocol")
	if protocol != "" && protocol != "ALL" {
		tx = vc.DB.Raw("SELECT c.id, c.created_at, c.updated_at, c.country_id, c.name, COUNT(s.id) as servers_available FROM cities AS c INNER JOIN servers AS s ON s.city_id = c.id WHERE s.is_active = true AND s.is_banned = false AND s.protocol = ? AND c.country_id = ? GROUP BY c.id ORDER BY servers_available DESC", protocol, countryId).Scan(&cities)
	} else {
		tx = vc.DB.Raw("SELECT c.id, c.created_at, c.updated_at, c.country_id, c.name, COUNT(s.id) as servers_available FROM cities AS c INNER JOIN servers AS s ON s.city_id = c.id WHERE s.is_active = true AND s.is_banned = false AND c.country_id = ? GROUP BY c.id ORDER BY servers_available DESC", countryId).Scan(&cities)
	}

	if tx.Error != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			middleware.RespondOK(c, []models.City{})
			return
		} else {
			reason := "failed to get cities: " + err.Error()
			middleware.RespondErr(c, middleware.APIErrorUnknown, reason)
			vc.Logger.Error(reason)
			return
		}
	}

	middleware.RespondOK(c, cities)
}

func (vc VPNController) GetServers(c *gin.Context) {
	countryId, err := strconv.ParseUint(c.Params.ByName("country_id"), 10, 64)
	if err != nil {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid country id: "+err.Error())
		return
	}

	cityId, err := strconv.ParseUint(c.Params.ByName("city_id"), 10, 64)
	if err != nil {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid city id: "+err.Error())
		return
	}

	var servers []models.Server

	query := vc.DB.Model(&models.Server{}).Where("country_id = ? AND city_id = ? AND is_active = ? AND is_banned = ?", countryId, cityId, true, false)

	sortBy := c.Query("sortBy")
	if sortBy != "" {
		switch sortBy {
		case "CURRENT_LOAD":
			query = query.Order("current_load desc")
			break
		default:
			middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid sortBy")
			return
		}
	}

	offset := c.Query("offset")
	if offset != "" {
		offset, err := strconv.Atoi(offset)
		if err != nil {
			middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid offset: "+err.Error())
			return
		}

		query = query.Offset(offset)
	}

	limit := c.Query("limit")
	if limit != "" {
		limit, err := strconv.Atoi(limit)
		if err != nil {
			middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid limit: "+err.Error())
			return
		}

		query = query.Limit(limit)
	}

	protocol := c.Query("protocol")
	if protocol != "" && protocol != "ALL" {
		switch protocol {
		case "WIREGUARD", "V2RAY":
			query = query.Where("protocol = ?", protocol)
			break
		default:
			middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid protocol")
			return
		}
	}

	tx := query.Find(&servers)
	if tx.Error != nil {
		reason := "failed to get servers: " + tx.Error.Error()
		middleware.RespondErr(c, middleware.APIErrorUnknown, reason)
		vc.Logger.Error(reason)
		return
	}

	middleware.RespondOK(c, servers)
}

func (vc VPNController) GetServersByIds(c *gin.Context) {
	type requestPayload struct {
		Addresses []string `json:"addresses"`
	}

	var payload requestPayload
	if err := c.BindJSON(&payload); err != nil {
		middleware.RespondErr(c, middleware.APIErrorInvalidRequest, "invalid request payload: "+err.Error())
		return
	}

	var servers []models.Server
	query := vc.DB.Model(&models.Server{}).Where("address = ANY(?)", payload.Addresses)
	tx := query.Find(&servers)
	if tx.Error != nil {
		reason := "failed to get servers: " + tx.Error.Error()
		middleware.RespondErr(c, middleware.APIErrorUnknown, reason)
		vc.Logger.Error(reason)
		return
	}

	middleware.RespondOK(c, servers)
}
