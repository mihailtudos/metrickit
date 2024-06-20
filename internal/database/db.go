package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/mihailtudos/metrickit/pkg/helpers"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type queryTracer struct{}

func (t *queryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	log.Printf("Running query %s (%v)", data.SQL, data.Args)
	return ctx
}

func (t *queryTracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	log.Printf("%v", data.CommandTag)
}
func InitPostgresDB(ctx context.Context, dsn string, logger *slog.Logger) (*pgxpool.Pool, error) {
	retries := 3
	backoffIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for attempt := 0; attempt <= retries; attempt++ {
		// Calculate wait duration for backoff
		if attempt > 0 {
			waitDuration := getWaitDuration(backoffIntervals, attempt)
			logger.DebugContext(ctx,
				"retrying in %v seconds...\n",
				slog.Float64("duration", waitDuration.Seconds()))
			time.Sleep(waitDuration)
		}

		w := 1 + attempt
		attemptCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(w))

		pool, err := attemptInit(attemptCtx, dsn, logger)
		cancel() // Cancel context after attemptInit returns
		if err == nil {
			return pool, nil
		}
		logger.DebugContext(ctx,
			fmt.Sprintf("attempt %d: failed to initialize DB", attempt+1),
			helpers.ErrAttr(err))
	}

	return nil, errors.New("all attempts to initialize DB failed")
}

func attemptInit(ctx context.Context, dsn string, logger *slog.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the dsn: %w", err)
	}

	poolCfg.ConnConfig.Tracer = &queryTracer{}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				logger.ErrorContext(ctx, "unique violation error while pinging the DB", helpers.ErrAttr(err))
			case pgerrcode.ConnectionException:
				logger.ErrorContext(ctx, "connection exception error while pinging the DB", helpers.ErrAttr(err))
			// Add more cases as needed
			default:
				logger.ErrorContext(ctx,
					"PostgreSQL error: "+pgErr.Code,
					helpers.ErrAttr(err))
			}
		} else {
			// General error logging
			logger.ErrorContext(ctx, "general error while pinging the DB", helpers.ErrAttr(err))
		}
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return pool, nil
}

func getWaitDuration(backoffIntervals []time.Duration, attempt int) time.Duration {
	if attempt >= len(backoffIntervals) {
		return 0
	}
	return backoffIntervals[attempt]
}
