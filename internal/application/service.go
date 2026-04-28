// Package application holds the core business logic for the traces-tempo plugin.
package application

import (
	"context"

	pluginsv1 "github.com/kleffio/plugin-sdk-go/v1"
)

// SpanStore is implemented by the Tempo adapter.
type SpanStore interface {
	Ingest(ctx context.Context, span *pluginsv1.Span) error
}

// Service orchestrates span ingestion.
type Service struct {
	store SpanStore
}

// New creates a Service backed by the given SpanStore.
func New(store SpanStore) *Service {
	return &Service{store: store}
}

// IngestSpan forwards a trace span to the backing store.
func (s *Service) IngestSpan(ctx context.Context, span *pluginsv1.Span) error {
	if span == nil {
		return nil
	}
	return s.store.Ingest(ctx, span)
}
