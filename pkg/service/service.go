package service

import (
	"context"
	"io"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "local/tages/pkg/proto"
	"local/tages/pkg/storage"
)

const chunkSize = 1024

type FileService struct {
	pb.UnimplementedFileServiceServer
	storageService *storage.Storage
	uploadsSem     chan struct{}
	listsSem       chan struct{}
	mu             sync.Mutex
}

func NewFileService(storageService *storage.Storage, maxUploads, maxLists int) *FileService {
	return &FileService{
		storageService: storageService,
		uploadsSem:     make(chan struct{}, maxUploads),
		listsSem:       make(chan struct{}, maxLists),
	}
}

func (s *FileService) UploadFile(stream pb.FileService_UploadFileServer) error {
	s.uploadsSem <- struct{}{}
	defer func() { <-s.uploadsSem }()

	var filename string
	var file *os.File

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "Failed to receive data: %v", err)
		}

		if req.GetFilename() != "" {
			filename = req.GetFilename()
			file, err = s.storageService.CreateFile(filename)
			if err != nil {
				return status.Errorf(codes.Internal, "Failed to create file: %v", err)
			}
		} else if req.GetChunkData() != nil {
			if file == nil {
				return status.Errorf(codes.InvalidArgument, "Filename not provided before chunk data")
			}
			_, err = file.Write(req.GetChunkData())
			if err != nil {
				return status.Errorf(codes.Internal, "Failed to write file: %v", err)
			}
		}
	}

	if file != nil {
		file.Close()
	}

	return stream.SendAndClose(&pb.UploadResponse{Success: true, Message: "File uploaded successfully"})
}

func (s *FileService) DownloadFile(req *pb.DownloadRequest, stream pb.FileService_DownloadFileServer) error {
	s.uploadsSem <- struct{}{}
	defer func() { <-s.uploadsSem }()

	file, err := s.storageService.ReadFile(req.GetFilename())
	if err != nil {
		return status.Errorf(codes.NotFound, "File not found: %v", err)
	}
	defer file.Close()

	buffer := make([]byte, chunkSize)
	for {
		bytesRead, err := file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "Failed to read file: %v", err)
		}

		if err := stream.Send(&pb.DownloadResponse{ChunkData: buffer[:bytesRead]}); err != nil {
			return status.Errorf(codes.Unknown, "Failed to send chunk: %v", err)
		}
	}

	return nil
}

func (s *FileService) ListFiles(ctx context.Context, req *pb.ListFilesRequest) (*pb.ListFilesResponse, error) {
	s.listsSem <- struct{}{}
	defer func() { <-s.listsSem }()

	files, err := s.storageService.ListFiles()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to list files: %v", err)
	}

	var fileInfos []*pb.FileInfo
	for _, file := range files {
		info, err := file.Info()
		if err != nil {
			continue
		}

		fileInfos = append(fileInfos, &pb.FileInfo{
			Filename:  file.Name(),
			CreatedAt: info.ModTime().Format(time.RFC3339),
			UpdatedAt: info.ModTime().Format(time.RFC3339),
		})
	}

	return &pb.ListFilesResponse{Files: fileInfos}, nil
}
