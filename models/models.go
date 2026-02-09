package models

import "time"

// Header represents a key-value pair for HTTP headers
type Header struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// Request represents an HTTP request configuration
type Request struct {
	Method  string   `json:"method"`
	URL     string   `json:"url"`
	Headers []Header `json:"headers"`
	Body    string   `json:"body"`
}

// Response represents an HTTP response
type Response struct {
	StatusCode   int               `json:"status_code"`
	Status       string            `json:"status"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	ResponseTime time.Duration     `json:"response_time"`
	Error        string            `json:"error,omitempty"`
}

// Template represents a saved request template
type Template struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Request   Request   `json:"request"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// HistoryItem represents a request history entry
type HistoryItem struct {
	ID        string    `json:"id"`
	Request   Request   `json:"request"`
	Response  Response  `json:"response"`
	Timestamp time.Time `json:"timestamp"`
}

// NewRequest creates a new request with default values
func NewRequest() *Request {
	return &Request{
		Method:  "GET",
		URL:     "",
		Headers: []Header{},
		Body:    "",
	}
}

// Clone creates a copy of the request
func (r *Request) Clone() *Request {
	headers := make([]Header, len(r.Headers))
	copy(headers, r.Headers)
	return &Request{
		Method:  r.Method,
		URL:     r.URL,
		Headers: headers,
		Body:    r.Body,
	}
}
