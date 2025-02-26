package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	pb "local/tages/pkg/proto"
	"log"
	"os"
	"path/filepath"
)

const (
	address   = "localhost:50051"
	chunkSize = 1024
)

var (
	action   = flag.String("action", "", "Action to perform: upload, download, list")
	filename = flag.String("filename", "", "Filename for upload or download")
)

func uploadFile(client pb.FileServiceClient, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	stream, err := client.UploadFile(context.Background())
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	// Отправляем имя файла
	if err := stream.Send(&pb.UploadRequest{
		Data: &pb.UploadRequest_Filename{
			Filename: filepath.Base(filename),
		},
	}); err != nil {
		log.Fatalf("Failed to send filename: %v", err)
	}

	buffer := make([]byte, chunkSize)
	for {
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}

		// Отправляем данные файла
		if err := stream.Send(&pb.UploadRequest{
			Data: &pb.UploadRequest_ChunkData{
				ChunkData: buffer[:bytesRead],
			},
		}); err != nil {
			log.Fatalf("Failed to send chunk: %v", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Failed to receive response: %v", err)
	}

	fmt.Println(resp.Message)
}

func downloadFile(client pb.FileServiceClient, filename string) {
	req := &pb.DownloadRequest{
		Filename: filename,
	}
	stream, err := client.DownloadFile(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Failed to receive chunk: %v", err)
		}

		if _, err := file.Write(resp.GetChunkData()); err != nil {
			log.Fatalf("Failed to write chunk: %v", err)
		}
	}

	fmt.Printf("File downloaded successfully\n")
}

func listFiles(client pb.FileServiceClient) {
	req := &pb.ListFilesRequest{}
	resp, err := client.ListFiles(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to list files: %v", err)
	}

	for _, fileInfo := range resp.GetFiles() {
		fmt.Printf("%s | %s | %s\n", fileInfo.GetFilename(), fileInfo.GetCreatedAt(), fileInfo.GetUpdatedAt())
	}
}

func main() {
	flag.Parse()

	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewFileServiceClient(conn)

	switch *action {
	case "upload":
		if *filename == "" {
			log.Fatal("Filename is required for upload")
		}
		uploadFile(client, *filename)
	case "download":
		if *filename == "" {
			log.Fatal("Filename is required for download")
		}
		downloadFile(client, *filename)
	case "list":
		listFiles(client)
	default:
		log.Fatal("Invalid action. Use 'upload', 'download', or 'list'")
	}
}
