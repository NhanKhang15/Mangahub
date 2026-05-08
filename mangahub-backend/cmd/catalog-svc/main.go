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
	"mangahub-backend/internal/modules/manga/grpcserver"
	mangaService "mangahub-backend/internal/modules/manga/service"
	"mangahub-backend/internal/platform/grpcinterceptor"
	catalogpb "mangahub-backend/proto/catalogpb"
)

func main() {
	port := getenv("GRPC_PORT", "50051")
	mongoURI := getenv("MONGO_URI", "mongodb://localhost:27017")
	mongoDB := getenv("MONGO_DB", "mangahub")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cl, err := mongoplat.Connect(rootCtx, mongoURI, mongoDB)
	if err != nil {
		log.Fatalf("catalog-svc: mongo connect: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = cl.Close(ctx)
	}()

	if err := cl.EnsureIndexes(rootCtx); err != nil {
		log.Printf("catalog-svc: ensure indexes: %v", err)
	}

	repo := mangaService.NewMongoRepo(cl.DB)
	svc := mangaService.NewService(repo)
	server := grpcserver.New(svc)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcinterceptor.Recovery(),
			grpcinterceptor.Logging(),
		),
	)
	catalogpb.RegisterMangaCatalogServer(grpcSrv, server)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("catalog-svc: listen :%s: %v", port, err)
	}

	go func() {
		log.Printf("catalog-svc listening on :%s (mongo=%s db=%s)", port, mongoURI, mongoDB)
		if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("catalog-svc: serve: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Println("catalog-svc: shutting down")
	grpcSrv.GracefulStop()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
