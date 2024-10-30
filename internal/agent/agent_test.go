package agent

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/domain/repositories"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	as "github.com/mihailtudos/metrickit/internal/service/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgent(t *testing.T) {
	t.Setenv("RATE_LIMIT", "2")
	agentCfg, err := config.NewAgentConfig()
	require.NoError(t, err)

	logger := slog.Default()
	metricsStore := storage.NewMetricsCollection()
	metricsRepo := repositories.NewAgentRepository(metricsStore, logger)
	agentService := as.NewAgentService(metricsRepo, agentCfg.Log, &agentCfg.Key)

	err = agentService.MetricsService.Collect()
	require.NoError(t, err)

	metrics, err := metricsRepo.GetAll()
	require.NoError(t, err)

	v, ok := metrics.CounterMetrics["PollCount"]
	assert.True(t, ok)
	assert.Equal(t, entities.Counter(1), v)

	// assert.Equal(t, 1, metrics.CounterMetrics[""])
	fmt.Printf("%v\n", metrics.CounterMetrics)
	assert.Less(t, entities.Gauge(0), metrics.GaugeMetrics["Sys"])
	assert.Less(t, entities.Gauge(0), metrics.GaugeMetrics["Mallocs"])
	assert.Less(t, entities.Gauge(0), metrics.GaugeMetrics["HeapObjects"])
	// assert.Less(t, entities.Gauge(0), metrics.GaugeMetrics["CPUutilization1"])

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request details if needed
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Simulate a successful response
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Metrics received"))
	}))
	defer testServer.Close()

	err = agentService.MetricsService.Send(testServer.URL[len("http://"):])
	require.NoError(t, err)
}
