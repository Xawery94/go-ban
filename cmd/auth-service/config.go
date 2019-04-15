package main

import (
	"flag"
	"time"
)

var (
	redisURL     = flag.String("redisUrl", "localhost:6379", "Redis Url")
	addr         = flag.String("addr", ":80", "Listen address")
	readTimeout  = flag.Duration("readTimeout", 10*time.Second, "TCP read timeout")
	writeTimeout = flag.Duration("writeTimeout", 10*time.Second, "TCP write timeout")
)
