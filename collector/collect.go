package collector

import (
	"net/http"

	"github.com/go-kit/log"
)

const userAgent = "ttn-prometheus-exporter/1.0.0"

type GatewayData struct {
	GatewayID string
	Name      string
	Connected bool
	Stats     *GatewayStatsResponse
}

func GetInfo(client *http.Client, uri string, apiKey string, logger log.Logger) ([]GatewayData, error) {
	gateways, err := getGatewayList(client, uri, apiKey, logger)
	if err != nil {
		return nil, err
	}

	allStats := []GatewayData{}
	for _, gateway := range gateways {
		gatewayID := gateway.IDs.GatewayID
		gatewayStats, err := getGatewayStats(client, uri, apiKey, gatewayID, logger)
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
