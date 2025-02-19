// Package server implements the gRPC server for the metrics service.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
	"github.com/mihailtudos/metrickit/pkg/helpers"
	pb "github.com/mihailtudos/metrickit/proto/metrics"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MetricsService implement interface.
type MetricsService struct {
	pb.UnimplementedMetricServiceServer
	services *server.MetricsService
	logger   *slog.Logger
}

const (
	metricKey = "metric"
)

// NewMetricsService creates a new MetricsService.
func NewMetricsService(services *server.MetricsService,
	logger *slog.Logger) *MetricsService {
	return &MetricsService{
		services: services,
		logger:   logger,
	}
}

// CreateMetric creates a new metric.
func (ms *MetricsService) CreateMetric(ctx context.Context,
	req *pb.CreateMetricRequest) (*pb.CreateMetricResponse, error) {
	metric := req.GetMetric()

	ms.logger.InfoContext(ctx, "received", slog.Any(metricKey, metric))

	// Validate request
	if err := metric.Validate(); err != nil {
		ms.logger.InfoContext(ctx, "Validation failed: %v", helpers.ErrAttr(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	err := ms.services.Create(entities.Metrics{
		ID:    metric.GetId(),
		MType: metric.GetMType(),
		Value: proto.Float64(metric.GetValue()),
		Delta: proto.Int64(metric.GetDelta()),
	})

	if err != nil {
		return nil, fmt.Errorf("create metric: %w", err)
	}

	return &pb.CreateMetricResponse{Message: "Metric created successfully"}, fmt.Errorf("create metric: %w", err)
}

func (ms *MetricsService) CreateMetrics(ctx context.Context,
	req *pb.CreateMetricsRequest) (*pb.CreateMetricsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		ms.logger.InfoContext(ctx, "validation failed", helpers.ErrAttr(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	m := req.GetMetrics()
	ms.logger.InfoContext(ctx, "received metrics", slog.Any(metricKey, m))
	metrics := make([]entities.Metrics, 0, len(m))

	// Handle bulk creation logic
	for _, metric := range m {
		ms.logger.InfoContext(ctx, "processing metric", slog.Any(metricKey, metric))
		mm := entities.Metrics{
			ID:    metric.GetId(),
			MType: metric.GetMType(),
		}

		if mm.MType == string(entities.CounterMetricName) {
			mm.Delta = proto.Int64(metric.GetDelta())
		} else {
			mm.Value = proto.Float64(metric.GetValue())
		}

		metrics = append(metrics, mm)
	}

	ms.logger.InfoContext(ctx, "metrics to be stored", slog.Any("metrics", metrics))

	if err := ms.services.StoreMetricsBatch(metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store metrics: %v", err)
	}

	return &pb.CreateMetricsResponse{
		Message: "Metrics created successfully",
	}, nil
}

func (ms *MetricsService) GetMetric(ctx context.Context,
	req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		ms.logger.InfoContext(ctx, "validation failed", helpers.ErrAttr(err))
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// Retrieve the metric by ID from the database
	id := req.GetId()
	mType := req.GetMType()
	ms.logger.InfoContext(ctx, "received metric",
		slog.String("id", id), slog.String("mType", mType))

	m, err := ms.services.Get(entities.MetricName(id), entities.MetricType(mType))
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "metric not found")
		}

		return nil, status.Errorf(codes.Internal, "server error: %v", err)
	}

	return &pb.GetMetricResponse{
		Metric: &pb.Metric{
			Id:    m.ID,
			MType: m.MType,
			Value: m.Value,
			Delta: m.Delta,
		},
		Message: "Metric retrieved successfully",
	}, nil
}

func (ms *MetricsService) GetMetrics(ctx context.Context,
	_ *emptypb.Empty) (*pb.GetMetricsResponse, error) {
	m, err := ms.services.GetAll()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "metric not found")
		}

		return nil, status.Errorf(codes.Internal, "server error: %v", err)
	}

	metrics := make([]*pb.Metric, 0, len(m.Counter)+len(m.Gauge))

	for k, metric := range m.Counter {
		ms.logger.InfoContext(ctx, "processing counter metric",
			slog.Any(metricKey, metric))
		metrics = append(metrics, &pb.Metric{
			Id:    string(k),
			MType: string(entities.CounterMetricName),
			Delta: proto.Int64(int64(metric)),
		})
	}

	for k, metric := range m.Gauge {
		ms.logger.InfoContext(ctx, "processing gauge metric",
			slog.Any(metricKey, metric))
		metrics = append(metrics, &pb.Metric{
			Id:    string(k),
			MType: string(entities.GaugeMetricName),
			Value: proto.Float64(float64(metric)),
		})
	}

	return &pb.GetMetricsResponse{
		Metric:  metrics,
		Message: "Metrics retrieved successfully",
	}, nil
}
