package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

func NewDumpRateHandler(logger logrus.FieldLogger) *DumpRateHandler {
	return &DumpRateHandler{
		logger: logger,
	}
}

type DumpRateHandler struct {
	logger logrus.FieldLogger
}

func (h *DumpRateHandler) Dump(
	resp http.ResponseWriter, req *http.Request,
) {
	currentRate, err := getCurrentUsedRate(req.Header)
	if err != nil {
		h.logger.Errorf("get current used rate error: %s", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte(strconv.Itoa(currentRate))) // nolint: errcheck
}

func getCurrentUsedRate(header http.Header) (int, error) {
	limitStr := header.Get("X-RATE-LIMIT-LIMIT")
	if limitStr == "" {
		return 0, fmt.Errorf("header X-RATE-LIMIT-LIMIT not exist")
	}

	remainingStr := header.Get("X-RATE-LIMIT-REMAINING")
	if remainingStr == "" {
		return 0, fmt.Errorf("header X-RATE-LIMIT-REMAINING not exist")
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return 0, fmt.Errorf("convert limit %s error: %s", limitStr, err)
	}

	remaining, err := strconv.Atoi(remainingStr)
	if err != nil {
		return 0, fmt.Errorf("convert remaining %s error: %s", limitStr, err)
	}
	return limit - remaining, nil
}
