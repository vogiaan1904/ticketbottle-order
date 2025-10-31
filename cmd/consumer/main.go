package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/vogiaan1904/ticketbottle-order/config"
	"github.com/vogiaan1904/ticketbottle-order/internal/activities"
	"github.com/vogiaan1904/ticketbottle-order/internal/infra/kafka"
	"github.com/vogiaan1904/ticketbottle-order/internal/infra/mongo"
	oCons "github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/kafka/consumer"
	oProd "github.com/vogiaan1904/ticketbottle-order/internal/order/delivery/kafka/producer"
	oRepo "github.com/vogiaan1904/ticketbottle-order/internal/order/repository"
	oSvc "github.com/vogiaan1904/ticketbottle-order/internal/order/service"
	eSvc "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/event"
	inSvc "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/inventory"
	pSvc "github.com/vogiaan1904/ticketbottle-order/pkg/grpc/payment"
	pkgLog "github.com/vogiaan1904/ticketbottle-order/pkg/logger"
	pkgTemporal "github.com/vogiaan1904/ticketbottle-order/pkg/temporal"
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
	inSvcCli, inClose, err := inSvc.NewInventoryClient(cfg.Microservice.Inventory)
	if err != nil {
		l.Fatalf(ctx, "Failed to create inventory service client: %v", err)
		os.Exit(1)
	}
	defer inClose()

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

	// Initialize Kafka consumer group
	kConsGr, err := kafka.NewConsumerGroup(cfg.Kafka)
	if err != nil {
		l.Fatalf(ctx, "Failed to create Kafka consumer group: %v", err)
		os.Exit(1)
	}

	// Initialize producers
	oProd := oProd.NewProducer(kProd, l)

	// Initialize repositories
	db := mCli.Database(cfg.Mongo.Database)
	oRepo := oRepo.New(l, db)

	// Initialize Temporal client
	temporalClient, err := pkgTemporal.NewClient(pkgTemporal.Config{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		l.Fatalf(ctx, "Failed to create Temporal client: %v", err)
		os.Exit(1)
	}
	defer temporalClient.Close()

	// Initialize Temporal worker
	temporalWorker, err := pkgTemporal.NewWorker(temporalClient, cfg.Temporal.TaskQueue, pkgTemporal.WorkerDependencies{
		OrderActivities:     activities.NewOrderActivities(oRepo),
		PaymentActivities:   activities.NewPaymentActivities(pmtSvcCli),
		InventoryActivities: activities.NewInventoryActivities(inSvcCli),
		EventActivities:     activities.NewEventActivities(eSvcCli),
	})
	if err != nil {
		l.Fatalf(ctx, "Failed to create Temporal worker: %v", err)
		os.Exit(1)
	}

	// Start Temporal worker
	go func() {
		l.Infof(ctx, "Starting Temporal worker on task queue: %s", cfg.Temporal.TaskQueue)
		if err := temporalWorker.Run(nil); err != nil {
			l.Fatalf(ctx, "Temporal worker failed: %v", err)
		}
	}()

	// Initialize services
	oSvc := oSvc.New(l, oSvc.Config{
		PaymentTimeoutSeconds: int32(cfg.Server.PaymentTimeoutSeconds),
		TemporalTaskQueue:     cfg.Temporal.TaskQueue,
	}, oRepo, inSvcCli, eSvcCli, pmtSvcCli, oProd, temporalClient)

	// Create consumer
	cons := oCons.NewConsumer(kConsGr, oSvc, l)

	// Start message processors
	if err := cons.Start(ctx); err != nil {
		l.Fatalf(ctx, "Failed to start consumer: %v", err)
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	l.Info(ctx, "Consumer Server shutting down...")

	// Stop Temporal worker
	temporalWorker.Stop()

	cancel()

	if err := cons.Close(); err != nil {
		l.Errorf(ctx, "Error closing consumer: %v", err)
	}

	if err := kProd.Close(); err != nil {
		l.Errorf(ctx, "Error closing Kafka producer: %v", err)
	}

	l.Info(ctx, "Consumer server exited")
}
