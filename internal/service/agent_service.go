package service

import (
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/repositories"
)

type MetricsService interface {
	Collect()
	Send(serverAddr string)
	Clear()
}

type AgentService struct {
	MetricsService
}

func NewAgentService(repository *repositories.AgentRepository, logger *slog.Logger) *AgentService {
	return &AgentService{
		MetricsService: NewMetricsCollectionService(repository, logger),
	}
}
