package service

import (
	"context"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type emailService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	pb.UnimplementedEmailOtpServiceServer
}

func NewEmailService(cfg config.Config, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI) *emailService {
	return &emailService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (e *emailService) Create(ctx context.Context, req *pb.Email) (*pb.Email, error) {
	e.log.Info("---EmailService.Create--->", logger.Any("req", req))

	res, err := e.strg.Email().Create(ctx, req)
	if err != nil {
		e.log.Error("!!!EmailService.Create--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}

func (e *emailService) GetEmailByID(ctx context.Context, req *pb.EmailOtpPrimaryKey) (*pb.Email, error) {
	e.log.Info("---EmailService.GetEmailByID--->", logger.Any("req", req))

	res, err := e.strg.Email().GetByPK(ctx, req)
	if err != nil {
		e.log.Error("!!!EmailService.GetEmailByID--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}
