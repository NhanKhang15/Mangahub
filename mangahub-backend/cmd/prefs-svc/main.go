package main

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	mongoplat "mangahub-backend/internal/core/database"
	"mangahub-backend/internal/modules/prefs/grpcserver"
	prefsService "mangahub-backend/internal/modules/prefs/service"
	"mangahub-backend/internal/platform/grpcinterceptor"
	prefspb "mangahub-backend/proto/prefspb"
)

func main() {
	port := getenv("GRPC_PORT", "50054")
	mongoURI := getenv("MONGO_URI", "mongodb://localhost:27017")
	mongoDB := getenv("MONGO_DB", "mangahub")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cl, err := mongoplat.Connect(rootCtx, mongoURI, mongoDB)
	if err != nil {
		log.Fatalf("prefs-svc: mongo connect: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = cl.Close(ctx)
	}()

	if err := cl.EnsureIndexes(rootCtx); err != nil {
		log.Printf("prefs-svc: ensure indexes: %v", err)
	}

	repo := prefsService.NewMongoRepo(cl.DB)
	svc := prefsService.NewService(repo)
	server := grpcserver.New(svc)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcinterceptor.Recovery(),
			grpcinterceptor.Logging(),
		),
	)
	prefspb.RegisterUserPreferencesServer(grpcSrv, server)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("prefs-svc: listen :%s: %v", port, err)
	}

	go func() {
		log.Printf("prefs-svc listening on :%s", port)
		if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("prefs-svc: serve: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Println("prefs-svc: shutting down")
	grpcSrv.GracefulStop()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
