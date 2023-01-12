package service

import (
	"context"
	"github.com/google/uuid"
	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"
)

type apiKeysService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	pb.UnimplementedApiKeysServer
}

func NewApiKeysService(cfg config.Config, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI) *apiKeysService {
	return &apiKeysService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (s *apiKeysService) Create(ctx context.Context, req *pb.CreateReq) (*pb.CreateRes, error) {
	s.log.Info("---Create--->", logger.Any("req", req))
	id, err := uuid.NewUUID()
	if err != nil {
		s.log.Error("!!!Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	req.Id = id.String()

	secretKey := helper.GenerateSecretKey(32)

	req.AppSecret = secretKey

	res, err := s.strg.ApiKeys().Create(ctx, req)
	if err != nil {
		s.log.Error("!!!Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "error on creating new api key")
	}

	return res, nil
}

func (s *apiKeysService) Update(ctx context.Context, req *pb.UpdateReq) (*pb.UpdateRes, error) {
	s.log.Info("---Update--->", logger.Any("req", req))

	res, err := s.strg.ApiKeys().Update(ctx, req)
	if err != nil {
		s.log.Error("!!!Update--->", logger.Error(err))
		return &pb.UpdateRes{
			RowEffected: int32(res),
		}, status.Error(codes.Internal, "error on updating new api key")
	}

	return &pb.UpdateRes{
		RowEffected: int32(res),
	}, nil
}

func (s *apiKeysService) Get(ctx context.Context, req *pb.GetReq) (*pb.GetRes, error) {
	s.log.Info("---Get--->", logger.Any("req", req))

	res, err := s.strg.ApiKeys().Get(ctx, req)
	if err != nil {
		s.log.Error("!!!Get--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "error on getting api key")
	}

	return res, nil
}

func (s *apiKeysService) GetList(ctx context.Context, req *pb.GetListReq) (*pb.GetListRes, error) {
	s.log.Info("---GetList--->", logger.Any("req", req))

	res, err := s.strg.ApiKeys().GetList(ctx, req)
	if err != nil {
		s.log.Error("!!!GetList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, "error on getting api keys")
	}

	return res, nil
}

func (s *apiKeysService) Delete(ctx context.Context, req *pb.DeleteReq) (*pb.DeleteRes, error) {
	s.log.Info("---GetList--->", logger.Any("req", req))

	res, err := s.strg.ApiKeys().Delete(ctx, req)
	if err != nil {
		s.log.Error("!!!GetList--->", logger.Error(err))
		return &pb.DeleteRes{
			RowEffected: int32(res),
		}, status.Error(codes.Internal, "error on deleting api keys")
	}

	return &pb.DeleteRes{
		RowEffected: int32(res),
	}, nil
}
