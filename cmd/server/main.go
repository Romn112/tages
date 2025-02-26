package main

import (
	"google.golang.org/grpc"
	pb "local/tages/pkg/proto"
	"local/tages/pkg/service"
	"local/tages/pkg/storage"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const (
	port       = ":50051"
	uploadDir  = "./uploads"
	maxUploads = 10
	maxLists   = 100
)

func main() {
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	defer lis.Close()

	s := grpc.NewServer()
	storageService := storage.NewStorage(uploadDir)
	srv := service.NewFileService(storageService, maxUploads, maxLists)
	pb.RegisterFileServiceServer(s, srv)

	log.Printf("Server listening at %v", lis.Addr())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down the server...")
	s.GracefulStop()
	log.Println("Server stopped")
}
