package middleware

import (
	"fmt"
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/verybada/dcard_take_home_test/internal/ratelimiter"
)

func NewRateLimitMiddleware(
	rateLimiter ratelimiter.RateLimiter, logger logrus.FieldLogger,
) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		rateLimiter: rateLimiter,
		logger:      logger,
	}
}

type RateLimitMiddleware struct {
	rateLimiter ratelimiter.RateLimiter
	logger      logrus.FieldLogger
}

func (m *RateLimitMiddleware) Do(next http.Handler) http.Handler {
	return http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		ipAddress, err := getIPAddress(req)
		if err != nil {
			m.logger.Errorf("get ip address error: %s", err)
			resp.WriteHeader(http.StatusInternalServerError)
		}

		rate, err := m.rateLimiter.Add(ipAddress)
		if err != nil {
			m.logger.Errorf("add and validate %s error: %s", ipAddress, err)
			resp.WriteHeader(http.StatusInternalServerError)
			return
		}

		maxRate := m.rateLimiter.GetMaxRate()
		if rate > maxRate {
			m.logger.Warnf("%s is blocked due to too many request", ipAddress)
			resp.WriteHeader(http.StatusTooManyRequests)
			resp.Write([]byte("Error")) // nolint: errcheck
			return
		}
		req.Header.Set("X-RATE-LIMIT-LIMIT", fmt.Sprintf("%d", maxRate))
		req.Header.Set("X-RATE-LIMIT-REMAINING",
			fmt.Sprintf("%d", maxRate-rate))
		next.ServeHTTP(resp, req)
	})
}

func getIPAddress(req *http.Request) (string, error) {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return "", fmt.Errorf("split request host and port error: %s", err)
	}
	return host, nil
}
