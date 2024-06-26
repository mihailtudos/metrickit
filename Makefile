ARGS := "-a=localhost:8080"

run/server:
	go run ./cmd/server/. -d="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" $(ARGS)

run/agent:
	go run ./cmd/agent/. $(ARGS)

run/tests:
	go test -v -coverpkg=./... -coverprofile=profile.cov ./...

show/cover:
	go tool cover -html=profile.cov

gci/report:
	cat ./golangci-lint/report-unformatted.json | jq > ./golangci-lint/report.json

db/run:
	docker-compose up -d db

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

autotest/run9:
	TEMP_FILE=out.txt metricstest -test.v -test.run="^TestIteration9$$" \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
        -server-port=8080 \
        -source-path=. \
        -file-storage-path=/tmp/metrics-db.json \
        -agent-binary-path=cmd/agent/agent \

autotest/run10: db/run
	 SERVER_PORT=8080 TEMP_FILE=out.txt metricstest -test.v -test.run="^TestIteration10[AB]$$" \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
        -server-port=8080 \
        -source-path=.

autotest/run11: db/run
	 SERVER_PORT=8080 TEMP_FILE=out.txt metricstest -test.v -test.run="^TestIteration11$$" \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
        -server-port=8080 \
        -source-path=.

autotest/run12: db/run
	 SERVER_PORT=8080 TEMP_FILE=out.txt metricstest -test.v -test.run="^TestIteration12$$" \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
        -server-port=8080 \
        -source-path=.

autotest/run13: db/run
	 SERVER_PORT=8080 TEMP_FILE=out.txt metricstest -test.v -test.run="^TestIteration13$$" \
		-agent-binary-path=cmd/agent/agent \
		-binary-path=cmd/server/server \
		-database-dsn='postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable' \
        -server-port=8080 \
        -source-path=.

.PHONY: run/server, run/agent, run/tests, show/cover, gci/report, \
		autotest/run1, autotest/run2, autotest/run3, \
		autotest/run4, autotest/run5, autotest/run6, \
		autotest/run7, autotest/run8, autotest/run9, \
		autotest/run10, autotest/run11, autotest/run12, \
		autotest/run13, db/run

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
