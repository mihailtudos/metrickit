// Package agent provides functionality for sending metrics to a specified server.
// It defines the SendMetricsTask struct, which encapsulates the logic for
// processing the metric sending task.
package agent

import (
	"context"
	"log/slog"

	"github.com/mihailtudos/metrickit/pkg/helpers"
)

// SendMetricsTask represents a task for sending metrics to a server.
// It contains the necessary dependencies such as the AgentService,
// a logger for logging messages, and the server address to which
// the metrics will be sent.
type SendMetricsTask struct {
	Service    *AgentService // The AgentService that provides the metrics service
	Log        *slog.Logger  // Logger for logging errors and messages
	ServerAddr string        // The address of the server to send metrics to
}

// Process executes the task of sending metrics to the specified server address.
// If an error occurs during the sending process, it logs the error using the
// provided logger.
func (t *SendMetricsTask) Process() {
	if err := t.Service.MetricsService.Send(t.ServerAddr); err != nil {
		t.Log.ErrorContext(context.Background(),
			"failed to process send task",
			helpers.ErrAttr(err)) // Logs the error with additional attributes
	}
}
