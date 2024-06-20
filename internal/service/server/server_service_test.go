package server

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/logger"
)

func TestCounterService(t *testing.T) {
	cfg, err := config.NewServerConfig()
	if err != nil {
		log.Fatal("failed to get configs: ", err.Error())
	}

	l, err := logger.NewLogger(os.Stdout, cfg.Envs.LogLevel)
	if err != nil {
		log.Fatal("failed to crate new log: ", err.Error())
	}

	ctx := context.Background()
	store, err := storage.NewStorage(nil, l, cfg.Envs.StoreInterval, cfg.Envs.StorePath)
	if err != nil {
		log.Fatal("failed to initiate storage: " + err.Error())
	}
	_ = store.Close(ctx)
	tests := []struct {
		name  string
		err   error
		value entities.Counter
		key   entities.MetricName
		want  entities.Counter
	}{
		{
			name:  "value numeric value",
			err:   nil,
			value: 222,
			key:   entities.PollCount,
			want:  222,
		},
		{
			name:  "negative value",
			err:   nil,
			value: -1,
			key:   entities.PollCount,
			want:  221,
		},
		{
			name:  "valid 0 value",
			err:   nil,
			value: 0,
			key:   entities.PollCount,
			want:  221,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}
