package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/blake2b"
)

const ServerListenAddr = "http://localhost:3000"

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

	return r
}

type ShortRequest struct {
	URL string `json:"URL"`
}

type ShortResponse struct {
	URL string `json:"URL"`
}

func ShortHandler(c *gin.Context) {
	var req ShortRequest
	if err := c.BindJSON(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	// Now we parse it.
	longURL, err := url.Parse(req.URL)
	if err != nil {
		slog.Error("could not parse URL", "url", req.URL, "error", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	slog.Info("URL parsed successfully", "url", req.URL, "domain", longURL.Host, "path", longURL.Path)

	// Hash it and encode to base64
	hasher, _ := blake2b.New256([]byte("Hey InfraCloud!"))
	_, _ = hasher.Write([]byte(req.URL))
	hash := hasher.Sum(nil)
	encodedStr := base64.URLEncoding.EncodeToString(hash)

	// Form new URL
	newURL := fmt.Sprintf("%s/long/%s", ServerListenAddr, encodedStr)
	// Save in db/map
	db.Store(encodedStr, req.URL)
	slog.Debug("new URL mapping stored", "short", newURL, "long", req.URL)

	c.IndentedJSON(http.StatusOK, ShortResponse{
		URL: newURL,
	})

	// Triggering an async routine, there is no chance of this getting stuck in a deadlock,
	// So no need to wrap a context and timeout. This should not cause goroutine leak.
	go IncrementCounter(&metrics, longURL.Host)
}

func LongHandler(c *gin.Context) {
	// Check for valid shortURL
	blurb := c.Param("blurb")
	if val, ok := db.Load(blurb); ok {
		c.Header("Location", val.(string))
		c.Status(http.StatusTemporaryRedirect)
		slog.Info("found short to long mapping", "short", blurb, "long", val.(string))
		return
	}
	slog.Info("not found short to long mapping", "short", blurb)
	c.AbortWithStatus(http.StatusNotFound)
}

func MetricsHandler(c *gin.Context) {
	// Simple for now, just flatten the map then sort

}

func IncrementCounter(counters *sync.Map, key string) {
	val, _ := counters.LoadOrStore(key, new(int64))
	ptr := val.(*int64)
	atomic.AddInt64(ptr, 1)
}
