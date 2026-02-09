package http

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"percentman/models"
)

// Client handles HTTP requests
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new HTTP client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendRequest sends an HTTP request and returns the response
func (c *Client) SendRequest(req *models.Request) *models.Response {
	response := &models.Response{}

	// Validate URL
	if req.URL == "" {
		response.Error = "URL is required"
		return response
	}

	// Add http:// if no protocol specified
	url := req.URL
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "http://" + url
	}

	// Create request body
	var body io.Reader
	if req.Body != "" {
		body = bytes.NewBufferString(req.Body)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, url, body)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	// Add headers
	for _, h := range req.Headers {
		if h.Enabled && h.Key != "" {
			httpReq.Header.Set(h.Key, h.Value)
		}
	}

	// Set default Content-Type for requests with body
	if req.Body != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Send request and measure time
	startTime := time.Now()
	httpResp, err := c.httpClient.Do(httpReq)
	response.ResponseTime = time.Since(startTime)

	if err != nil {
		response.Error = err.Error()
		return response
	}
	defer httpResp.Body.Close()

	// Read response
	response.StatusCode = httpResp.StatusCode
	response.Status = httpResp.Status

	// Copy response headers
	response.Headers = make(map[string]string)
	for k, v := range httpResp.Header {
		if len(v) > 0 {
			response.Headers[k] = strings.Join(v, ", ")
		}
	}

	// Read response body
	bodyBytes, err := io.ReadAll(httpResp.Body)
	if err != nil {
		response.Error = "Failed to read response body: " + err.Error()
		return response
	}

	response.Body = string(bodyBytes)

	return response
}

// FormatJSON formats a JSON string with indentation
func FormatJSON(input string) string {
	if input == "" {
		return ""
	}

	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		// Not valid JSON, return as-is
		return input
	}

	formatted, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return input
	}

	return string(formatted)
}

// IsJSON checks if a string is valid JSON
func IsJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
