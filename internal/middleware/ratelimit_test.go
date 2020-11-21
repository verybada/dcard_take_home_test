package middleware

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/verybada/dcard_take_home_test/internal/ratelimiter"
)

func TestNoRateLimit(t *testing.T) {
	maxRate := int64(100)
	currentRate := rand.Int63n(maxRate)
	mockedLimiter := &ratelimiter.MockRateLimiter{}
	mockedLimiter.On("GetMaxRate").Return(maxRate)
	mockedLimiter.On(
		"Add", mock.Anything).Return(currentRate, nil)

	handler := http.HandlerFunc(
		func(resp http.ResponseWriter, req *http.Request) {
			limit := getHeaderInt64(t, req, "X-RATE-LIMIT-LIMIT")
			require.Equal(t, maxRate, limit)

			remaining := getHeaderInt64(t, req, "X-RATE-LIMIT-REMAINING")
			require.Equal(t, maxRate-currentRate, remaining)

			resp.WriteHeader(999)
		})
	recorder := runWrappedHandler(t, mockedLimiter, handler)
	require.Equal(t, 999, recorder.Code)
}

func TestRateLimited(t *testing.T) {
	mockedLimiter := &ratelimiter.MockRateLimiter{}
	mockedLimiter.On("GetMaxRate").Return(int64(0))
	mockedLimiter.On(
		"Add", mock.Anything).Return(rand.Int63(), nil)
	handler := http.HandlerFunc(
		func(resp http.ResponseWriter, req *http.Request) {
			require.FailNow(t, "should not reach here")
		})
	recorder := runWrappedHandler(t, mockedLimiter, handler)
	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assertBody(t, recorder, "Error")
}

func assertBody(
	t *testing.T, recorder *httptest.ResponseRecorder, expectedBody string,
) {
	body, err := ioutil.ReadAll(recorder.Body)
	require.NoError(t, err)
	require.Equal(t, expectedBody, string(body))
}

func TestRateLimiterError(t *testing.T) {
	mockedLimiter := &ratelimiter.MockRateLimiter{}
	mockedLimiter.On(
		"Add", mock.Anything).Return(
		rand.Int63(), fmt.Errorf("err"))
	handler := http.HandlerFunc(
		func(resp http.ResponseWriter, req *http.Request) {
			require.FailNow(t, "should not reach here")
		})
	recorder := runWrappedHandler(t, mockedLimiter, handler)
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func runWrappedHandler(
	t *testing.T,
	mockedLimiter ratelimiter.RateLimiter, handler http.Handler,
) *httptest.ResponseRecorder {
	wrappedHandler := getWrappedHandler(mockedLimiter, handler)
	recorder := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(
		recorder,
		httptest.NewRequest("GET", "/", nil))
	return recorder
}

func getWrappedHandler(
	mockedLimiter ratelimiter.RateLimiter, handler http.Handler,
) http.Handler {
	middleware := NewRateLimitMiddleware(mockedLimiter, logrus.New())
	return middleware.Do(handler)
}

func getHeaderInt64(t *testing.T, req *http.Request, key string) int64 {
	value, err := strconv.Atoi(req.Header.Get(key))
	require.NoError(t, err)
	return int64(value)
}
