ARGS := "-a=localhost:8080"

run/server:
	go run ./cmd/server/. $(ARGS)

run/agent:
	go run ./cmd/agent/. $(ARGS)

run/tests:
	go test -v -coverpkg=./... -coverprofile=profile.cov ./...

show/cover:
	go tool cover -html=profile.cov

gci/report:
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

agent/build:
	cd cmd/agent && \
      go build -buildvcs=false  -o agent

server/build:
	cd cmd/server && \
      go build -buildvcs=false  -o server

autotest/run1: server/build
	metricstest -test.v -test.run="^TestIteration1$$" \
		-binary-path=cmd/server/server

autotest/run2: agent/build
	metricstest -test.v -test.run="^TestIteration2[AB]*$$" \
		-source-path=. \
 		-agent-binary-path=cmd/agent/agent

autotest/run3:
	metricstest -test.v -test.run="^TestIteration3[AB]*$$" \
		 -source-path=. \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server

autotest/run4:
	ADDRESS=localhost:8080 TEMP_FILE=out.txt metricstest -test.v -test.run="^TestIteration4$$" \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
        -server-port=8080 \
        -source-path=. \
        -agent-binary-path=cmd/agent/agent \


autotest/run5:
	ADDRESS=localhost:8080 TEMP_FILE=out.txt metricstest -test.v -test.run="^TestIteration5$$" \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
        -server-port=8080 \
        -source-path=. \
        -agent-binary-path=cmd/agent/agent \

autotest/run6:
	ADDRESS=localhost:8080 TEMP_FILE=out.txt metricstest -test.v -test.run="^TestIteration6$$" \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
        -server-port=8080 \
        -source-path=. \
        -agent-binary-path=cmd/agent/agent \

autotest/run7:
	TEMP_FILE=out.txt metricstest -test.v -test.run=^TestIteration7 \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
        -server-port=8080 \
        -source-path=. \
        -agent-binary-path=cmd/agent/agent \

autotest/run8:
	TEMP_FILE=out.txt metricstest -test.v -test.run=^TestIteration8$ \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
        -server-port=8080 \
        -source-path=.-agent-binary-path=cmd/agent/agent \

.PHONY: run/server, run/agent, run/tests, show/cover, gci/report, autotest/run1, autotest/run8, autotest/run7

GOLANGCI_LINT_CACHE?=/tmp/praktikum-golangci-lint-cache

.PHONY: golangci-lint-run
golangci-lint-run: _golangci-lint-rm-unformatted-report

.PHONY: _golangci-lint-reports-mkdir
_golangci-lint-reports-mkdir:
	mkdir -p ./golangci-lint

.PHONY: _golangci-lint-run
_golangci-lint-run: _golangci-lint-reports-mkdir
	-docker run --rm \
    -v $(shell pwd):/app \
    -v $(GOLANGCI_LINT_CACHE):/root/.cache \
    -w /app \
    golangci/golangci-lint:v1.57.2 \
        golangci-lint run \
            -c .golangci.yml \
	> ./golangci-lint/report-unformatted.json

.PHONY: _golangci-lint-format-report
_golangci-lint-format-report: _golangci-lint-run
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

.PHONY: _golangci-lint-rm-unformatted-report
_golangci-lint-rm-unformatted-report: _golangci-lint-format-report
	rm ./golangci-lint/report-unformatted.json

.PHONY: golangci-lint-clean
golangci-lint-clean:
	sudo rm -rf ./golangci-lint
