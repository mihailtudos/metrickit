package server

import (
	"log/slog"
	"os"
	"testing"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCounterService(t *testing.T) {
	memStore := storage.NewMemStorage()
	repos := repositories.NewRepository(memStore)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cs := NewCounterService(repos.CounterRepository, logger)

	tests := []struct {
		name  string
		want  error
		value string
		key   entities.MetricName
	}{
		{
			name:  "value is a valid numeric floating point value",
			want:  ErrInvalidValue,
			value: "222.213",
			key:   "metric0",
		},
		{
			name:  "value is a valid numeric value",
			want:  nil,
			value: "222",
			key:   "metric1",
		},
		{
			name:  "value is a non-numerical string",
			want:  ErrInvalidValue,
			value: "sada",
			key:   "metric2",
		},
		{
			name:  "value is a mix of numeric and alphabetical letters",
			want:  ErrInvalidValue,
			value: "12invalid",
			key:   "metric3",
		},
		{
			name:  "value an empty string",
			want:  ErrInvalidValue,
			value: "",
			key:   "metric4",
		},
		{
			name:  "value is a floating point value delimited with comma",
			want:  ErrInvalidValue,
			value: "222,21",
			key:   "metric5",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := cs.Create(test.key, test.value)
			require.ErrorIs(t, err, test.want)
			if err == nil {
				assert.Contains(t, memStore.Counter, test.key)
			} else {
				assert.NotContains(t, memStore.Counter, test.key)
			}
		})
	}
}
