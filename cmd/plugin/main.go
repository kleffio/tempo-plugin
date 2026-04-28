// Command plugin is the entrypoint for the traces-tempo Kleff plugin.
// It receives workload trace spans via gRPC and writes them to Tempo.
package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	pluginsv1 "github.com/kleffio/plugin-sdk-go/v1"
	grpcadapter "github.com/kleffio/tempo-plugin/internal/adapters/grpc"
	tempoadapter "github.com/kleffio/tempo-plugin/internal/adapters/tempo"
	"github.com/kleffio/tempo-plugin/internal/application"
	"google.golang.org/grpc"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	tempoURL := env("TEMPO_URL", "http://tempo:9411")
	tempoClient := tempoadapter.New(tempoURL)
	logger.Info("tempo backend", "url", tempoURL)

	svc := application.New(tempoClient)
	srv := grpcadapter.New(svc)

	gs := grpc.NewServer()
	pluginsv1.RegisterPluginHealthServer(gs, srv)
	pluginsv1.RegisterMonitoringTracesServer(gs, srv)

	port := env("PLUGIN_PORT", "50051")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("listen failed", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("plugin gRPC listening", "port", port)
		if err := gs.Serve(lis); err != nil {
			logger.Error("gRPC server error", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop
	logger.Info("shutting down")
	gs.GracefulStop()
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
