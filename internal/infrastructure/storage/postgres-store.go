/*
Package storage provides functionality for storing and retrieving metrics in a PostgreSQL database.
It defines a DBStore struct that implements methods to create, read, and manage metrics data.
*/
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

// DBStore is a struct that provides methods for interacting with a PostgreSQL database to store and retrieve metrics.
type DBStore struct {
	db     *pgxpool.Pool // PostgreSQL connection pool
	logger *slog.Logger  // Logger for logging operations
}

// NewPostgresStorage creates a new DBStore instance and initializes the database schema.
// It takes a pgxpool.Pool for database connections and a logger for logging.
// Returns a pointer to the DBStore and an error if the initialization fails.
func NewPostgresStorage(db *pgxpool.Pool, logger *slog.Logger) (*DBStore, error) {
	dbs := &DBStore{
		db:     db,
		logger: logger,
	}

	if err := dbs.createScheme(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create the db schema %w", err)
	}

	return dbs, nil
}

// CreateRecord inserts a new metric record into the database or updates an existing one.
// It accepts an entities.Metrics object containing the metric data.
// Returns an error if the operation fails.
func (ds *DBStore) CreateRecord(metric entities.Metrics) error {
	var err error
	ctx := context.Background()

	switch {
	case metric.Delta != nil:
		err = ds.createCounterMetric(ctx, metric)
	case metric.Value != nil:
		_, err = ds.db.Exec(ctx, `
			INSERT INTO gauge_metrics (name, value) 
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE 
			    SET value = EXCLUDED.value
		`, metric.ID, *metric.Value)
	default:
		return errors.New("invalid metric: must have either delta or value set")
	}

	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	return nil
}

// GetRecord retrieves a metric record by its name and type from the database.
// It returns the corresponding entities.Metrics object and an error if the record is not found or another issue occurs.
func (ds *DBStore) GetRecord(mName entities.MetricName, mType entities.MetricType) (entities.Metrics, error) {
	var metrics entities.Metrics
	var err error

	ctx := context.Background()
	switch mType {
	case "counter":
		err = ds.db.QueryRow(ctx, `
			SELECT name, value FROM counter_metrics WHERE name = $1
		`, mName).Scan(&metrics.ID, &metrics.Delta)
		metrics.MType = "counter"
	case "gauge":
		err = ds.db.QueryRow(ctx, `
			SELECT name, value FROM gauge_metrics WHERE name = $1
		`, mName).Scan(&metrics.ID, &metrics.Value)
		metrics.MType = "gauge"
	default:
		return metrics, fmt.Errorf("invalid metric type: %s", mType)
	}

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return metrics, fmt.Errorf("metric not found: %w", ErrNotFound)
		}
		return metrics, fmt.Errorf("failed to get record: %w", err)
	}

	return metrics, nil
}

// GetAllRecords retrieves all metric records from the database and returns them as a MetricsStorage object.
// It includes both gauge and counter metrics.
func (ds *DBStore) GetAllRecords() (*MetricsStorage, error) {
	ctx := context.Background()
	metricsStorage := &MetricsStorage{
		Counter: make(map[entities.MetricName]entities.Counter),
		Gauge:   make(map[entities.MetricName]entities.Gauge),
	}

	// Retrieve all gauge metrics
	rows, err := ds.db.Query(ctx, `
		SELECT name, value FROM gauge_metrics
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all gauge records: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, fmt.Errorf("failed to scan gauge record: %w", err)
		}
		metricsStorage.Gauge[entities.MetricName(name)] = entities.Gauge(value)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading gauge db row error %w", err)
	}

	// Retrieve all counter metrics
	rows, err = ds.db.Query(ctx, `
		SELECT name, value FROM counter_metrics
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all counter records: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var delta int64
		if err := rows.Scan(&name, &delta); err != nil {
			return nil, fmt.Errorf("failed to scan counter record: %w", err)
		}
		metricsStorage.Counter[entities.MetricName(name)] = entities.Counter(delta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading counter type db rows error %w", err)
	}

	return metricsStorage, nil
}

// GetAllRecordsByType retrieves all metric records of a specific type (gauge or counter) from the database.
// It returns a map of entities.Metrics indexed by metric names and an error if the operation fails.
func (ds *DBStore) GetAllRecordsByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error) {
	ctx := context.Background()
	metricsMap := make(map[entities.MetricName]entities.Metrics)

	selectCountersStmt := `SELECT name, value FROM counter_metrics`
	selectGaugesStmt := `SELECT name, value FROM gauge_metrics`

	var stmt string

	switch mType {
	case entities.GaugeMetricName:
		stmt = selectGaugesStmt
	case entities.CounterMetricName:
		stmt = selectCountersStmt
	default:
		return nil, errors.New("invalid metric type")
	}

	rows, err := ds.db.Query(ctx, stmt)
	if err != nil {
		return nil, fmt.Errorf("failed to get all counter records by type: %w", err)
	}
	defer rows.Close()

	// Scan the results based on the metric type
	switch mType {
	case entities.CounterMetricName:
		for rows.Next() {
			var name string
			var delta int64
			if err := rows.Scan(&name, &delta); err != nil {
				return nil, fmt.Errorf("failed to scan counter record: %w", err)
			}
			metricsMap[entities.MetricName(name)] = entities.Metrics{
				ID:    name,
				MType: string(entities.CounterMetricName),
				Delta: &delta,
			}
		}
	case entities.GaugeMetricName:
		for rows.Next() {
			var name string
			var value float64
			if err := rows.Scan(&name, &value); err != nil {
				return nil, fmt.Errorf("failed to scan gauge record: %w", err)
			}
			metricsMap[entities.MetricName(name)] = entities.Metrics{
				ID:    name,
				MType: string(entities.GaugeMetricName),
				Value: &value,
			}
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("reading counter db rows error %w", err)
	}

	return metricsMap, nil
}

// Close shuts down the database connection pool and logs the action.
// It accepts a context for logging and returns an error if the shutdown fails.
func (ds *DBStore) Close(ctx context.Context) error {
	ds.logger.DebugContext(ctx,
		"shutting down the db connection pool")
	ds.db.Close()
	return nil
}

// createScheme initializes the database schema by creating necessary tables for metrics storage.
// It is called during the initialization of the DBStore.
// Returns an error if the schema creation fails.
func (ds *DBStore) createScheme(ctx context.Context) error {
	trx, err := ds.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start the transaction %w", err)
	}

	defer func() {
		if err = trx.Rollback(ctx); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				ds.logger.ErrorContext(ctx,
					"failed to rollback the transaction ",
					helpers.ErrAttr(err))
			}
		}
	}()

	createGaugeTable := `
		CREATE TABLE IF NOT EXISTS gauge_metrics (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			value DOUBLE PRECISION NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`

	createCounterTable := `
		CREATE TABLE IF NOT EXISTS counter_metrics (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			value BIGINT NOT NULL CHECK  (value >= 0) ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`

	createStatements := []string{
		createGaugeTable,
		createCounterTable,
	}

	for _, stmt := range createStatements {
		if _, err := trx.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute transaction %w", err)
		}
	}

	if err := trx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit the transaction %w", err)
	}

	return nil
}

// createCounterMetric creates or updates a counter metric in the database.
// If the metric already exists, its value is updated by adding the new delta.
// If the metric does not exist, it is created with the provided delta value.
func (ds *DBStore) createCounterMetric(ctx context.Context, metric entities.Metrics) error {
	existingMetric, err := ds.GetRecord(entities.MetricName(metric.ID), entities.MetricType(metric.MType))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("create counter metric failed to get current metric: %w", err)
	}

	if errors.Is(err, ErrNotFound) {
		_, err = ds.db.Exec(ctx, `
			INSERT INTO counter_metrics (name, value) 
			VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE 
			    SET value = EXCLUDED.value
		`, metric.ID, *metric.Delta)
		if err != nil {
			return fmt.Errorf("failed to create new counter metric: %w", err)
		}
	} else {
		newValue := *metric.Delta + *existingMetric.Delta
		metric.Delta = &newValue
		_, err = ds.db.Exec(ctx, `
			INSERT INTO counter_metrics (name, value) VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value
		`, metric.ID, *metric.Delta)
		if err != nil {
			return fmt.Errorf("failed to update existing counter metric: %w", err)
		}
	}

	return nil
}

// StoreMetricsBatch stores a batch of metrics in the database.
// It separates metrics into counters and gauges, processing them accordingly.
// Counter metrics are summed if they already exist, while gauge metrics
// only store the latest value.
func (ds *DBStore) StoreMetricsBatch(metrics []entities.Metrics) error {
	counterMetrics := make(map[string]entities.Metrics)
	gaugeMetrics := make(map[string]entities.Metrics)

	for _, metric := range metrics {
		switch entities.MetricType(metric.MType) {
		case entities.CounterMetricName:
			if existing, ok := counterMetrics[metric.ID]; ok {
				*existing.Delta += *metric.Delta
				counterMetrics[metric.ID] = existing
			} else {
				counterMetrics[metric.ID] = metric
			}
		case entities.GaugeMetricName:
			gaugeMetrics[metric.ID] = metric // only the latest gauge value is relevant
		}
	}

	ctx := context.Background()

	if len(counterMetrics) > 0 {
		existingCounter, err := ds.GetAllRecordsByType(entities.CounterMetricName)
		if err != nil {
			return fmt.Errorf("failed to get existing counter metrics: %w", err)
		}

		for id, metric := range counterMetrics {
			if v, ok := existingCounter[entities.MetricName(id)]; ok {
				*metric.Delta += *v.Delta
			}
			existingCounter[entities.MetricName(id)] = metric
		}

		counterMetricsList := make([]entities.Metrics, 0, len(existingCounter))
		for _, v := range existingCounter {
			counterMetricsList = append(counterMetricsList, v)
		}

		if err = ds.storeBatchMetrics(ctx, counterMetricsList); err != nil {
			return fmt.Errorf("store counter metrics failed %w", err)
		}
	}

	if len(gaugeMetrics) > 0 {
		gaugeMetricsList := make([]entities.Metrics, 0, len(gaugeMetrics))
		for _, v := range gaugeMetrics {
			gaugeMetricsList = append(gaugeMetricsList, v)
		}

		if err := ds.storeBatchMetrics(ctx, gaugeMetricsList); err != nil {
			return fmt.Errorf("store gauge metrics failed %w", err)
		}
	}

	return nil
}

// storeBatchMetrics stores a batch of metrics in the database within a transaction.
// It constructs and executes the necessary SQL statements to insert or update
// counter and gauge metrics.
func (ds *DBStore) storeBatchMetrics(ctx context.Context, metrics []entities.Metrics) error {
	tx, err := ds.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			ds.logger.ErrorContext(ctx, "failed to rollback", helpers.ErrAttr(err))
		}
	}()

	// Counters update statement
	countersUpdateStmt := `
		INSERT INTO counter_metrics (name, value)
		VALUES %s
		ON CONFLICT (name) DO UPDATE
		SET value = excluded.value, updated_at = NOW();
	`

	// Gauges update statement
	gaugesUpdateStmt := `
		INSERT INTO gauge_metrics (name, value)
		VALUES %s
		ON CONFLICT (name) DO UPDATE
		SET value = excluded.value, updated_at = NOW();
	`

	var counterValues []string
	var counterParams []interface{}
	var gaugeValues []string
	var gaugeParams []interface{}
	const metricsTypesCount = 2
	const firstParam = 1
	const secondParam = 2

	for i, metric := range metrics {
		if metric.Delta != nil {
			counterValues = append(counterValues, fmt.Sprintf("($%d, $%d)",
				metricsTypesCount*i+firstParam, metricsTypesCount*i+secondParam))
			counterParams = append(counterParams, metric.ID, *metric.Delta)
		}

		if metric.Value != nil {
			gaugeValues = append(gaugeValues, fmt.Sprintf("($%d, $%d)",
				metricsTypesCount*i+firstParam, metricsTypesCount*i+secondParam))
			gaugeParams = append(gaugeParams, metric.ID, *metric.Value)
		}
	}

	if len(counterValues) > 0 {
		_, err = ds.db.Exec(ctx, fmt.Sprintf(countersUpdateStmt, strings.Join(counterValues, ",")), counterParams...)
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			return fmt.Errorf("store counter SQL exec failed %w, started rollback %w", err, rollbackErr)
		}
	}

	if len(gaugeValues) > 0 {
		_, err = ds.db.Exec(ctx, fmt.Sprintf(gaugesUpdateStmt, strings.Join(gaugeValues, ",")), gaugeParams...)
		if err != nil {
			rollbackErr := tx.Rollback(ctx)
			return fmt.Errorf("store gauge SQL exec failed %w, started rollback %w", err, rollbackErr)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to execute transaction commit %w", err)
	}

	return nil
}
