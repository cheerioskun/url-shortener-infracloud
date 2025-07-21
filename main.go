package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/blake2b"
)

const ServerListenAddr = "localhost:3000"

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run(ServerListenAddr)
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
	// Hash it!
	hasher, _ := blake2b.New256([]byte("Hey InfraCloud!"))
	_, _ = hasher.Write([]byte(req.URL))
	hash := hasher.Sum(nil)
	// Encode it base64
	encodedStr := base64.URLEncoding.EncodeToString(hash)
	// Form new URL
	newURL := fmt.Sprintf("%s/long/%s", ServerListenAddr, encodedStr)
	c.IndentedJSON(http.StatusOK, ShortResponse{
		URL: newURL,
	})
	// TODO: Update metrics
	// TODO: Store mapping from hash to long URL

}
