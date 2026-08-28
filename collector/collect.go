package collector

import (
	"log/slog"
	"net/http"

	"golang.org/x/sync/errgroup"
)

const userAgent = "ttn-prometheus-exporter/1.0.0"

type GatewayData struct {
	GatewayID string
	Name      string
	Connected bool
	Stats     *GatewayStatsResponse
}

// GetInfo lists all gateways and fetches each one's connection stats
// concurrently, bounded by concurrency. A single gateway's stats fetch
// failing only marks that gateway as disconnected; it does not fail the
// whole scrape.
func GetInfo(client *http.Client, uri string, apiKey string, concurrency int, logger *slog.Logger) ([]GatewayData, error) {
	gateways, err := getGatewayList(client, uri, apiKey, logger)
	if err != nil {
		return nil, err
	}

	allStats := make([]GatewayData, len(gateways))

	g := new(errgroup.Group)
	g.SetLimit(concurrency)

	for i, gateway := range gateways {
		g.Go(func() error {
			gatewayID := gateway.IDs.GatewayID
			gatewayStats, err := getGatewayStats(client, uri, apiKey, gatewayID, logger)
			if err != nil {
				logger.Warn("Failed to scrape gateway", "gatewayID", gatewayID, "err", err.Error())
			}
			allStats[i] = GatewayData{
				GatewayID: gatewayID,
				Name:      gateway.Name,
				Connected: err == nil,
				Stats:     gatewayStats,
			}
			return nil
		})
	}
	_ = g.Wait()

	return allStats, nil
}
