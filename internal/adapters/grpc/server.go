// Package grpc is the inbound gRPC adapter for the traces-tempo plugin.
package grpc

import (
	"context"

	pluginsv1 "github.com/kleffio/plugin-sdk-go/v1"
	"github.com/kleffio/tempo-plugin/internal/application"
)

// Server implements PluginHealth and MonitoringTraces gRPC services.
type Server struct {
	pluginsv1.UnimplementedPluginHealthServer
	pluginsv1.UnimplementedMonitoringTracesServer
	svc *application.Service
}

// New creates a Server wired to the given application service.
func New(svc *application.Service) *Server {
	return &Server{svc: svc}
}

// Health reports the plugin as healthy.
func (s *Server) Health(_ context.Context, _ *pluginsv1.HealthRequest) (*pluginsv1.HealthResponse, error) {
	return &pluginsv1.HealthResponse{Status: pluginsv1.HealthStatusHealthy}, nil
}

// GetCapabilities declares the monitoring.traces capability.
func (s *Server) GetCapabilities(_ context.Context, _ *pluginsv1.GetCapabilitiesRequest) (*pluginsv1.GetCapabilitiesResponse, error) {
	return &pluginsv1.GetCapabilitiesResponse{
		Capabilities: []string{pluginsv1.CapabilityMonitoringTraces},
	}, nil
}

// IngestSpan receives a trace span from the platform and writes it to Tempo.
func (s *Server) IngestSpan(ctx context.Context, req *pluginsv1.IngestSpanRequest) (*pluginsv1.IngestSpanResponse, error) {
	if err := s.svc.IngestSpan(ctx, req.Span); err != nil {
		return &pluginsv1.IngestSpanResponse{
			Error: &pluginsv1.PluginError{
				Code:    pluginsv1.ErrorCodeInternal,
				Message: err.Error(),
			},
		}, nil
	}
	return &pluginsv1.IngestSpanResponse{}, nil
}
