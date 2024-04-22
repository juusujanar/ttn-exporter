package collector

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

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

func getGatewayList(client *http.Client, uri string, apiKey string) (*[]Gateway, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	u.Path = "/api/v3/gateways"
	u.RawQuery = "limit=100&field_mask=name" // Include gateway name in the response

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Authorization", "Bearer "+apiKey)

	res, err := client.Do(req)
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
		return nil, fmt.Errorf("HTTP status %d on getting gateway list", res.StatusCode)
	}

	var gatewayList GatewayListResponse
	err = json.Unmarshal(body, &gatewayList)
	if err != nil {
		return nil, err
	}

	return &gatewayList.Gateways, nil
}
