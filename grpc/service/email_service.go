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
	"google.golang.org/protobuf/types/known/emptypb"
)

type emailService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedEmailOtpServiceServer
}

func NewEmailService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *emailService {
	return &emailService{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
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

func (e *emailService) CreateEmailSettings(ctx context.Context, req *pb.EmailSettings) (*pb.EmailSettings, error) {
	e.log.Info("---EmailService.CreateEmailSettings--->", logger.Any("req", req))

	res, err := e.strg.Email().CreateEmailSettings(
		ctx,
		req,
	)
	if err != nil {
		e.log.Error("!!!EmailService.CreateEmailSettings--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}

func (e *emailService) UpdateEmailSettings(ctx context.Context, req *pb.UpdateEmailSettingsRequest) (*pb.EmailSettings, error) {
	e.log.Info("---EmailService.UpdateEmailSettings--->", logger.Any("req", req))

	res, err := e.strg.Email().UpdateEmailSettings(
		ctx,
		req,
	)

	if err != nil {
		e.log.Error("!!!EmailService.CreateEmailSettings--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}

func (e *emailService) GetListEmailSettings(ctx context.Context, req *pb.GetListEmailSettingsRequest) (*pb.UpdateEmailSettingsResponse, error) {
	e.log.Info("---EmailService.GetListEmailSettings--->", logger.Any("req", req))

	res, err := e.strg.Email().GetListEmailSettings(
		ctx,
		req,
	)
	if err != nil {
		e.log.Error("!!!EmailService.GetEmailSettings--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}

func (e *emailService) DeleteEmailSettings(ctx context.Context, req *pb.EmailSettingsPrimaryKey) (*emptypb.Empty, error) {
	e.log.Info("---EmailService.DeleteEmailSettings--->", logger.Any("req", req))

	res, err := e.strg.Email().DeleteEmailSettings(
		ctx,
		req,
	)

	if err != nil {
		e.log.Error("!!!EmailService.CreateEmailSettings--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}
