package service

import (
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"strconv"
)

type mockCounterService struct{}

func (m *mockCounterService) Create(key string, val string) error {
	if _, err := strconv.Atoi(val); err != nil {
		return err
	}

	return nil
}

func (m *mockCounterService) Get(key string) (entities.Counter, bool) {
	return entities.Counter(2), true
}

func (m *mockCounterService) GetAll() map[string]entities.Counter {
	return make(map[string]entities.Counter)
}

type mockGaugeService struct{}

func (m *mockGaugeService) Create(key string, val string) error {
	if _, err := strconv.ParseFloat(val, 64); err != nil {
		return err
	}

	return nil
}

func (m *mockGaugeService) Get(key string) (entities.Gauge, bool) {
	return entities.Gauge(2.2), true
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
