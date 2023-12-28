package service

import (
	"context"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"

	l "ucode/ucode_go_auth_service/pkg/logger"

	"github.com/golang/protobuf/ptypes/empty"
)

// AdminService ...
type PingService struct {
	logger l.LoggerI
	pb.UnimplementedAuthPingServiceServer
}

func NewPingService(log l.LoggerI, services ServiceNodesI) *PingService {
	return &PingService{
		logger: log,
	}
}

func (s *PingService) Ping(ctx context.Context, req *empty.Empty) (res *pb.PingResponse, err error) {

	s.logger.Info("--AuthServicePing-- requested")
	
	return &pb.PingResponse{
		Message: "Pong",
	}, nil
}
