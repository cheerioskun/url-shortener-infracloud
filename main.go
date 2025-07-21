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

const ServerListenAddr = "localhost:3000"

var db sync.Map
var metrics map[string]*atomic.Int64 = make(map[string]*atomic.Int64)

func main() {
	r := setupRouter()
	r.Run(ServerListenAddr)
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
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
	go metrics[longURL.Host].Add(1)

}

func LongHandler(c *gin.Context) {
	// Check for valid shortURL
}
