package service

import (
	"fmt"
	"strconv"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

type mockCounterService struct{}

func (m *mockCounterService) Create(key string, val string) error {
	if _, err := strconv.Atoi(val); err != nil {
		return fmt.Errorf("unable parse val=%v to int, erroor: %w", val, err)
	}

	return nil
}

func (m *mockCounterService) Get(key string) (entities.Counter, bool) {
	const testValue = 2
	return entities.Counter(testValue), true
}

func (m *mockCounterService) GetAll() map[string]entities.Counter {
	return make(map[string]entities.Counter)
}

type mockGaugeService struct{}

func (m *mockGaugeService) Create(key string, val string) error {
	if _, err := strconv.ParseFloat(val, 64); err != nil {
		return fmt.Errorf("unable parse val=%v to float64, erroor: %w", val, err)
	}

	return nil
}

func (m *mockGaugeService) Get(key string) (entities.Gauge, bool) {
	const testValue = 2.2
	return entities.Gauge(testValue), true
}

func (m *mockGaugeService) GetAll() map[string]entities.Gauge {
	return make(map[string]entities.Gauge)
}

func NewMockService() *Service {
	return &Service{
		GaugeService:   &mockGaugeService{},
		CounterService: &mockCounterService{},
	}
}
