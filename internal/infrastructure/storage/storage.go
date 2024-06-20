package storage

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
)

type Storage interface {
	CreateRecord(metrics entities.Metrics) error
	GetRecord(mName entities.MetricName, mType entities.MetricType) (entities.Metrics, error)
	GetAllRecords() (*MetricsStorage, error)
	GetAllRecordsByType(mType entities.MetricType) (map[entities.MetricName]entities.Metrics, error)
	StoreMetricsBatch(metrics []entities.Metrics) error
	Close(ctx context.Context) error
}

func NewStorage(db *pgxpool.Pool, logger *slog.Logger, storeInterval int, storePath string) (Storage, error) {
	if db != nil {
		return NewPostgresStorage(db, logger)
	}

	if storeInterval >= 0 {
		return NewFileStorage(logger, storeInterval, storePath)
	}

	return NewMemStorage(logger)
}
