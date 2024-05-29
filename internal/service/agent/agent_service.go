package agent

import (
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

// TODO(SSH): where do you use this interface ? Why do you need it ?
type MetricsService interface {
	Collect() error
	SendJSONMetric(serverAddr string) error
}

type AgentService struct {
	MetricsService
}

func NewAgentService(repository *repositories.AgentRepository, logger *slog.Logger) *AgentService {
	return &AgentService{
		MetricsService: NewMetricsCollectionService(repository, logger),
	}
}
