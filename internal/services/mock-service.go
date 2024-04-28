package services

import (
	"strconv"
)

type mockCounterService struct{}

func (m *mockCounterService) Create(key string, val string) error {
	if _, err := strconv.Atoi(val); err != nil {
		return err
	}

	return nil
}

type mockGaugeService struct{}

func (m *mockGaugeService) Create(key string, val string) error {
	if _, err := strconv.ParseFloat(val, 64); err != nil {
		return err
	}

	return nil
}

func NewMockService() *Service {
	return &Service{
		GaugeService:   &mockCounterService{},
		CounterService: &mockCounterService{},
	}
}
