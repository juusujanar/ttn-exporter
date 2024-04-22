package main

import (
	"crypto/tls"
	"errors"
	"net/http"
	_ "net/http/pprof" // #nosec G108 profiling error
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/juusujanar/ttn-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"
)

const (
	// Exporter namespace.
	namespace                = "ttn"
	envThingsNetworkAPIToken = "TTN_API_KEY" // #nosec G101 this is not a hardcoded credential
)

// Metrics descriptors.
var (
	gatewayConnected = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "gateway", "connected"),
		"Gateway connection status",
		[]string{"gateway_id", "name"}, nil,
	)
	uplinkCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "gateway", "uplink_count"),
		"Number of uplink packets received by gateway",
		[]string{"gateway_id", "name"}, nil,
	)
	downlinkCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "gateway", "downlink_count"),
		"Number of downlink packets sent by gateway",
		[]string{"gateway_id", "name"}, nil,
	)
	txAcknowledgementCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "gateway", "tx_acknowledgement_count"),
		"Number of TX acknowledgements received by gateway",
		[]string{"gateway_id", "name"}, nil,
	)
	roundTripMin = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "gateway", "round_trip_min"),
		"Minimum measured round trip time between gateway and TTN in seconds",
		[]string{"gateway_id", "name"}, nil,
	)
	roundTripMax = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "gateway", "round_trip_max"),
		"Maximum measured round trip time between gateway and TTN in seconds",
		[]string{"gateway_id", "name"}, nil,
	)
	roundTripMedian = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "gateway", "round_trip_median"),
		"Median measured round trip time between gateway and TTN in seconds",
		[]string{"gateway_id", "name"}, nil,
	)
	roundTripCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "gateway", "round_trip_count"),
		"Total round trip measurements between gateway and TTN",
		[]string{"gateway_id", "name"}, nil,
	)

	ttnInfo = prometheus.NewDesc(prometheus.BuildFQName(namespace, "version", "info"), "Things Network version info.", []string{"version"}, nil)
	ttnUp   = prometheus.NewDesc(prometheus.BuildFQName(namespace, "", "up"), "Was the last scrape of TTN successful.", nil, nil)
)

// Exporter collects HAProxy stats from the given URI and exports them using
// the prometheus metrics package.
type Exporter struct {
	URI          string
	apiKey       string
	client       *http.Client
	up           prometheus.Gauge
	totalScrapes prometheus.Counter
	logger       log.Logger
}

// NewExporter returns an initialized Exporter.
func NewExporter(uri string, apiKey string, sslVerify bool, timeout time.Duration, logger log.Logger) (*Exporter, error) {
	if !strings.HasPrefix(uri, "https://") && !strings.HasPrefix(uri, "http://") {
		return nil, errors.New("invalid URI scheme")
	}

	return &Exporter{
		URI:    uri,
		apiKey: apiKey,
		client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !sslVerify, // #nosec G402 -- allow insecure TLS when requested by user
				},
			},
		},
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "up",
			Help: "Was the last scrape of TTN successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "exporter_scrapes_total",
			Help: "Total number of scrapes.",
		}),
		logger: logger,
	}, nil
}

// Describe describes all the metrics ever exported by the HAProxy exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// for _, m := range e.serverMetrics {
	// 	ch <- m.Desc
	// }
	ch <- ttnInfo
	ch <- ttnUp
	ch <- e.totalScrapes.Desc()
}

// Collect fetches the stats from configured HAProxy location and delivers them
// as Prometheus metrics. It implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	up := e.scrape(ch)

	ch <- prometheus.MustNewConstMetric(ttnUp, prometheus.GaugeValue, up)
	ch <- e.totalScrapes
}

func (e *Exporter) scrape(ch chan<- prometheus.Metric) (up float64) {
	e.totalScrapes.Inc()
	var err error
	data, err := collector.GetInfo(e.client, e.URI, e.apiKey, e.logger)
	if err != nil {
		_ = level.Error(e.logger).Log("msg", "Can't scrape TTN", "err", err)
		return 0
	}

	e.parseVersion(ch, data)

	for _, gw := range data {
		if gw.Stats != nil {
			ch <- prometheus.MustNewConstMetric(gatewayConnected, prometheus.GaugeValue, BoolToFloat(gw.Connected), gw.GatewayID, gw.Name)
			uplinkCountFloat, err := strconv.ParseFloat(gw.Stats.UplinkCount, 64)
			if err != nil {
				_ = level.Error(e.logger).Log("msg", "Failed to convert UplinkCount to float64", "err", err)
			}
			downlinkCountFloat, err := strconv.ParseFloat(gw.Stats.DownlinkCount, 64)
			if err != nil {
				_ = level.Error(e.logger).Log("msg", "Failed to convert DownlinkCount to float64", "err", err)
			}
			txAcknowledgementCountFloat, err := strconv.ParseFloat(gw.Stats.TxAcknowledgementCount, 64)
			if err != nil {
				_ = level.Error(e.logger).Log("msg", "Failed to convert TxAcknowledgementCount to float64", "err", err)
			}
			ch <- prometheus.MustNewConstMetric(uplinkCount, prometheus.CounterValue, uplinkCountFloat, gw.GatewayID, gw.Name)
			ch <- prometheus.MustNewConstMetric(downlinkCount, prometheus.CounterValue, downlinkCountFloat, gw.GatewayID, gw.Name)
			ch <- prometheus.MustNewConstMetric(txAcknowledgementCount, prometheus.CounterValue, txAcknowledgementCountFloat, gw.GatewayID, gw.Name)
			ch <- prometheus.MustNewConstMetric(roundTripMin, prometheus.GaugeValue, gw.Stats.RoundTripTimes.Min.Seconds(), gw.GatewayID, gw.Name)
			ch <- prometheus.MustNewConstMetric(roundTripMax, prometheus.GaugeValue, gw.Stats.RoundTripTimes.Max.Seconds(), gw.GatewayID, gw.Name)
			ch <- prometheus.MustNewConstMetric(roundTripMedian, prometheus.GaugeValue, gw.Stats.RoundTripTimes.Median.Seconds(), gw.GatewayID, gw.Name)
			ch <- prometheus.MustNewConstMetric(roundTripCount, prometheus.GaugeValue, float64(gw.Stats.RoundTripTimes.Count), gw.GatewayID, gw.Name)
		} else {
			ch <- prometheus.MustNewConstMetric(gatewayConnected, prometheus.GaugeValue, 0, gw.GatewayID, gw.Name)
		}
	}
	return 1
}

func (e *Exporter) parseVersion(ch chan<- prometheus.Metric, data []collector.GatewayData) {
	for _, gw := range data {
		if gw.Stats != nil {
			if gw.Stats.LastStatus.Versions.GatewayServerVersion != nil {
				ch <- prometheus.MustNewConstMetric(ttnInfo, prometheus.GaugeValue, 1, *gw.Stats.LastStatus.Versions.GatewayServerVersion)
				return
			} else if gw.Stats.LastStatus.Versions.Firmware != nil {
				ch <- prometheus.MustNewConstMetric(ttnInfo, prometheus.GaugeValue, 1, *gw.Stats.LastStatus.Versions.Firmware)
				return
			}
		}
	}
}

func BoolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func main() {
	var (
		webConfig    = webflag.AddFlags(kingpin.CommandLine, ":9981")
		metricsPath  = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
		ttnURI       = kingpin.Flag("ttn.uri", "URI on which Things Stack is used.").Default("https://eu1.cloud.thethings.network/").String()
		ttnSSLVerify = kingpin.Flag("ttn.ssl-verify", "Flag that enables SSL certificate verification to the TTN API URI").Default("true").Bool()
		ttnTimeout   = kingpin.Flag("ttn.timeout", "Timeout for trying to get stats from TTN API.").Default("5s").Duration()
	)

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("ttn_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
	logger := promlog.New(promlogConfig)

	_ = level.Info(logger).Log("msg", "Starting ttn_exporter", "version", version.Info())
	_ = level.Info(logger).Log("msg", "Build context", "context", version.BuildContext())

	ttnAPIKey := os.Getenv(envThingsNetworkAPIToken)
	if ttnAPIKey == "" {
		_ = level.Error(logger).Log("msg", "TTN API key not set as environment variable", "variable", envThingsNetworkAPIToken)
		os.Exit(1)
	}

	exporter, err := NewExporter(*ttnURI, ttnAPIKey, *ttnSSLVerify, *ttnTimeout, logger)
	if err != nil {
		_ = level.Error(logger).Log("msg", "Error creating an exporter", "err", err)
		os.Exit(1)
	}
	registry := prometheus.NewRegistry()
	registry.MustRegister(exporter)
	registry.MustRegister(version.NewCollector("ttn_exporter"))

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
      <head><title>TTN Exporter</title></head>
      <body>
        <h1>The Things Network Exporter</h1>
        <p><a href='` + *metricsPath + `'>Metrics</a></p>
        <p><a href='` + *ttnURI + `'>Things Stack</a></p>
        <p><a href='https://github.com/juusujanar/ttn-exporter'>GitHub</a></p>
      </body>
    </html>`))
	})
	srv := &http.Server{
		ReadHeaderTimeout: 1 * time.Second,
	}
	if err := web.ListenAndServe(srv, webConfig, logger); err != nil {
		_ = level.Error(logger).Log("msg", "Error starting HTTP server", "err", err)
		os.Exit(1)
	}
}
