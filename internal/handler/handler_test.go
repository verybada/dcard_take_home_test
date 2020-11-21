package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-RATE-LIMIT-LIMIT", "100")
	req.Header.Set("X-RATE-LIMIT-REMAINING", "60")

	handler := NewDumpRateHandler(logrus.New())
	handler.Dump(recorder, req)

	require.Equal(t, http.StatusOK, recorder.Code)
	data, err := ioutil.ReadAll(recorder.Body)
	require.NoError(t, err)
	require.Equal(t, "40", string(data))
}

func TestPartialHeader(t *testing.T) {
	headers := []string{
		"X-RATE-LIMIT-LIMIT",
		"X-RATE-LIMIT-REMAINING",
	}
	for _, header := range headers {
		t.Run(header, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set(header, "100")

			handler := NewDumpRateHandler(logrus.New())
			handler.Dump(recorder, req)

			require.Equal(t, http.StatusInternalServerError, recorder.Code)
		})
	}
}

func TestHeaderNotInt(t *testing.T) {
	headers := []string{
		"X-RATE-LIMIT-LIMIT",
		"X-RATE-LIMIT-REMAINING",
	}
	for _, header := range headers {
		t.Run(header, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("X-RATE-LIMIT-LIMIT", "100")
			req.Header.Set("X-RATE-LIMIT-REMAINING", "60")
			req.Header.Set(header, "foobar")

			handler := NewDumpRateHandler(logrus.New())
			handler.Dump(recorder, req)

			require.Equal(t, http.StatusInternalServerError, recorder.Code)
		})
	}
}
