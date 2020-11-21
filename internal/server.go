package internal

import (
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/verybada/dcard_take_home_test/internal/handler"
	"github.com/verybada/dcard_take_home_test/internal/middleware"
	"github.com/verybada/dcard_take_home_test/internal/ratelimiter"
)

func Main(
	host string, limitDuration time.Duration, maxRate int64,
	redisHost string,
) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	dumpRateHandler := handler.NewDumpRateHandler(logger)

	redisClient := redis.NewClient(&redis.Options{Addr: redisHost})
	ratelimiter := ratelimiter.NewRedisRateLimiter(
		redisClient, limitDuration, maxRate, logger)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(
		ratelimiter, logger)

	router := mux.NewRouter()
	router.Use(rateLimitMiddleware.Do)
	router.PathPrefix("/").HandlerFunc(dumpRateHandler.Dump)

	server := &http.Server{
		Handler: router,
		Addr:    host,
	}
	logger.Infof("server started on %s", host)
	if err := server.ListenAndServe(); err != nil {
		logger.Fatalf("server stopped due to error: %s", err)
	}
	logger.Info("server stopped")
}
