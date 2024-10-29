package repositories

import (
	"log/slog"
	"testing"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentRepository(t *testing.T) {
	// Setup
	store := &storage.MetricsCollection{}
	logger := slog.Default()

	t.Run("successfully creates new agent repository", func(t *testing.T) {
		repo := NewAgentRepository(store, logger)
		assert.NotNil(t, repo)
		assert.NotNil(t, repo.MetricsCollectionRepository)
	})
}

func TestAgentRepository_Store(t *testing.T) {
	store := storage.NewMetricsCollection()
	logger := slog.Default()
	repo := NewAgentRepository(store, logger)

	tests := []struct {
		name    string
		metrics map[entities.MetricName]entities.Gauge
		wantErr bool
	}{
		//nolint:exhaustive // Ignoring exhaustive check, only testing a subset of metrics
		{
			name: "successfully stores gauge metrics",
			metrics: map[entities.MetricName]entities.Gauge{
				"test_metric": 123.45,
			},
			wantErr: false,
		},
		{
			name:    "successfully stores empty metrics",
			metrics: map[entities.MetricName]entities.Gauge{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Store(tt.metrics)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAgentRepository_GetAll(t *testing.T) {
	// Setup
	store := storage.NewMetricsCollection()
	logger := slog.Default()
	repo := NewAgentRepository(store, logger)

	t.Run("successfully retrieves all metrics", func(t *testing.T) {
		metrics, err := repo.GetAll()
		require.NoError(t, err)
		assert.NotNil(t, metrics)
	})
}
