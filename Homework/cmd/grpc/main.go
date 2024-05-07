package main

import (
	"Homework/internal/db"
	"Homework/internal/domain"
	"Homework/internal/model"
	"Homework/internal/pb"
	"Homework/internal/repository"
	"Homework/internal/repository/in_memory_cache"
	"Homework/internal/repository/postgresql"
	repositoryRedis "Homework/internal/repository/redis"
	"Homework/internal/service"
	"Homework/internal/storage"
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	reg = prometheus.NewRegistry()

	givenOrdersMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "given_orders",
		Help: "Total number of given orders",
	})

	refundedOrdersMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "refunded_orders",
		Help: "Total number of refunded orders",
	})

	addedPickupPointsMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "added_pickup_points",
		Help: "Total number of added pickup points",
	})

	deletedPickupPointsMetric = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "deleted_pickup_points",
		Help: "Total number of deleted pickup points",
	})
)

func initProvider() (func(context.Context) error, error) {
	ctx := context.Background()

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("test-service"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	traceExporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithEndpoint(os.Getenv("JAEGER_HOST")+":"+os.Getenv("JAEGER_PORT")))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tracerProvider.Shutdown, nil
}

func init() {
	reg.MustRegister(givenOrdersMetric)
	reg.MustRegister(refundedOrdersMetric)
	reg.MustRegister(addedPickupPointsMetric)
	reg.MustRegister(deletedPickupPointsMetric)
}

type PickupPointService struct {
	pb.UnimplementedPickupPointsServer
	BusinessService domain.BusinessService
}

func (p *PickupPointService) Add(ctx context.Context, req *pb.PickupPointRequest) (*pb.PickupPointId, error) {
	if req.Name == "" {
		return &pb.PickupPointId{}, status.Errorf(codes.InvalidArgument, "name of pickup point isn't specified")
	}

	if req.Address == "" {
		return &pb.PickupPointId{}, status.Errorf(codes.InvalidArgument, "address of pickup point isn't specified")
	}

	if req.Name == "" {
		return &pb.PickupPointId{}, status.Errorf(codes.InvalidArgument, "contact details of pickup point isn't specified")
	}

	pickupPointRepo := &repository.PickupPoint{
		Name:           req.Name,
		Address:        req.Address,
		ContactDetails: req.ContactDetails,
	}

	id, err := p.BusinessService.Add(ctx, *pickupPointRepo)
	if err != nil {
		return &pb.PickupPointId{}, err
	}
	defer addedPickupPointsMetric.Add(1)
	return &pb.PickupPointId{Id: id}, nil
}

func (p *PickupPointService) GetById(ctx context.Context, req *pb.PickupPointId) (*pb.PickupPointWithId, error) {
	if req.Id == 0 {
		return &pb.PickupPointWithId{}, status.Errorf(codes.InvalidArgument, "ID of pickup point isn't specified")
	}

	pickupPoint, err := p.BusinessService.GetByID(ctx, req.Id)
	if err != nil {
		return &pb.PickupPointWithId{}, err
	}
	return &pb.PickupPointWithId{
		Id:             pickupPoint.ID,
		Name:           pickupPoint.Name,
		Address:        pickupPoint.Address,
		ContactDetails: pickupPoint.ContactDetails,
	}, nil
}

func (p *PickupPointService) Delete(ctx context.Context, req *pb.PickupPointId) (*empty.Empty, error) {
	if req.Id == 0 {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "ID of pickup point isn't specified")
	}

	err := p.BusinessService.Delete(ctx, req.Id)
	if err != nil {
		return &empty.Empty{}, err
	}
	defer deletedPickupPointsMetric.Add(1)
	return &empty.Empty{}, nil
}

func (p *PickupPointService) Update(ctx context.Context, req *pb.PickupPointWithId) (*empty.Empty, error) {
	if req.Id == 0 {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "ID of pickup point isn't specified")
	}

	if req.Name == "" {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "name of pickup point isn't specified")
	}

	if req.Address == "" {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "address of pickup point isn't specified")
	}

	if req.ContactDetails == "" {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "contact details of pickup point isn't specified")
	}

	pickupPoint := &repository.PickupPoint{
		ID:             req.Id,
		Name:           req.Name,
		Address:        req.Address,
		ContactDetails: req.ContactDetails,
	}

	err := p.BusinessService.Update(ctx, *pickupPoint)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (p *PickupPointService) List(ctx context.Context, _ *empty.Empty) (*pb.PickupPointResponseList, error) {
	pickupPoints, err := p.BusinessService.List(ctx)
	if err != nil {
		return &pb.PickupPointResponseList{}, err
	}
	pickupPointsProto := pb.PickupPointResponseList{}
	for _, pickupPoint := range pickupPoints {
		pickupPointsProto.PickupPoints = append(pickupPointsProto.PickupPoints, &pb.PickupPointWithId{
			Id:             pickupPoint.ID,
			Name:           pickupPoint.Name,
			Address:        pickupPoint.Address,
			ContactDetails: pickupPoint.ContactDetails,
		})
	}
	return &pickupPointsProto, nil
}

type OrderService struct {
	pb.UnimplementedOrdersServer
	OrderService service.OrderService
}

func (o *OrderService) AcceptOrderFromCourier(_ context.Context, req *pb.OrderRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "ID of order isn't specified")
	}

	if req.CustomerId == "" {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "Customer id isn't specified")
	}

	if req.TermKeeping == nil {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "term keeping of order isn't specified")
	}

	if req.Weight == 0 {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "weight of order isn't specified")
	}

	if req.Price == 0 {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "price of order isn't specified")
	}

	termKeeping, err := ptypes.Timestamp(req.TermKeeping)
	if err != nil {
		return &empty.Empty{}, err
	}
	orderInputDto := model.OrderInputDTO{
		OrderId:     req.Id,
		CustomerId:  req.CustomerId,
		TermKeeping: termKeeping,
		Weight:      req.Weight,
		Price:       req.Price,
	}
	err = o.OrderService.AcceptOrderFromCourier(orderInputDto, req.PackageType)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (o *OrderService) ReturnOrderToCourier(_ context.Context, req *pb.OrderId) (*empty.Empty, error) {
	if req.Id == "" {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "ID of order isn't specified")
	}

	err := o.OrderService.ReturnOrderToCourier(req.Id)
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (o *OrderService) GiveOrder(_ context.Context, req *pb.OrderIdList) (*empty.Empty, error) {
	orders := make([]string, 0)
	for _, OrderId := range req.OrderIds {
		orders = append(orders, OrderId.Id)
	}

	if len(orders) == 0 {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "list of orders id is empty")
	}

	err := o.OrderService.GiveOrder(orders)
	if err != nil {
		return &empty.Empty{}, err
	}
	defer givenOrdersMetric.Add(1)
	return &empty.Empty{}, nil
}

func (o *OrderService) AcceptRefund(_ context.Context, req *pb.RefundRequest) (*empty.Empty, error) {
	if req.OrderId == "" {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "ID of order isn't specified")
	}

	if req.CustomerId == "" {
		return &empty.Empty{}, status.Errorf(codes.InvalidArgument, "Customer id isn't specified")
	}

	err := o.OrderService.AcceptRefund(req.CustomerId, req.OrderId)
	if err != nil {
		return &empty.Empty{}, err
	}
	defer refundedOrdersMetric.Add(1)
	return &empty.Empty{}, nil
}

func (o *OrderService) GetOrders(_ context.Context, req *pb.GetRequest) (*pb.OrderResponseList, error) {
	if req.CustomerId == "" {
		return &pb.OrderResponseList{}, status.Errorf(codes.InvalidArgument, "Customer id isn't specified")
	}

	statuses := map[model.OrderStatus]pb.OrderStatus{
		model.Issued:   pb.OrderStatus_ORDER_STATUS_ISSUED,
		model.Refunded: pb.OrderStatus_ORDER_STATUS_REFUNDED,
		model.Accepted: pb.OrderStatus_ORDER_STATUS_ACCEPTED,
	}
	orders, err := o.OrderService.GetOrders(req.CustomerId, int(req.Limit), req.OnlyNotIssued)
	if err != nil {
		return &pb.OrderResponseList{}, err
	}

	ordersProto := pb.OrderResponseList{}
	for _, order := range orders {
		ordersProto.Orders = append(ordersProto.Orders, &pb.OrderResponse{
			OrderId:     order.OrderId,
			CustomerId:  order.CustomerId,
			TermKeeping: timestamppb.New(order.TermKeeping),
			Status:      statuses[order.Status],
			Weight:      order.Weight,
			Price:       order.Price,
		})
	}
	return &ordersProto, nil
}

func (o *OrderService) GetListRefund(_ context.Context, req *pb.GetListRefundRequest) (*pb.OrderResponseList, error) {
	statuses := map[model.OrderStatus]pb.OrderStatus{
		model.Issued:   pb.OrderStatus_ORDER_STATUS_ISSUED,
		model.Refunded: pb.OrderStatus_ORDER_STATUS_REFUNDED,
		model.Accepted: pb.OrderStatus_ORDER_STATUS_ACCEPTED,
	}
	orders, err := o.OrderService.GetListRefund(int(req.PageNumber), int(req.PageSize))
	if err != nil {
		return &pb.OrderResponseList{}, err
	}

	ordersProto := pb.OrderResponseList{}
	for _, order := range orders {
		ordersProto.Orders = append(ordersProto.Orders, &pb.OrderResponse{
			OrderId:     order.OrderId,
			CustomerId:  order.CustomerId,
			TermKeeping: timestamppb.New(order.TermKeeping),
			Status:      statuses[order.Status],
			Weight:      order.Weight,
			Price:       order.Price,
		})
	}
	return &ordersProto, nil
}

func main() {
	if err := godotenv.Load("configs/config.env"); err != nil {
		log.Print("No .env file found")
		return
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	shutdownProvider, err := initProvider()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdownProvider(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	database, err := db.NewDb(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer database.GetPool(ctx).Close()

	pickupPointsRepo := postgresql.NewPickupPoints(database)
	inMemoryCache := in_memory_cache.NewInMemoryCache()
	newRedis := repositoryRedis.NewRedis(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "",
		DB:       0,
	})
	businessService := domain.BusinessService{Repo: pickupPointsRepo, InMemoryCache: inMemoryCache, Redis: newRedis}

	pickupPointSrv := PickupPointService{BusinessService: businessService}

	orderStor, err := storage.NewOrderStorage()
	if err != nil {
		log.Fatal("не удалось подключиться к хранилищу заказов")
		return
	}
	orderServ := service.NewOrderService(&orderStor)

	orderSrv := OrderService{OrderService: orderServ}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", os.Getenv("GRPC_PORT")))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	grpcMetrics := grpcPrometheus.NewServerMetrics()
	reg.MustRegister(grpcMetrics)

	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(grpcMetrics.UnaryServerInterceptor()),
		grpc.ChainStreamInterceptor(grpcMetrics.StreamServerInterceptor()),
	)

	pb.RegisterOrdersServer(grpcServer, &orderSrv)
	pb.RegisterPickupPointsServer(grpcServer, &pickupPointSrv)
	grpcMetrics.InitializeMetrics(grpcServer)

	go http.ListenAndServe(":"+os.Getenv("METRICS_PORT"), promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true}))

	log.Fatal(grpcServer.Serve(lis))
}
