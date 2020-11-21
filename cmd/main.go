package main

import (
	"flag"
	"time"

	"github.com/verybada/dcard_take_home_test/internal"
)

func main() {
	var host string
	var redisHost string
	var duration int
	var maxRate int64

	flag.StringVar(&host, "host", ":8080", "server host")
	flag.IntVar(&duration, "duration", 60, "limit duration")
	flag.Int64Var(&maxRate, "max-rate", 60, "max rate per duration")
	flag.StringVar(&redisHost, "redis-host", "127.0.0.1:6379", "redis host")
	flag.Parse()

	internal.Main(
		host, time.Duration(duration)*time.Second,
		maxRate, redisHost)
}
