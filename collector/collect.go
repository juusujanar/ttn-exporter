package collector

import (
	"log/slog"
	"net/http"
)

const userAgent = "ttn-prometheus-exporter/1.0.0"

type GatewayData struct {
	GatewayID string
	Name      string
	Connected bool
	Stats     *GatewayStatsResponse
}

func GetInfo(client *http.Client, uri string, apiKey string, logger *slog.Logger) ([]GatewayData, error) {
	gateways, err := getGatewayList(client, uri, apiKey)
	if err != nil {
		return nil, err
	}

	allStats := []GatewayData{}
	for _, gateway := range *gateways {
		gatewayID := gateway.IDs.GatewayID
		gatewayStats, err := getGatewayStats(client, uri, apiKey, gatewayID)
		if err != nil {
			logger.Warn("Failed to scrape gateway", "gatewayID", gatewayID, "err", err.Error())
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
