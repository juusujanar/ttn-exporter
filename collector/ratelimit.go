package collector

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"
)

const (
	rateLimitRetryHeader = "X-Rate-Limit-Retry"
	maxRateLimitRetries  = 2
	defaultRateLimitWait = time.Second
	maxRateLimitWait     = 10 * time.Second
)

// doWithRateLimitRetry performs req and retries it when the TTN API responds
// with HTTP 429, honoring the X-Rate-Limit-Retry header (seconds to wait).
// req must have a nil body (e.g. a GET request) since it may be sent more than once.
func doWithRateLimitRetry(client *http.Client, req *http.Request, logger *slog.Logger) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		res, err := client.Do(req)
		if err != nil || res.StatusCode != http.StatusTooManyRequests || attempt >= maxRateLimitRetries {
			return res, err
		}

		wait := defaultRateLimitWait
		if v := res.Header.Get(rateLimitRetryHeader); v != "" {
			if seconds, convErr := strconv.Atoi(v); convErr == nil {
				wait = time.Duration(seconds) * time.Second
			}
		}
		if wait > maxRateLimitWait {
			wait = maxRateLimitWait
		}

		_ = res.Body.Close()
		logger.Warn("Rate limited by TTN API, retrying", "path", req.URL.Path, "wait", wait, "attempt", attempt+1)
		time.Sleep(wait)
	}
}
