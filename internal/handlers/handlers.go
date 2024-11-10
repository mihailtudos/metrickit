package handlers

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"path"
	"strconv"
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
	privateKey  *rsa.PrivateKey
	services    server.Metrics
	logger      *slog.Logger
	TemplatesFs embed.FS
	db          *pgxpool.Pool
	secret      string
}

// NewHandler initializes a new ServerHandler and registers the application routes.
// It takes services, logger, database connection, and a secret key as parameters.
func NewHandler(services server.Metrics, logger *slog.Logger,
	conn *pgxpool.Pool, secret string, privateKey *rsa.PrivateKey) *ServerHandler {
	return &ServerHandler{
		services:    services,
		logger:      logger,
		TemplatesFs: templatesFs,
		db:          conn,
		secret:      secret,
		privateKey:  privateKey,
	}
}

// Router sets up the HTTP routes for the application.
// It returns an http.Handler with the configured routes.
func Router(logger *slog.Logger, sh *ServerHandler) http.Handler {
	mux := chi.NewMux()

	mux.Use(
		RequestLogger(logger),
		WithCompressedResponse(logger),
	)

	// GET http://<SERVER_ADDRESS>/value/<METRIC_TYPE>/<METRIC_NAME>
	// Content-Type: text/plain
	mux.Get("/value/{metricType}/{metricName}", sh.getMetricValue)
	mux.Get("/", sh.showMetrics(""))

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
func (sh *ServerHandler) showMetrics(templatePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmplPath := path.Join("templates", "index.html")

		if templatePath != "" {
			tmplPath = templatePath
		}
		tmpl, err := template.ParseFS(sh.TemplatesFs, tmplPath)
		if err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to parse the template: ",
				helpers.ErrAttr(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		metrics, err := sh.services.GetAll()
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
		record, err := sh.services.Get(entities.MetricName(metric.ID), entities.MetricType(metric.MType))
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("metric with type=%s, name=%s not found: %w", metric.MType, metric.MType, err)
			}

			return nil, fmt.Errorf("failed to get the given metric: %w", err)
		}

		return &record, nil
	}

	if entities.MetricType(metric.MType) == entities.GaugeMetricName {
		record, err := sh.services.Get(entities.MetricName(metric.ID), entities.MetricType(metric.MType))
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

// formatBodyMessageErrors formats and returns an error message for body reading errors.
func formatBodyMessageErrors(err error) error {
	return fmt.Errorf("error reading body: %w", err)
}

// handleUploads handles metric uploads, validating and storing them based on their type.
// //nolint:godot // this comment is part of the Swagger documentation
// @Summary Upload a metric
// @Description Uploads a metric of type gauge or counter. Returns status OK if successful.
// @Tags metrics
// @Accept text/plain
// @Produce text/plain
// @Param metricType path string true "Type of metric (counter/gauge)"
// @Param metricName path string true "Name of the metric"
// @Param metricValue path string true "Value of the metric"
// @Success 200 {string} string "Metric uploaded successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Metric type not found"
// @Router /upload/{metricType}/{metricName}/{metricValue} [post]
func (sh *ServerHandler) handleUploads(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "metricType")
	metricName := chi.URLParam(r, "metricName")
	metricValue := chi.URLParam(r, "metricValue")
	sh.logger.InfoContext(r.Context(),
		"received metric type: ",
		slog.String("metric_type", metricType),
		slog.String("metric_name", metricName),
		slog.String("metric_value", metricValue))
	// return http.StatusNotFound if metric type is not provided
	if entities.MetricType(metricType) != entities.GaugeMetricName &&
		entities.MetricType(metricType) != entities.CounterMetricName ||
		metricValue == "" {
		sh.logger.DebugContext(r.Context(),
			"invalid metric type")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch entities.MetricType(metricType) {
	case entities.CounterMetricName:
		delta, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			sh.logger.DebugContext(r.Context(),
				"failed to convert counter value to integer",
				slog.String("delta", metricValue))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		metric := entities.Metrics{ID: metricName, MType: metricType, Delta: &delta}
		if err = sh.services.Create(metric); err != nil {
			sh.logger.DebugContext(r.Context(),
				"failed to create the metric",
				helpers.ErrAttr(err),
				slog.String("url", r.RequestURI))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	case entities.GaugeMetricName:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			sh.logger.DebugContext(r.Context(),
				"failed to convert gauge value to float64",
				slog.String("value", metricValue))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		metric := entities.Metrics{ID: metricName, MType: metricType, Value: &value}
		if err = sh.services.Create(metric); err != nil {
			sh.logger.DebugContext(r.Context(),
				"failed to create the metric",
				helpers.ErrAttr(err),
				slog.String("url", r.RequestURI))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set(helpers.ContentType, helpers.ContentPlainText)

	// Upon successful reception, return StatusOK
	w.WriteHeader(http.StatusOK)
}

// handleBatchUploads handles batch metric uploads, validating the metrics and storing them.
// //nolint:godot // this comment is part of the Swagger documentation
// @Summary Upload a batch of metrics
// @Description Uploads multiple metrics in a single request. Returns status OK if successful.
// @Tags metrics
// @Accept application/json
// @Produce application/json
// @Param body []entities.Metrics true "List of metrics"
// @Success 200 {string} string "Metrics uploaded successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Metric type not found"
// @Router /upload/batch [post]
func (sh *ServerHandler) handleBatchUploads(w http.ResponseWriter, r *http.Request) {
	metrics := make([]entities.Metrics, 0)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.DebugContext(r.Context(), "failed to read request body")
		http.Error(w, formatBodyMessageErrors(err).Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		if err = r.Body.Close(); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to close request body",
				helpers.ErrAttr(err))
		}
	}()

	if sh.secret != "" {
		if !isBodyValid(body, r.Header.Get("HashSHA256"), sh.secret) {
			sh.logger.DebugContext(r.Context(),
				"request body failed integrity check")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		sh.logger.DebugContext(r.Context(),
			"request body passed integrity check")
	}

	isEncrypted := r.Header.Get("X-Encryption") == "RSA-AES"

	if isEncrypted {
		if sh.privateKey == nil {
			http.Error(w, "Server not configured for encryption", http.StatusInternalServerError)
			return
		}

		// Decrypt data
		keySize := sh.privateKey.Size()
		encryptedKey := body[:keySize]
		encryptedData := body[keySize:]
		aesKey, err := rsa.DecryptPKCS1v15(rand.Reader, sh.privateKey, encryptedKey)
		if err != nil {
			sh.logger.DebugContext(r.Context(), "failed to decrypt AES key", helpers.ErrAttr(err))
			http.Error(w, "Failed to decrypt AES key", http.StatusBadRequest)
			return
		}

		block, err := aes.NewCipher(aesKey)
		if err != nil {
			http.Error(w, "Failed to create AES cipher", http.StatusInternalServerError)
			return
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			http.Error(w, "Failed to create GCM", http.StatusInternalServerError)
			return
		}

		// Extract nonce and ciphertext
		nonceSize := gcm.NonceSize()
		if len(encryptedData) < nonceSize {
			http.Error(w, "Malformed encrypted data", http.StatusBadRequest)
			return
		}
		nonce := encryptedData[:nonceSize]
		ciphertext := encryptedData[nonceSize:]

		// Decrypt the data
		body, err = gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			http.Error(w, "Failed to decrypt data", http.StatusBadRequest)
			return
		}
	}

	err = json.Unmarshal(body, &metrics)
	if err != nil {
		sh.logger.DebugContext(r.Context(),
			"failed to unmarshal the request",
			slog.String(bodyKey, string(body)))
		http.Error(w, formatBodyMessageErrors(err).Error(), http.StatusInternalServerError)
		return
	}

	// Validate each metric type
	for _, metric := range metrics {
		if !isValidMetricType(metric.MType) {
			sh.logger.DebugContext(r.Context(),
				"unsupported metric type",
				slog.String("metric_type", metric.MType))
			http.Error(w, "Unsupported metric type", http.StatusInternalServerError)
			return
		}
	}

	sh.logger.DebugContext(r.Context(),
		fmt.Sprintf("received batch of %d metrics", len(metrics)))
	// return http.StatusNotFound if metric type is not provided

	w.Header().Set("Content-Type", "application/json")
	err = sh.services.StoreMetricsBatch(metrics)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to batch write the metrics",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handleJSONUploads handles JSON metric uploads, validating and storing them.
// //nolint:godot // this comment is part of the Swagger documentation
// @Summary Upload a JSON metric
// @Description Uploads a single metric in JSON format. Returns status OK if successful.
// @Tags metrics
// @Accept application/json
// @Produce application/json
// @Param body entities.Metrics true "Metric"
// @Success 200 {object} entities.Metrics "Metric uploaded successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Metric type not found"
// @Router /upload/json [post]
func (sh *ServerHandler) handleJSONUploads(w http.ResponseWriter, r *http.Request) {
	metric := entities.Metrics{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		sh.logger.DebugContext(r.Context(), "failed to read request body")
		http.Error(w, formatBodyMessageErrors(err).Error(), http.StatusBadRequest)
		return
	}
	defer func() {
		if err = r.Body.Close(); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to close request body",
				helpers.ErrAttr(err))
		}
	}()

	// Handle JSON unmarshal error
	if err = json.Unmarshal(body, &metric); err != nil {
		sh.logger.DebugContext(r.Context(),
			"failed to unmarshal the request",
			slog.String(bodyKey, string(body)))
		http.Error(w, formatBodyMessageErrors(err).Error(), http.StatusBadRequest)
		return
	}

	sh.logger.DebugContext(r.Context(), "received", slog.String("metric", string(body)))
	// return http.StatusNotFound if metric type is not provided
	if metric.MType != string(entities.GaugeMetricName) &&
		metric.MType != string(entities.CounterMetricName) {
		sh.logger.DebugContext(r.Context(),
			"invalid metric received",
			slog.String("metric", string(body)),
		)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if !isMetricNameAndValuePresent(metric) {
		sh.logger.DebugContext(r.Context(),
			"invalid metric received",
			slog.String("metric", string(body)),
		)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	switch entities.MetricType(metric.MType) {
	case entities.CounterMetricName:
		if err = sh.services.Create(metric); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to crete the counter metric",
				helpers.ErrAttr(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	case entities.GaugeMetricName:
		if err = sh.services.Create(metric); err != nil {
			sh.logger.ErrorContext(r.Context(),
				"failed to create the gauge metric",
				helpers.ErrAttr(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	default:
		sh.logger.DebugContext(r.Context(),
			"no valid metric received")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	updatedMetric, err := sh.getResponseMetric(metric)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed generate response",
			helpers.ErrAttr(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(updatedMetric)
	if err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to marshal the response",
			helpers.ErrAttr(err),
			slog.String(bodyKey, string(body)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = w.Write(response); err != nil {
		sh.logger.ErrorContext(r.Context(),
			"failed to write the response",
			helpers.ErrAttr(err),
			slog.String("response", string(response)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// isMetricNameAndValuePresent checks if the metric has a valid name and value.
// It returns true if the metric is of type Counter and has a non-nil Delta and a non-empty ID,
// or if the metric is of type Gauge and has a non-nil Value and a non-empty ID.
func isMetricNameAndValuePresent(metric entities.Metrics) bool {
	if metric.MType == string(entities.CounterMetricName) &&
		metric.Delta != nil &&
		metric.ID != "" {
		return true
	}

	if metric.MType == string(entities.GaugeMetricName) &&
		metric.Value != nil &&
		metric.ID != "" {
		return true
	}

	return false
}

// getResponseMetric retrieves the current value of the metric for response generation.
// If the metric is of type Gauge, it returns the metric as is.
// If the metric is of type Counter, it retrieves the current delta value from the MetricsService.
// It returns the metric and any error encountered during retrieval.
func (sh *ServerHandler) getResponseMetric(metric entities.Metrics) (*entities.Metrics, error) {
	if entities.MetricType(metric.MType) == entities.GaugeMetricName {
		return &metric, nil
	} else {
		currentDelta, err := sh.services.Get(entities.MetricName(metric.ID), entities.CounterMetricName)
		if err != nil {
			return nil, fmt.Errorf("failed to get the counter value %w", err)
		}

		return &currentDelta, nil
	}
}

// isBodyValid verifies the integrity of the request body by comparing the provided hash with a computed hash.
// It returns true if the request hash is not empty and matches the computed hash using the provided secret.
func isBodyValid(data []byte, reqHash, secret string) bool {
	if reqHash == "" {
		return false
	}
	hashedStr := getHash(data, secret)
	return hashedStr == reqHash
}

// getHash computes the HMAC SHA-256 hash of the provided data using the provided secret.
// It returns the hex-encoded hash as a string.
func getHash(data []byte, secret string) string {
	hash := hmac.New(sha256.New, []byte(secret))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

// isValidMetricType checks if the provided metric type is valid.
// It returns true if the metric type is either Counter or Gauge; otherwise, it returns false.
func isValidMetricType(mType string) bool {
	return mType == string(entities.CounterMetricName) || mType == string(entities.GaugeMetricName)
}
