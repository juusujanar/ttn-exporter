package collector

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/go-kit/log"
)

type Duration struct {
	time.Duration
}

// Go seems to have issues parsing time.Duration from a JSON: https://github.com/golang/go/issues/10275
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type GatewayStatsResponse struct {
	Protocol   string `json:"protocol"`
	LastStatus struct {
		Time     time.Time `json:"time"`
		Versions struct {
			GatewayServerVersion *string `json:"ttn-lw-gateway-server"`
			Firmware             *string `json:"firmware"`
			Package              *string `json:"package"`
			Platform             *string `json:"platform"`
			Station              *string `json:"station"`
		} `json:"versions"`
		IP      string `json:"ip"`
		Metrics struct {
			Ackr int64 `json:"ackr"`
			TxIn int64 `json:"txin"`
			TxOk int64 `json:"txok"`
			RxIn int64 `json:"rxin"`
			RxOk int64 `json:"rxok"`
			RxFw int64 `json:"rxfw"`
		} `json:"metrics"`
	} `json:"last_status"`
	UplinkCount            string `json:"uplink_count"`
	DownlinkCount          string `json:"downlink_count"`
	TxAcknowledgementCount string `json:"tx_acknowledgement_count"`
	RoundTripTimes         *struct {
		Min    Duration `json:"min"`
		Max    Duration `json:"max"`
		Median Duration `json:"median"`
		Count  int64    `json:"count"`
	} `json:"round_trip_times"`
	GatewayRemoteAddress struct {
		IP string `json:"ip"`
	} `json:"gateway_remote_address"`
}

func getGatewayStats(client *http.Client, uri string, apiKey string, gatewayID string, logger log.Logger) (*GatewayStatsResponse, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("/api/v3/gs/gateways/%s/connection/stats", gatewayID)

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
		return nil, fmt.Errorf("HTTP status %d on getting gateway stats", res.StatusCode)
	}

	var gatewayStats *GatewayStatsResponse
	err = json.Unmarshal(body, &gatewayStats)
	if err != nil {
		return nil, err
	}

	return gatewayStats, nil
}
