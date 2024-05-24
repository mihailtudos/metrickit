package server

import (
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"

	"github.com/stretchr/testify/require"
)

func TestCounterService(t *testing.T) {
	memStore := storage.NewMemStorage()
	repos := repositories.NewRepository(memStore)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cs := NewCounterService(repos.CounterRepository, logger)

	tests := []struct {
		name  string
		err   error
		value int64
		key   string
		want  int64
	}{
		{
			name:  "value numeric value",
			err:   nil,
			value: 222,
			key:   "PollCount",
			want:  222,
		},
		{
			name:  "negative value",
			err:   nil,
			value: -1,
			key:   "PollCount",
			want:  221,
		},
		{
			name:  "valid 0 value",
			err:   nil,
			value: 0,
			key:   "PollCount",
			want:  221,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := cs.Create(entities.Metrics{ID: test.key, MType: string(entities.CounterMetricName), Delta: &test.value})
			require.ErrorIs(t, err, test.err)
			if err == nil {
				v, ok := memStore.Counter[entities.MetricName(test.key)]
				assert.True(t, ok)
				assert.Equal(t, v, entities.Counter(test.want))
			} else {
				_, ok := memStore.Counter[entities.MetricName(test.key)]
				assert.False(t, ok)
			}
		})
	}
}
