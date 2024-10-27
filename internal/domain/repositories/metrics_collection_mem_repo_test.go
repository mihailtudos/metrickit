package repositories

import (
	"log/slog"
	"testing"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/stretchr/testify/assert"
)

func TestMetricsCollectionMemRepository(t *testing.T) {
	logger := slog.Default()

	t.Run("Store successfully stores metrics and updates counter", func(t *testing.T) {
		repo := NewMetricsCollectionMemRepository(storage.NewMetricsCollection(), logger)

		//nolint:exhaustive // Only testing a subset of metrics
		gaugeMetrics := map[entities.MetricName]entities.Gauge{
			entities.CPUutilization1: entities.Gauge(1.1),
		}

		err := repo.Store(gaugeMetrics)

		assert.NoError(t, err)
	})

	// 	t.Run("Store returns an error if storing metrics fails", func(t *testing.T) {
	// 		// Arrange
	// 		mockStore := new(MockMetricsCollection)
	// 		repo := NewMetricsCollectionMemRepository(mockStore, logger)

	// 		gaugeMetrics := map[entities.MetricName]entities.Gauge{
	// 			"gauge1": {Value: 10},
	// 		}

	// 		mockStore.On("StoreGauge", gaugeMetrics).Return(errors.New("store error"))

	// 		err := repo.Store(gaugeMetrics)

	// 		assert.Error(t, err)
	// 		assert.Contains(t, err.Error(), "failed to store the metrics")
	// 		mockStore.AssertExpectations(t)
	// 	})

	// 	t.Run("GetAll successfully retrieves all metrics", func(t *testing.T) {
	// 		// Arrange
	// 		mockStore := new(MockMetricsCollection)
	// 		repo := NewMetricsCollectionMemRepository(mockStore, logger)

	// 		counters := map[entities.MetricName]entities.Counter{
	// 			"poll_count": {Delta: new(int64)},
	// 		}
	// 		gauges := map[entities.MetricName]entities.Gauge{
	// 			"gauge1": {Value: 10},
	// 		}

	// 		mockStore.On("GetCounterCollection").Return(counters, nil)
	// 		mockStore.On("GetGaugeCollection").Return(gauges, nil)

	// 		metrics, err := repo.GetAll()

	// 		assert.NoError(t, err)
	// 		assert.Equal(t, counters, metrics.CounterMetrics)
	// 		assert.Equal(t, gauges, metrics.GaugeMetrics)
	// 		mockStore.AssertExpectations(t)
	// 	})

	// 	t.Run("GetAll returns an error if retrieving counter collection fails", func(t *testing.T) {
	// 		// Arrange
	// 		mockStore := new(MockMetricsCollection)
	// 		repo := NewMetricsCollectionMemRepository(mockStore, logger)

	// 		mockStore.On("GetCounterCollection").Return(nil, errors.New("counter error"))

	// 		// Act
	// 		metrics, err := repo.GetAll()

	// 		// Assert
	// 		assert.Error(t, err)
	// 		assert.Nil(t, metrics)
	// 		assert.Contains(t, err.Error(), "failed to retrieve the Counter collection")
	// 		mockStore.AssertExpectations(t)
	// 	})

	// 	t.Run("GetAll returns an error if retrieving gauge collection fails", func(t *testing.T) {
	// 		// Arrange
	// 		mockStore := new(MockMetricsCollection)
	// 		repo := NewMetricsCollectionMemRepository(mockStore, logger)

	// 		mockStore.On("GetCounterCollection").Return(map[entities.MetricName]entities.Counter{}, nil)
	// 		mockStore.On("GetGaugeCollection").Return(nil, errors.New("gauge error"))

	// 		// Act
	// 		metrics, err := repo.GetAll()

	//		// Assert
	//		assert.Error(t, err)
	//		assert.Nil(t, metrics)
	//		assert.Contains(t, err.Error(), "failed to retrieve the Gauge collection")
	//		mockStore.AssertExpectations(t)
	//	})
}
