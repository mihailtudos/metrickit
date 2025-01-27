// Package database provides functions for initializing and managing connections
// to a PostgreSQL database using the pgx library. It includes support for connection
// pooling, query tracing, and retry mechanisms with backoff.
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

// queryTracer is a custom tracer that logs the start and end of database queries.
// It implements pgx's QueryTracer interface to track SQL execution and arguments.
type queryTracer struct{}

// TraceQueryStart logs the start of a database query, including the SQL statement
// and its arguments.
func (t *queryTracer) TraceQueryStart(
	ctx context.Context,
	_ *pgx.Conn,
	data pgx.TraceQueryStartData,
) context.Context {
	log.Printf("Running query %s (%v)", data.SQL, data.Args)
	return ctx
}

// TraceQueryEnd logs the end of a database query, including the command tag returned
// by the query execution.
func (t *queryTracer) TraceQueryEnd(_ context.Context, _ *pgx.Conn, data pgx.TraceQueryEndData) {
	log.Printf("%v", data.CommandTag)
}

// InitPostgresDB initializes a PostgreSQL connection pool with retries and backoff.
// It tries to establish the connection up to a specified number of retries, with
// exponential backoff in case of failures.
//
// Parameters:
// - ctx: The context for managing request lifetime.
// - dsn: The Data Source Name (DSN) for the database connection.
// - logger: A logger for structured logging.
//
// Returns:
// - A pgxpool.Pool instance representing the database connection pool.
// - An error if all attempts to connect fail.
func InitPostgresDB(ctx context.Context, dsn string, logger *slog.Logger) (*pgxpool.Pool, error) {
	retries := 3
	backoffIntervals := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

	for attempt := 0; attempt <= retries; attempt++ {
		// Apply backoff wait duration between retries
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
		cancel() // Ensure context is canceled after each attempt
		if err == nil {
			return pool, nil
		}
		logger.DebugContext(ctx,
			fmt.Sprintf("attempt %d: failed to initialize DB", attempt+1),
			helpers.ErrAttr(err))
	}

	return nil, errors.New("all attempts to initialize DB failed")
}

// attemptInit performs the actual initialization of the PostgreSQL connection pool.
//
// Parameters:
// - ctx: The context for managing request lifetime.
// - dsn: The Data Source Name (DSN) for the database connection.
// - logger: A logger for structured logging.
//
// Returns:
// - A pgxpool.Pool instance representing the database connection pool.
// - An error if the pool initialization fails.
func attemptInit(ctx context.Context, dsn string, logger *slog.Logger) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	poolCfg.ConnConfig.Tracer = &queryTracer{}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		handlePingError(ctx, err, logger)
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return pool, nil
}

// handlePingError logs detailed information about a database ping error.
// It checks if the error is a PostgreSQL error and logs appropriate details.
func handlePingError(ctx context.Context, err error, logger *slog.Logger) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			logger.ErrorContext(ctx, "unique violation error while pinging the DB", helpers.ErrAttr(err))
		case pgerrcode.ConnectionException:
			logger.ErrorContext(ctx, "connection exception error while pinging the DB", helpers.ErrAttr(err))
		default:
			logger.ErrorContext(ctx,
				"PostgreSQL error: "+pgErr.Code,
				helpers.ErrAttr(err))
		}
	} else {
		// General error logging
		logger.ErrorContext(ctx, "general error while pinging the DB", helpers.ErrAttr(err))
	}
}

// getWaitDuration returns the backoff wait duration for a given attempt.
// If the attempt exceeds the length of backoffIntervals, it returns zero.
func getWaitDuration(backoffIntervals []time.Duration, attempt int) time.Duration {
	if attempt >= len(backoffIntervals) {
		return 0
	}
	return backoffIntervals[attempt]
}
