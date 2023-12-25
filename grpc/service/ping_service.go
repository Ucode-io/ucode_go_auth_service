package service

import (
	"context"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"

	"github.com/golang/protobuf/ptypes/empty"
)

type pingService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	pb.UnimplementedPingServiceServer
}

func NewPingService(cfg config.Config, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI) *pingService {
	return &pingService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (s *pingService) Ping(ctx context.Context, req *empty.Empty) (res *pb.PingResponse, err error) {
	return &pb.PingResponse{
		Message: "Pong",
	}, nil
}
