ARGS := "-a=localhost:8080"

run/server:
	go run ./cmd/server/. $(ARGS)

run/agent:
	go run ./cmd/agent/. $(ARGS)

run/tests:
	go test -v -coverpkg=./... -coverprofile=profile.cov ./...

.PHONY: run/server, run/agent, run/tests