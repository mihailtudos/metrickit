package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/mihailtudos/metrickit/internal/config"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/pkg/helpers"
)

type DBStore struct {
	db  *sql.DB
	cfg *config.ServerConfig
}

func NewPostgresStorage(cfg *config.ServerConfig) (*DBStore, error) {
	dbs := &DBStore{
		db:  cfg.DB,
		cfg: cfg,
	}

	if err := dbs.createScheme(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to create the db schema %w", err)
	}

	return dbs, nil
}

func (ds *DBStore) CreateRecord(metric entities.Metrics) error {
	var err error
	ctx := context.Background()

	switch {
	case metric.Delta != nil:
		err = ds.createCounterMetric(ctx, metric)
	case metric.Value != nil:
		_, err = ds.db.ExecContext(ctx, `
			INSERT INTO gauge_metrics (name, value) VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET value = EXCLUDED.value
		`, metric.ID, *metric.Value)
	default:
		return errors.New("invalid metric: must have either delta or value set")
	}

	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	return nil
}

func (ds *DBStore) GetRecord(mName entities.MetricName, mType entities.MetricType) (entities.Metrics, error) {
	var metrics entities.Metrics
	var err error

	ctx := context.Background()
	switch mType {
	case "counter":
		err = ds.db.QueryRowContext(ctx, `
			SELECT name, delta FROM counter_metrics WHERE name = $1
		`, mName).Scan(&metrics.ID, &metrics.Delta)
		metrics.MType = "counter"
	case "gauge":
		err = ds.db.QueryRowContext(ctx, `
			SELECT name, value FROM gauge_metrics WHERE name = $1
		`, mName).Scan(&metrics.ID, &metrics.Value)
		metrics.MType = "gauge"
	default:
		return metrics, fmt.Errorf("invalid metric type: %s", mType)
	}

	if err != nil {
		if err == sql.ErrNoRows {
			return metrics, fmt.Errorf("metric not found: %w", ErrNotFound)
		}
		return metrics, fmt.Errorf("failed to get record: %w", err)
	}

	return metrics, nil
}

func (ds *DBStore) GetAllRecords() (*MetricsStorage, error) {
	ctx := context.Background()
	metricsStorage := &MetricsStorage{
		Counter: make(map[entities.MetricName]entities.Counter),
		Gauge:   make(map[entities.MetricName]entities.Gauge),
	}

	rows, err := ds.db.QueryContext(ctx, `
		SELECT name, value FROM gauge_metrics
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all gauge records: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			ds.cfg.Log.ErrorContext(ctx, "failed to close gauge rows", helpers.ErrAttr(err))
		}
	}()

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

	rows, err = ds.db.QueryContext(ctx, `
		SELECT name, delta FROM counter_metrics
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get all counter records: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			ds.cfg.Log.ErrorContext(ctx, "failed to close counter rows", helpers.ErrAttr(err))
		}
	}()

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

func (ds *DBStore) GetAllRecordsByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error) {
	ctx := context.Background()
	metricsMap := make(map[entities.MetricName]entities.Metrics)

	switch mType {
	case entities.CounterMetricName:
		rows, err := ds.db.QueryContext(ctx, `
			SELECT name, delta FROM counter_metrics
		`)
		if err != nil {
			return nil, fmt.Errorf("failed to get all counter records by type: %w", err)
		}

		defer func() {
			if err := rows.Close(); err != nil {
				ds.cfg.Log.ErrorContext(ctx, "failed to close counter rows", helpers.ErrAttr(err))
			}
		}()

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

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("reading counter db rows error %w", err)
		}
	case entities.GaugeMetricName:
		rows, err := ds.db.QueryContext(ctx, `
			SELECT name, value FROM gauge_metrics
		`)
		if err != nil {
			return nil, fmt.Errorf("failed to get all gauge records by type: %w", err)
		}
		defer func() {
			if err := rows.Close(); err != nil {
				ds.cfg.Log.ErrorContext(ctx, "failed to close gauge rows", helpers.ErrAttr(err))
			}
		}()

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

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("reading gauge db rows error %w", err)
		}
	default:
		return nil, fmt.Errorf("invalid metric type: %s", mType)
	}

	return metricsMap, nil
}

func (ds *DBStore) Close() error {
	return fmt.Errorf("failed to close connection %w", ds.db.Close())
}

func (ds *DBStore) createScheme(ctx context.Context) error {
	trx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start a transaction %w", err)
	}
	defer func() {
		if err := trx.Rollback(); err != nil {
			if errors.Is(err, sql.ErrTxDone) {
				ds.cfg.Log.DebugContext(ctx, "failed to rollback the transaction ", helpers.ErrAttr(err))
			}
		}
	}()

	createGaugeTable := `
		CREATE TABLE IF NOT EXISTS gauge_metrics (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			value DOUBLE PRECISION NOT NULL
		);
	`

	createCounterTable := `
		CREATE TABLE IF NOT EXISTS counter_metrics (
			id SERIAL PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			delta BIGINT NOT NULL CHECK  (delta >= 0) 
		);
	`

	createStatements := []string{createGaugeTable, createCounterTable}

	for _, stmt := range createStatements {
		if _, err := trx.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("failed to execute transaction %w", err)
		}
	}

	if err := trx.Commit(); err != nil {
		return fmt.Errorf("failed to commit the transaction %w", err)
	}

	return nil
}

func (ds *DBStore) createCounterMetric(ctx context.Context, metric entities.Metrics) error {
	existingMetric, err := ds.GetRecord(entities.MetricName(metric.ID), entities.MetricType(metric.MType))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("create counter metric failed to get current metric: %w", err)
	}

	if errors.Is(err, ErrNotFound) {
		_, err = ds.db.ExecContext(ctx, `
			INSERT INTO counter_metrics (name, delta) VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET delta = EXCLUDED.delta
		`, metric.ID, *metric.Delta)
		if err != nil {
			return fmt.Errorf("failed to create new counter metric: %w", err)
		}
	} else {
		newValue := *metric.Delta + *existingMetric.Delta
		metric.Delta = &newValue
		_, err = ds.db.ExecContext(ctx, `
			INSERT INTO counter_metrics (name, delta) VALUES ($1, $2)
			ON CONFLICT (name) DO UPDATE SET delta = EXCLUDED.delta
		`, metric.ID, *metric.Delta)
		if err != nil {
			return fmt.Errorf("failed to update existing counter metric: %w", err)
		}
	}

	return nil
}
