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
	"google.golang.org/grpc/credentials/insecure"

	mongoplat "mangahub-backend/internal/core/database"
	"mangahub-backend/internal/modules/artist/grpcserver"
	artistService "mangahub-backend/internal/modules/artist/service"
	"mangahub-backend/internal/platform/grpcinterceptor"
	artistpb "mangahub-backend/proto/artistpb"
	catalogpb "mangahub-backend/proto/catalogpb"
)

func main() {
	port := getenv("GRPC_PORT", "50052")
	mongoURI := getenv("MONGO_URI", "mongodb://localhost:27017")
	mongoDB := getenv("MONGO_DB", "mangahub")
	catalogAddr := getenv("CATALOG_GRPC_ADDR", "localhost:50051")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cl, err := mongoplat.Connect(rootCtx, mongoURI, mongoDB)
	if err != nil {
		log.Fatalf("artist-svc: mongo connect: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = cl.Close(ctx)
	}()

	if err := cl.EnsureIndexes(rootCtx); err != nil {
		log.Printf("artist-svc: ensure indexes: %v", err)
	}

	catalogConn, err := grpc.NewClient(catalogAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("artist-svc: dial catalog %s: %v", catalogAddr, err)
	}
	defer func() { _ = catalogConn.Close() }()
	catalogClient := catalogpb.NewMangaCatalogClient(catalogConn)

	repo := artistService.NewMongoRepo(cl.DB)
	svc := artistService.NewService(repo)
	server := grpcserver.New(svc, catalogClient)

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcinterceptor.Recovery(),
			grpcinterceptor.Logging(),
		),
	)
	artistpb.RegisterArtistServer(grpcSrv, server)

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("artist-svc: listen :%s: %v", port, err)
	}

	go func() {
		log.Printf("artist-svc listening on :%s (catalog=%s)", port, catalogAddr)
		if err := grpcSrv.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("artist-svc: serve: %v", err)
		}
	}()

	<-rootCtx.Done()
	log.Println("artist-svc: shutting down")
	grpcSrv.GracefulStop()
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
