package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

const gatewayListPageSize = 100

type Gateway struct {
	IDs struct {
		GatewayID string `json:"gateway_id"`
		EUI       string `json:"eui"`
	} `json:"ids"`
	Name string `json:"name"`
}

type GatewayListResponse struct {
	Gateways []Gateway `json:"gateways"`
}

// getGatewayList fetches all gateways, following pagination until a page
// comes back shorter than gatewayListPageSize.
func getGatewayList(client *http.Client, uri string, apiKey string, logger *slog.Logger) ([]Gateway, error) {
	var all []Gateway
	for page := 1; ; page++ {
		gateways, err := getGatewayListPage(client, uri, apiKey, page, logger)
		if err != nil {
			return nil, err
		}
		all = append(all, gateways...)
		if len(gateways) < gatewayListPageSize {
			return all, nil
		}
	}
}

func getGatewayListPage(client *http.Client, uri string, apiKey string, page int, logger *slog.Logger) ([]Gateway, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	u.Path = "/api/v3/gateways"
	query := url.Values{}
	query.Set("limit", strconv.Itoa(gatewayListPageSize))
	query.Set("page", strconv.Itoa(page))
	query.Set("field_mask", "name") // Include gateway name in the response
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	res, err := doWithRateLimitRetry(client, req, logger)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = res.Body.Close()
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d on getting gateway list (page %d)", res.StatusCode, page)
	}

	var gatewayList GatewayListResponse
	err = json.Unmarshal(body, &gatewayList)
	if err != nil {
		return nil, err
	}

	return gatewayList.Gateways, nil
}
