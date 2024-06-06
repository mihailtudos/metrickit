package agent

import (
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

type MetricsService interface {
	Collect() error
	Send(serverAddr string) error
}

type AgentService struct {
	MetricsService MetricsService
}

func NewAgentService(repository *repositories.AgentRepository, logger *slog.Logger) *AgentService {
	return &AgentService{
		MetricsService: NewMetricsCollectionService(repository, logger),
	}
}
