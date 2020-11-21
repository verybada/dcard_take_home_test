package ratelimiter

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	//"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetMaxRate(t *testing.T) {
	rate := rand.Int63()
	limiter := NewRedisRateLimiter(
		nil, time.Second, rate, logrus.New())

	require.Equal(t, rate, limiter.GetMaxRate())
}

func TestAddNewIP(t *testing.T) {
	server := startRedisServer(t)
	defer server.Close()

	rate := int64(10)
	ipAddress := "foobar"
	duration := time.Hour
	windowTime := time.Now().UTC().Truncate(duration)

	limiter := newLimiter(rate, duration, server.Addr())
	assertAdd(t, limiter, ipAddress, int64(1))

	key := fmt.Sprintf("%s-%d", ipAddress, windowTime.Unix())
	value, err := server.Get(key)
	require.NoError(t, err)
	require.Equal(t, "1", value)

	ttl := server.TTL(key)
	require.NotEmpty(t, ttl)
}

func TestAddOverMaxRate(t *testing.T) {
	server := startRedisServer(t)
	defer server.Close()

	rate := int64(10)
	ipAddress := "foobar"
	duration := time.Hour
	limiter := newLimiter(rate, duration, server.Addr())

	for i := int64(1); i <= rate; i++ {
		assertAdd(t, limiter, ipAddress, int64(i))
	}

	for i := int64(1); i <= rate; i++ {
		assertAdd(t, limiter, ipAddress, rate+i)
	}
}

func TestAddMultipleIPs(t *testing.T) {
	server := startRedisServer(t)
	defer server.Close()

	rate := int64(1)
	ipAddress := "foobar"
	duration := time.Hour
	limiter := newLimiter(rate, duration, server.Addr())

	for i := 0; i < 100; i++ {
		ipAddress := fmt.Sprintf("%s_%d", ipAddress, i)
		assertAdd(t, limiter, ipAddress, int64(1))
	}

	require.Len(t, server.Keys(), 100)
}

func TestAddOverDuration(t *testing.T) {
	server := startRedisServer(t)
	defer server.Close()

	rate := int64(1)
	ipAddress := "foobar"
	duration := 3 * time.Second
	limiter := newLimiter(rate, duration, server.Addr())
	assertAdd(t, limiter, ipAddress, int64(1))

	keys := server.Keys()
	require.Len(t, keys, 1)

	now := time.Now()
	nextWindowTime := now.Truncate(duration).Add(duration)
	diff := nextWindowTime.Sub(now)
	time.Sleep(diff)
	server.FastForward(duration)

	assertAdd(t, limiter, ipAddress, int64(1))

	keysAfter := server.Keys()
	require.Len(t, keysAfter, 1)

	require.NotEqual(t, keys, keysAfter)
}

func startRedisServer(t *testing.T) *miniredis.Miniredis {
	server, err := miniredis.Run()
	require.NoError(t, err)
	return server
}

func newLimiter(
	rate int64, duration time.Duration, redisHost string,
) RateLimiter {
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	return NewRedisRateLimiter(
		redisClient, duration, rate, logrus.New())
}

func assertAdd(
	t *testing.T, limiter RateLimiter,
	ipAddress string, expectedRate int64,
) {
	rate, err := limiter.Add(ipAddress)
	require.NoError(t, err)
	require.Equal(t, expectedRate, rate)
}
