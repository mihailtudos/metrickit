package handlers

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"text/template"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
	"github.com/mihailtudos/metrickit/pkg/helpers"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/mihailtudos/metrickit/swagger"
	httpSwagger "github.com/swaggo/http-swagger"
)

// ErrUnknownMetric is an error indicating that an unknown metric type was encountered.
var ErrUnknownMetric = errors.New("unknown metric type")

const bodyKey = "body"

//go:embed templates
var templatesFs embed.FS

// ServerHandler is a struct that encapsulates the services, logger,
// database connection, and template filesystem for handling HTTP requests.
type ServerHandler struct {
	services    *server.Service
	logger      *slog.Logger
	TemplatesFs embed.FS
	db          *pgxpool.Pool
	secret      string
}

// NewHandler initializes a new ServerHandler and registers the application routes.
// It takes services, logger, database connection, and a secret key as parameters.
func NewHandler(services *server.Service, logger *slog.Logger,
	conn *pgxpool.Pool, secret string) http.Handler {
	handlers := &ServerHandler{
		services:    services,
		logger:      logger,
		TemplatesFs: templatesFs,
		db:          conn,
		secret:      secret,
	}

	return handlers.registerRoutes()
}

// registerRoutes sets up the HTTP routes for the application.
// It returns an http.Handler with the configured routes.
func (sh *ServerHandler) registerRoutes() http.Handler {
	mux := chi.NewMux()

	mux.Use(
		RequestLogger(sh.logger),
		WithCompressedResponse(sh.logger),
	)

	// GET http://<SERVER_ADDRESS>/value/<METRIC_TYPE>/<METRIC_NAME>
	// Content-Type: text/plain
	mux.Get("/value/{metricType}/{metricName}", sh.getMetricValue)
	mux.Get("/", sh.showMetrics)

	// handlers to handle metrics following the format:
	// http://<SERVER_ADR>/update/<METRIC_TYPE>/<METRIC_NAME>/<METRIC_VALUE>
	// Content-Type: text/plain
	mux.Post("/update/{metricType}/{metricName}/{metricValue}", sh.handleUploads)

	mux.Post("/update/", sh.handleJSONUploads)
	mux.Post("/updates/", sh.handleBatchUploads)
	mux.Post("/value/", sh.getJSONMetricValue)

	mux.Get("/ping", sh.handleDBPing)

	// Register pprof handlers
	mux.Get("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Get("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Get("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Get("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Get("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	mux.Handle("/debug/pprof/{profile}", http.HandlerFunc(pprof.Index))

	// Serve Swagger documentation and Swagger UI
	mux.Handle("/swagger/*", http.StripPrefix("/swagger/", http.FileServer(http.Dir("./swagger"))))
	mux.Get("/swagger-ui/*", httpSwagger.WrapHandler)

	return mux
}

// showMetrics displays collected metrics in an HTML format.
// It responds with an HTML page containing the metrics or an error if the retrieval fails.
// //nolint:godot // this comment is part of the Swagger documentation
// Show Metrics
// @Tags Metrics
// @Summary Show collected metrics
// @ID infoMetrics
// @Accept json
// @Produce text/html
// @Success 200 {string} string "HTML response with the collected metrics"
// @Failure 500 {string} string "Internal Server Error"
// @Router / [get]
func (sh *ServerHandler) showMetrics(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(sh.TemplatesFs, "templates/index.html")
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to parse the template: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	metrics, err := sh.services.MetricsService.GetAll()
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to get the metrics: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set(helpers.ContentType, "text/html; charset=utf-8")
	err = tmpl.ExecuteTemplate(w, "index.html", metrics)

	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to execute template: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

// getMetricValue retrieves the value of a specific metric type and name.
// It responds with the metric value or an error if the metric is not found or
// the type is unknown.
// //nolint:godot // this comment is part of the Swagger documentation
// Get Metric Value
// @Tags Metrics
// @Summary Retrieve a metric's value by type and name
// @ID getMetricValue
// @Accept  json
// @Produce text/plain
// @Param metricType path string true "Metric Type" Enum("counter", "gauge")
// @Param metricName path string true "Metric Name"
// @Success 200 {string} string "Metric value returned successfully"
// @Failure 400 {string} string "Bad Request - Unknown metric type"
// @Failure 404 {string} string "Not Found - Metric not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /value/{metricType}/{metricName} [get]
func (sh *ServerHandler) getMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")

	metric := entities.Metrics{ID: metricName, MType: metricType}
	currentMetric, err := sh.getMetric(metric)
	if err != nil {
		sh.logger.DebugContext(r.Context(),
			"failed to get the metric struct",
			helpers.ErrAttr(err))
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrUnknownMetric) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sh.logger.ErrorContext(r.Context(),
			"failed to get metric: ",
			helpers.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(helpers.ContentType, "text/plain; charset=utf-8")
	switch entities.MetricType(currentMetric.MType) {
	case entities.CounterMetricName:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%v", *currentMetric.Delta)
	case entities.GaugeMetricName:
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, "%v", *currentMetric.Value)

	default:
		sh.logger.ErrorContext(r.Context(),
			"failed identify the correct metric type")
		w.WriteHeader(http.StatusNotFound)
	}
}

// getJSONMetricValue retrieves a metric's value in JSON format.
// It responds with the metric data or an error if the retrieval fails.
// //nolint:godot // this comment is part of the Swagger documentation
// Get Metric Value in JSON
// @Tags Metrics
// @Summary Retrieve a metric's value in JSON format
// @ID getJSONMetricValue
// @Accept  json
// @Produce json
// @Param metric body entities.Metrics true "Metric Data"
// @Success 200 {object} entities.Metrics "Metric value returned successfully"
// @Failure 400 {string} string "Bad Request - Invalid metric data"
// @Failure 404 {string} string "Not Found - Metric not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /value/ [get]
func (sh *ServerHandler) getJSONMetricValue(w http.ResponseWriter, r *http.Request) {
	metric := entities.Metrics{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to read body: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &metric)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to marshal request body: ",
			helpers.ErrAttr(err),
			slog.String(bodyKey, string(body)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	currentMetric, err := sh.getMetric(metric)
	if err != nil {
		sh.logger.DebugContext(context.Background(),
			"failed to get the metric struct",
			helpers.ErrAttr(err))
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if errors.Is(err, ErrUnknownMetric) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sh.logger.ErrorContext(context.Background(),
			"failed to get metric: ",
			helpers.ErrAttr(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(helpers.ContentType, "application/json; charset=utf-8")
	jsonMetric, err := json.MarshalIndent(currentMetric, "", "  ")
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to marshal metric: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(jsonMetric); err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to write response: ",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// getMetric retrieves a metric from the storage.
// It returns the metric or an error if the retrieval fails.
// It returns an error if the metric type is unknown.
// //nolint:godot // this comment is part of the Swagger documentation
// Get Metric Value
// @Tags Metrics
// @Summary Retrieve a metric's value
// @ID getMetricValue
// @Accept  json
// @Produce json
// @Param metric body entities.Metrics true "Metric Data"
// @Success 200 {object} entities.Metrics "Metric value returned successfully"
// @Failure
// @Failure 400 {string} string "Bad Request - Invalid metric data"
// @Failure 404 {string} string "Not Found - Metric not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /value/ [get]
func (sh *ServerHandler) getMetric(metric entities.Metrics) (*entities.Metrics, error) {
	if entities.MetricType(metric.MType) == entities.CounterMetricName {
		record, err := sh.services.MetricsService.Get(entities.MetricName(metric.ID), entities.MetricType(metric.MType))
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("metric with type=%s, name=%s not found: %w", metric.MType, metric.MType, err)
			}

			return nil, fmt.Errorf("failed to get the given metric: %w", err)
		}

		return &record, nil
	}

	if entities.MetricType(metric.MType) == entities.GaugeMetricName {
		record, err := sh.services.MetricsService.Get(entities.MetricName(metric.ID), entities.MetricType(metric.MType))
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("metric with type=%s, name=%s not found: %w", metric.MType, metric.MType, err)
			}

			return nil, fmt.Errorf("failed to get the given metric: %w", err)
		}

		return &record, nil
	}

	return nil, ErrUnknownMetric
}

// handleDBPing handles the DB ping request.
// It responds with an error if the ping fails.
// //nolint:godot // this comment is part of the Swagger documentation
// Ping DB
// @Tags Metrics
// @Summary Ping the DB
// @ID pingDB
// @Accept  json
// @Produce json
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Internal Server Error"
// @Router /ping [get]
func (sh *ServerHandler) handleDBPing(w http.ResponseWriter, r *http.Request) {
	if err := sh.db.Ping(r.Context()); err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to ping the DB",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}
