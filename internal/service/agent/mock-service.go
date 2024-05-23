package agent

import (
	"fmt"
	"strconv"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/service/server"
)

type mockCounterService struct{}

func (m *mockCounterService) Create(key entities.MetricName, val string) error {
	if _, err := strconv.Atoi(val); err != nil {
		return fmt.Errorf("unable parse val=%v to int, erroor: %w", val, err)
	}

	return nil
}

func (m *mockCounterService) Get(key entities.MetricName) (entities.Counter, error) {
	const testValue = 2
	return entities.Counter(testValue), nil
}

func (m *mockCounterService) GetAll() (map[entities.MetricName]entities.Counter, error) {
	return map[entities.MetricName]entities.Counter{}, nil
}

type mockGaugeService struct{}

func (m *mockGaugeService) Create(key entities.MetricName, val string) error {
	if _, err := strconv.ParseFloat(val, 64); err != nil {
		return fmt.Errorf("unable parse val=%v to float64, erroor: %w", val, err)
	}

	return nil
}

func (m *mockGaugeService) Get(key entities.MetricName) (entities.Gauge, error) {
	const testValue = 2.2
	return entities.Gauge(testValue), nil
}

func (m *mockGaugeService) GetAll() (map[entities.MetricName]entities.Gauge, error) {
	return map[entities.MetricName]entities.Gauge{}, nil
}

func NewMockService() *server.Service {
	return &server.Service{
		GaugeService:   &mockGaugeService{},
		CounterService: &mockCounterService{},
	}
}
