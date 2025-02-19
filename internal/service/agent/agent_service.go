// Package agent provides the agent service for collecting and sending metrics.
// It defines the MetricsService interface and implements it through the AgentService struct.
package agent

import (
	"crypto/rsa"
	"google.golang.org/grpc"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

// MetricsService defines the interface for metrics collection and transmission.
// It includes methods for collecting metrics and sending them to a server.
type MetricsService interface {
	// Collect gathers metrics data from the source.
	Collect() error

	// Send transmits the collected metrics to the specified server address.
	Send(serverAddr string) error
}

// AgentService implements the MetricsService interface.
// It is responsible for collecting metrics and sending them to a server.
type AgentService struct {
	MetricsService MetricsService // The metrics service used for collecting and sending metrics.
}

// NewAgentService creates a new instance of the AgentService struct.
// It initializes the agent service with the provided repository, logger, and secret.
func NewAgentService(repository *repositories.AgentRepository,
	logger *slog.Logger, secret *string,
	publicKey *rsa.PublicKey, gRPCConn *grpc.ClientConn) *AgentService {
	return &AgentService{
		MetricsService: NewMetricsCollectionService(repository,
			logger, secret, publicKey, gRPCConn), // Initialize the metrics collection service.
	}
}
