package agent

import (
	"context"
	"log/slog"

	"github.com/mihailtudos/metrickit/pkg/helpers"
)

type SendMetricsTask struct {
	Service    *AgentService
	Log        *slog.Logger
	ServerAddr string
}

func (t *SendMetricsTask) Process() {
	if err := t.Service.MetricsService.Send(t.ServerAddr); err != nil {
		t.Log.ErrorContext(context.Background(),
			"failed to process send task",
			helpers.ErrAttr(err))
	}
}
