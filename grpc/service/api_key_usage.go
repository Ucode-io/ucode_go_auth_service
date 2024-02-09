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
	"google.golang.org/protobuf/types/known/emptypb"
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

func (s *apiKeyUsageService) CheckLimit(ctx context.Context, req *pb.CheckLimitRequest) (*pb.CheckLimitResponse, error) {
	s.log.Info("---CheckLimitApiKeyUsage--->", logger.Any("req", req))

	res, err := s.strg.ApiKeyUsage().CheckLimit(ctx, req)
	if err != nil {
		s.log.Error("!!!CheckLimitApiKeyUsage--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "error on creating new api key")
	}

	if res.RpsCount <= 0 {
		res.IsLimitReached = true
		return res, status.Error(codes.ResourceExhausted, "rps limit exceeded")
	}

	if res.IsLimitReached {
		return res, status.Error(codes.ResourceExhausted, "monthly limit exceeded")
	}

	return res, nil
}

func (s *apiKeyUsageService) Get(ctx context.Context, req *pb.GetApiKeyUsageReq) (*pb.ApiKeyUsage, error) {
	s.log.Info("---GetApiKeyUsage--->", logger.Any("req", req))

	res, err := s.strg.ApiKeyUsage().Get(ctx, req)
	if err == pgx.ErrNoRows {
		s.log.Error("!!!GetApiKeyUsage--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, "api-key not found")
	}
	if err != nil {
		s.log.Error("!!!GetApiKeyUsage--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "error on getting api key")
	}

	return res, nil
}

func (s *apiKeyUsageService) Create(ctx context.Context, req *pb.ApiKeyUsage) (*emptypb.Empty, error) {
	s.log.Info("---UpsertApiKeyUsage--->", logger.Any("req", req))

	res := &emptypb.Empty{}

	err := s.strg.ApiKeyUsage().Create(ctx, req)
	if err == pgx.ErrNoRows {
		s.log.Error("!!!UpsertApiKeyUsage--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, "api-key not found")
	}
	if err != nil {
		s.log.Error("!!!UpsertApiKeyUsage--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "error on upserting api key usage")
	}

	return res, nil
}
