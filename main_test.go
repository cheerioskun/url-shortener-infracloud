package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFunctionalE2E(t *testing.T) {
	type testcase struct {
		name string
		run  func(t *testing.T)
	}
	longURL := "https://github.com/cheerioskun/constellation"
	testJSONSingle := &ShortRequest{
		URL: longURL,
	}
	testJSONMultiple := []*ShortRequest{
		{
			URL: "https://github.com/cheerioskun/constellation",
		},
		{
			URL: "https://linkedin.com/in/hemant-pandey-hx",
		},
		{
			URL: "https://www.linkedin.com/company/infracloudio/",
		},
		{
			URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		},
		{
			URL: "https://www.youtube.com/watch?v=kPaJfAUwViY",
		},
		{
			URL: "https://www.youtube.com/watch?v=LcWoP6KtZKw",
		},
	}
	testcases := []testcase{
		{
			name: "Call to shorten and then check if long returns the same",
			run: func(t *testing.T) {
				r := setupRouter()
				w := httptest.NewRecorder()
				body, _ := json.Marshal(testJSONSingle)
				req, _ := http.NewRequest("POST", "/short", bytes.NewReader(body))
				r.ServeHTTP(w, req)

				assert.Equal(t, w.Code, http.StatusOK)
				resp := &ShortResponse{}
				err := json.Unmarshal(w.Body.Bytes(), resp)
				assert.NoError(t, err)

				shortURL, err := url.Parse(resp.URL)
				assert.NoError(t, err)

				req, _ = http.NewRequest("GET", shortURL.Path, nil)

				w = httptest.NewRecorder()
				r.ServeHTTP(w, req)
				assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
				assert.Equal(t, testJSONSingle.URL, w.Header().Get("Location"))
			},
		},
		{
			name: "Call to shorten multiple times, then check metrics",
			run: func(t *testing.T) {
				r := setupRouter()
				for _, testJson := range testJSONMultiple {
					w := httptest.NewRecorder()
					body, _ := json.Marshal(testJson)
					req, _ := http.NewRequest("POST", "/short", bytes.NewReader(body))
					r.ServeHTTP(w, req)

					assert.Equal(t, w.Code, http.StatusOK)
				}
				w := httptest.NewRecorder()
				req, _ := http.NewRequest("GET", "/metrics", nil)
				r.ServeHTTP(w, req)
				assert.Equal(t, w.Body.String(), "youtube.com: 4\nlinkedin.com: 2\ngithub.com: 1\n")
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, tc.run)
	}
}
