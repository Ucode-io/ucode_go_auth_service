package service

import (
	"context"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/storage"

	"github.com/jackc/pgx/v4"
	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type apiKeyUsageService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedApiKeyUsageServiceServer
}

func NewApiKeyUsageService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *apiKeyUsageService {
	return &apiKeyUsageService{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
	}
}

func (s *apiKeyUsageService) CheckAndUpsertLimit(ctx context.Context, req *pb.CheckLimitRequest) (*pb.CheckLimitResponse, error) {
	s.log.Info("---CheckAndUpsertLimit--->", logger.Any("req", req))

	res, err := s.strg.ApiKeyUsage().CheckAndUpsertLimit(ctx, req)
	if err != nil {
		s.log.Error("!!!CheckAndUpsertLimit--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "error on creating new api key")
	}

	return res, nil
}

func (s *apiKeyUsageService) Get(ctx context.Context, req *pb.GetApiKeyUsageReq) (*pb.ApiKeyUsage, error) {
	s.log.Info("---Get--->", logger.Any("req", req))

	res, err := s.strg.ApiKeyUsage().Get(ctx, req)
	if err == pgx.ErrNoRows {
		s.log.Error("!!!Get--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, "api-key not found")
	}
	if err != nil {
		s.log.Error("!!!Get--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "error on getting api key")
	}

	return res, nil
}
