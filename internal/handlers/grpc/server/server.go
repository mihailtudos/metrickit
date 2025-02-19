package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/mihailtudos/metrickit/internal/domain/entities"
	"github.com/mihailtudos/metrickit/internal/infrastructure/storage"
	"github.com/mihailtudos/metrickit/internal/service/server"
	pb "github.com/mihailtudos/metrickit/proto/metrics"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"log"
)

// MetricsService implement interface
type MetricsService struct {
	pb.UnimplementedMetricServiceServer
	services *server.MetricsService
}

// NewMetricsService creates a new MetricsService.
func NewMetricsService(services *server.MetricsService) *MetricsService {
	return &MetricsService{
		services: services,
	}
}

// CreateMetric creates a new metric.
func (ms *MetricsService) CreateMetric(ctx context.Context, req *pb.CreateMetricRequest) (*pb.CreateMetricResponse, error) {
	log.Printf("Received metric: %v", req.Metric)
	metric := req.GetMetric()
	fmt.Printf("Create metric: %+v\n", metric)

	// Validate request
	if err := metric.Validate(); err != nil {
		log.Printf("Validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	err := ms.services.Create(entities.Metrics{
		ID:    metric.Id,
		MType: metric.MType,
		Value: metric.Value,
		Delta: metric.Delta,
	})

	if err != nil {
		return nil, err
	}

	return &pb.CreateMetricResponse{Message: "Metric created successfully"}, err
}

func (ms *MetricsService) CreateMetrics(ctx context.Context, req *pb.CreateMetricsRequest) (*pb.CreateMetricsResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		log.Printf("Validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}
	m := req.GetMetrics()
	fmt.Printf("Received metrics: %+v\n", m)
	metrics := make([]entities.Metrics, 0, len(m))

	// Handle bulk creation logic
	for _, metric := range m {
		fmt.Printf("Processing metric: %+v\n", metric)
		mm := entities.Metrics{
			ID:    metric.Id,
			MType: metric.MType,
		}

		if mm.MType == string(entities.CounterMetricName) {
			mm.Delta = metric.Delta
		} else {
			mm.Value = metric.Value
		}

		metrics = append(metrics, mm)
	}

	fmt.Printf("Metrics to be stored: %+v\n", metrics)

	if err := ms.services.StoreMetricsBatch(metrics); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to store metrics: %v", err)
	}

	return &pb.CreateMetricsResponse{
		Message: "Metrics created successfully",
	}, nil
}

func (ms *MetricsService) GetMetric(ctx context.Context, req *pb.GetMetricRequest) (*pb.GetMetricResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		log.Printf("Validation failed: %v", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid request: %v", err)
	}

	// Retrieve the metric by ID from the database
	id := req.GetId()
	mType := req.GetMType()
	log.Printf("Received metric id: %q and type: %q", id, mType)

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

func (ms *MetricsService) GetMetrics(ctx context.Context, _ *emptypb.Empty) (*pb.GetMetricsResponse, error) {
	m, err := ms.services.GetAll()
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, status.Errorf(codes.NotFound, "metric not found")
		}

		return nil, status.Errorf(codes.Internal, "server error: %v", err)
	}

	metrics := make([]*pb.Metric, 0, len(m.Counter)+len(m.Gauge))

	for k, metric := range m.Counter {
		fmt.Printf("Metric: %+v, %v\n", metric, k)
		metrics = append(metrics, &pb.Metric{
			Id:    string(k),
			MType: string(entities.CounterMetricName),
			Delta: proto.Int64(int64(metric)),
		})
	}

	for k, metric := range m.Gauge {
		fmt.Printf("Metric: %+v, %v\n", metric, k)
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
