package agent

type SendMetricsTask struct {
	Service    *AgentService
	ServerAddr string
}

func (t *SendMetricsTask) Process() {
	_ = t.Service.MetricsService.Send(t.ServerAddr)
}
