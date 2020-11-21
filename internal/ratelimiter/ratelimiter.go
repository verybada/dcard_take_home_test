package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RateLimiter interface {
	Add(ipAddr string) (int64, error)
	GetMaxRate() int64
}

func NewRedisRateLimiter(
	redisClient *redis.Client, duration time.Duration, maxRate int64,
	logger logrus.FieldLogger,
) RateLimiter {
	return &redisRateLimiter{
		ctx:      context.Background(),
		client:   redisClient,
		duration: duration,
		maxRate:  maxRate,
		logger:   logger,
	}
}

type redisRateLimiter struct {
	ctx      context.Context
	client   *redis.Client
	duration time.Duration
	maxRate  int64
	logger   logrus.FieldLogger
}

func (l *redisRateLimiter) GetMaxRate() int64 {
	return l.maxRate
}

func (l *redisRateLimiter) Add(ipAddress string) (int64, error) {
	now := time.Now().UTC()

	rate, err := l.getCurrentWindowRate(ipAddress, now)
	if err != nil {
		return 0, fmt.Errorf("get current window rate error: %s", err)
	}
	return rate, nil
}

func (l *redisRateLimiter) getCurrentWindowRate(
	ipAddress string, now time.Time,
) (int64, error) {
	windowTime := l.getCurrentWindowTime(now)
	key := l.getWindowKey(ipAddress, windowTime)
	rate, err := l.client.Incr(l.ctx, key).Result()
	if rate == 1 {
		nextWindowTime := windowTime.Add(l.duration)
		l.client.ExpireAt(l.ctx, key, nextWindowTime)
	}
	l.logger.Debugf("current window %d key %s rate %d",
		windowTime.Unix(), key, rate)
	return rate, err
}

func (l *redisRateLimiter) getCurrentWindowTime(now time.Time) time.Time {
	return now.Truncate(l.duration)
}

func (l *redisRateLimiter) getWindowKey(
	ipAddress string, t time.Time,
) string {
	return fmt.Sprintf("%s-%d", ipAddress, t.Unix())
}
