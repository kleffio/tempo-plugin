// Package tempo implements SpanStore backed by Grafana Tempo.
// Spans are written using the Zipkin v2 JSON push API:
// POST /api/v2/spans (Tempo Zipkin receiver on port 9411)
package tempo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	pluginsv1 "github.com/kleffio/plugin-sdk-go/v1"
)

// Client writes spans to a Tempo instance via the Zipkin push API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a Client that sends spans to the Tempo Zipkin endpoint at baseURL.
func New(baseURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// zipkinSpan is the Zipkin v2 span JSON schema subset that Tempo accepts.
type zipkinSpan struct {
	TraceID   string            `json:"traceId"`
	ID        string            `json:"id"`
	ParentID  string            `json:"parentId,omitempty"`
	Name      string            `json:"name"`
	Timestamp int64             `json:"timestamp"` // epoch microseconds
	Duration  int64             `json:"duration"`  // microseconds
	Tags      map[string]string `json:"tags,omitempty"`
}

// Ingest encodes the span in Zipkin v2 format and POSTs it to Tempo.
func (c *Client) Ingest(ctx context.Context, span *pluginsv1.Span) error {
	durationUs := (span.EndTimeNs - span.StartTimeNs) / 1000
	if durationUs < 0 {
		durationUs = 0
	}

	tags := map[string]string{
		"workload_id": span.WorkloadID,
		"org_id":      span.OrgID,
		"project_id":  span.ProjectID,
	}
	if span.Status != "" {
		tags["status"] = span.Status
	}
	for k, v := range span.Attributes {
		tags[k] = v
	}

	zs := zipkinSpan{
		TraceID:   span.TraceID,
		ID:        span.SpanID,
		ParentID:  span.ParentSpanID,
		Name:      span.Name,
		Timestamp: span.StartTimeNs / 1000,
		Duration:  durationUs,
		Tags:      tags,
	}

	body, err := json.Marshal([]zipkinSpan{zs})
	if err != nil {
		return fmt.Errorf("tempo: marshal span: %w", err)
	}

	url := c.baseURL + "/api/v2/spans"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("tempo: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("tempo: post span: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("tempo: unexpected status %d", resp.StatusCode)
	}
	return nil
}
