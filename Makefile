ARGS := "-a=localhost:8080"

run/server:
	go run ./cmd/server/. ${ARGS}

run/agent:
	go run ./cmd/agent/. ${ARGS}


.PHONY: run/server, run/agent