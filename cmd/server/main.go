package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vogiaan1904/ticketbottle-order/config"
	"github.com/vogiaan1904/ticketbottle-order/internal/infra/kafka"
	mongo "github.com/vogiaan1904/ticketbottle-order/internal/infra/mongo"
	oGrpc "github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/grpc"
	oKafka "github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/kafka/producer"
	oRepo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
	svc "github.com/vogiaan1904/ticketbottle-order/internal/order/service"
	eSvc "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/event"
	inSvc "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	opb "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/order"
	pSvc "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/payment"
	pkgLog "github.com/vogiaan1904/ticketbottle-order/pkg/logger"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	l := pkgLog.InitializeZapLogger(pkgLog.ZapConfig{
		Level:    cfg.Log.Level,
		Mode:     cfg.Log.Mode,
		Encoding: cfg.Log.Encoding,
	})

	mCli, err := mongo.Connect(cfg.Mongo)
	if err != nil {
		l.Fatalf(ctx, "Failed to connect to MongoDB: %v", err)
		os.Exit(1)
	}
	defer mongo.Disconnect(mCli)

	// Initialize gRpc service clients
	inSvcCli, invClose, err := inSvc.NewInventoryClient(cfg.Microservice.Inventory)
	if err != nil {
		l.Fatalf(ctx, "Failed to create inventory service client: %v", err)
		os.Exit(1)
	}
	defer invClose()

	eSvcCli, eClose, err := eSvc.NewEventClient(cfg.Microservice.Event)
	if err != nil {
		l.Fatalf(ctx, "Failed to create event service client: %v", err)
		os.Exit(1)
	}
	defer eClose()

	pmtSvcCli, pClose, err := pSvc.NewPaymentClient(cfg.Microservice.Payment)
	if err != nil {
		l.Fatalf(ctx, "Failed to create payment service client: %v", err)
		os.Exit(1)
	}
	defer pClose()

	// Initialize Kafka producer
	kProd, err := kafka.NewProducer(cfg.Kafka)
	if err != nil {
		l.Fatalf(ctx, "Failed to create Kafka producer: %v", err)
		os.Exit(1)
	}
	defer kProd.Close()

	// Initialize Kafka consumer group
	kConsGr, err := kafka.NewConsumerGroup(cfg.Kafka)
	if err != nil {
		l.Fatalf(ctx, "Failed to create Kafka consumer group: %v", err)
		os.Exit(1)
	}
	defer kConsGr.Close()

	// Initialize producers
	oProd := oKafka.NewProducer(kProd, l)

	// Initialize repositories
	db := mCli.Database(cfg.Mongo.Database)
	oRepo := oRepo.New(l, db)

	// Initialize services
	oSvc := svc.New(l, svc.Config{
		PaymentTimeoutSeconds: int32(cfg.Server.PaymentTimeoutSeconds),
	}, oRepo, inSvcCli, eSvcCli, pmtSvcCli, oProd)

	// Initialize gRpc services
	oGrpcSvc := oGrpc.NewGrpcService(oSvc, l)

	// Start gRpc server
	lnr, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.GRpcPort))
	if err != nil {
		l.Fatalf(ctx, "gRPC server failed to listen: %v", err)
	}

	gRpcSrv := grpc.NewServer()
	opb.RegisterOrderServiceServer(gRpcSrv, oGrpcSvc)

	go func() {
		l.Infof(ctx, "gRPC server is listening on port: %d", cfg.Server.GRpcPort)
		if err := gRpcSrv.Serve(lnr); err != nil {
			l.Fatalf(ctx, "Failed to serve gRPC: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info(ctx, "Server shutting down...")

	cancel()
	time.Sleep(1 * time.Second)
	gRpcSrv.GracefulStop()

	l.Info(ctx, "Server exited")
}
