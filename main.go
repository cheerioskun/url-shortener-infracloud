package main

import (
	"sync"

	"github.com/gin-gonic/gin"
)

const ServerListenAddr = "0.0.0.0:3000"

var db sync.Map
var metrics sync.Map

func main() {
	r := setupRouter()
	r.Run(ServerListenAddr)
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.POST("/short", ShortHandler)
	r.GET("/long/:blurb", LongHandler)
	r.GET("/metrics", MetricsHandler)
	return r
}
