package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/blake2b"
)

var db sync.Map
var metrics sync.Map

const ShortenedLen = 10 // Length of the blurb

type ShortRequest struct {
	URL string `json:"URL"`
}

type ShortResponse struct {
	URL string `json:"URL"`
}

// Handler for POST /short
// Args:
// - URL string (required) - Long URL to be shortened
// Returns:
// - URL string - Complete shortened URL redirecting to the long input
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
	encodedStr := base64.URLEncoding.EncodeToString(hash)[:ShortenedLen]

	// Form new URL
	newURL := fmt.Sprintf("http://%s/long/%s", ServerListenAddr, encodedStr)
	// Save in db/map
	db.Store(encodedStr, req.URL)
	slog.Debug("new URL mapping stored", "short", newURL, "long", req.URL)

	c.IndentedJSON(http.StatusOK, ShortResponse{
		URL: newURL,
	})

	IncrementCounter(&metrics, longURL.Host)
}

// Handler for GET /long/:blurb
// Accepts a URL shortened with the same service and redirects to the
// original mapping by returning a 3xx HTTP Status with Location header.
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

// Handler for GET /metrics
// Returns the top 3 shortened domains
func MetricsHandler(c *gin.Context) {
	// Simple for now, just flatten the map then sort
	type Entry struct {
		Name  string
		Count int64
	}
	flat := make([]Entry, 0)
	metrics.Range(func(key, value any) bool {
		flat = append(flat, Entry{Name: key.(string), Count: value.(*atomic.Int64).Load()})
		return true
	})
	sort.SliceStable(flat, func(i, j int) bool {
		return flat[i].Count > flat[j].Count
	})
	for i := range len(flat) {
		if i == 3 {
			break
		}
		_, _ = fmt.Fprintf(c.Writer, "%s: %d\n", flat[i].Name, flat[i].Count)
	}
	c.Status(http.StatusOK)
}

// IncrementCounter is a utility function to concurrent-safely update
// atomic counters in a sync.Map. This is used for metrics accounting.
func IncrementCounter(counters *sync.Map, key string) {
	val, _ := counters.LoadOrStore(key, new(atomic.Int64))
	ptr := val.(*atomic.Int64)
	ptr.Add(1)
}
