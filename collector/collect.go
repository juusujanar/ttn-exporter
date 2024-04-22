package collector

import (
	"net/http"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const userAgent = "ttn-prometheus-exporter/1.0.0"

type GatewayData struct {
	GatewayID string
	Name      string
	Connected bool
	Stats     *GatewayStatsResponse
}

func GetInfo(client *http.Client, uri string, apiKey string, logger log.Logger) ([]GatewayData, error) {
	gateways, err := getGatewayList(client, uri, apiKey)
	if err != nil {
		return nil, err
	}

	allStats := []GatewayData{}
	for _, gateway := range *gateways {
		gatewayID := gateway.IDs.GatewayID
		gatewayStats, err := getGatewayStats(client, uri, apiKey, gatewayID)
		if err != nil {
			_ = level.Warn(logger).Log("msg", "Failed to scrape gateway", "gatewayID", gatewayID, "err", err)
		}
		connected := err == nil
		allStats = append(allStats, GatewayData{
			GatewayID: gatewayID,
			Name:      gateway.Name,
			Connected: connected,
			Stats:     gatewayStats,
		})
	}

	client.CloseIdleConnections()

	return allStats, err
}
