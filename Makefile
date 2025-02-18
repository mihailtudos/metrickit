VERSION?=1.0.0
COMMIT=$(if $(shell git rev-parse HEAD),$(shell git rev-parse HEAD),"N/A")
DATE=$(shell date "+%Y/%m/%d %H:%M:%S")
ARGS := "-a=localhost:8080"

mock:
	mockgen -destination=internal/mocks/mock_system_service.go -package=mocks metrics/internal/core/service Pinger	
	mockgen -destination=internal/mocks/mock_db_store.go -package=mocks metrics/internal/core/service Store	

docs:
	godoc -http=:8000 -goroot=$(shell pwd)

docs/gen:
	wget -r -np -nH -N -E -p -P ./docs -k http://localhost:8080/pkg/github.com/mihailtudos/metrickit/

docs/show:
	godoc -goroot="." -http=:8080

swag:
	swag init -g ./cmd/server/main.go --output ./docs/swagger

swag/gen:
	swag init --generalInfo ./cmd/server/main.go --parseInternal   --output ./swagger/
	
run/server:
	go run ./cmd/server/. \
		-crypto-key="./private.pem" \
		-d="postgres://metrics:metrics@localhost:5432/metrics?sslmode=disable" $(ARGS)

run/agent:
	go run ./cmd/agent/. $(ARGS) -crypto-key=public.pem

run/tests:
	#go test -v -coverpkg=./... -coverprofile=profile.cov ./...
	go test ./... -count=1 -coverprofile ./profiles/cover.out && go tool cover -func ./profiles/cover.out

show/cover:
	go tool cover -html=./profiles/cover.out

run/pprof-snap:
	curl http://localhost:8080/debug/pprof/profile\?seconds=30 -o ./profiles/result.pprof

show/pprof-base:
	go tool pprof -http=":9090" ./profiles/base.pprof

show/pprof-res:
	go tool pprof -http=":9090" ./profiles/result.pprof

show/pprof-diff:
	pprof -top -diff_base=profiles/base.pprof profiles/result.pprof


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

staticlint/build:
	cd cmd/staticlint && go build -o staticlint && mv staticlint ../../staticlint
	cd ../..

staticlint/run: staticlint/build
	./staticlint ./...

gen/metric-proto:
	 protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           proto/metrics/metrics.proto

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

autotest/run18:
		shortenertestbeta -test.v -test.run="^TestIteration18$$" \
                      -source-path=. \

.PHONY: run/server, run/agent, run/tests, show/cover, gci/report, \
		autotest/run1, autotest/run2, autotest/run3, \
		autotest/run4, autotest/run5, autotest/run6, \
		autotest/run7, autotest/run8, autotest/run9, \
		autotest/run10, autotest/run11, autotest/run12, \
		autotest/run13, db/run, autotest/run18, gen/metric-proto

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
    golangci/golangci-lint:v1.61.0 \
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
