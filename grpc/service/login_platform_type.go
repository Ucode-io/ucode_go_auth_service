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

type loginPlatformType struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedLoginPlatformTypeLoginServiceServer
}

func NewLoginPlatformTypeService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *loginPlatformType {
	return &loginPlatformType{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
	}
}

func (e *loginPlatformType) CreateLoginPlatformType(ctx context.Context, req *pb.LoginPlatform) (*pb.LoginPlatform, error) {
	e.log.Info("---LoginPlatformType.CreateLoginPlatformType--->", logger.Any("req", req))
	res, err := e.strg.LoginPlatformType().CreateLoginPlatformType(
		ctx,
		req,
	)
	if err != nil {
		e.log.Error("!!!---LoginPlatformType.CreateLoginPlatformType--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}

func (e *loginPlatformType) UpdateLoginPlatformType(ctx context.Context, req *pb.UpdateLoginPlatformTypeRequest) (*pb.LoginPlatform, error) {
	e.log.Info("---LoginPlatformType.UpdateLoginPlatformType--->", logger.Any("req", req))

	types, err := e.strg.LoginPlatformType().GetLoginPlatformType(
		ctx,
		&pb.LoginPlatformTypePrimaryKey{
			Id: req.Id,
		},
	)

	id, err := e.strg.LoginPlatformType().UpdateLoginPlatformType(
		ctx,
		req,
		types.Type,
	)

	if err != nil {
		e.log.Error("!!!---LoginPlatformType.UpdateLoginPlatformType--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := e.strg.LoginPlatformType().GetLoginPlatformType(
		ctx,
		&pb.LoginPlatformTypePrimaryKey{
			Id: id,
		},
	)

	if err != nil {
		e.log.Error("!!!---LoginPlatformType.UpdateLoginPlatformType--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}

func (e *loginPlatformType) GetListLoginPlatformType(ctx context.Context, req *pb.GetListLoginPlatformTypeRequest) (*pb.GetListLoginPlatformTypeResponse, error) {
	e.log.Info("---LoginPlatformType.GetListLoginPlatformType--->", logger.Any("req", req))

	res, err := e.strg.LoginPlatformType().GetListLoginPlatformType(
		ctx,
		req,
	)
	if err != nil {
		e.log.Error("!!!---LoginPlatformType.GetListLoginPlatformType--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}

func (e *loginPlatformType) GetLoginPlatformType(ctx context.Context, req *pb.LoginPlatformTypePrimaryKey) (*pb.LoginPlatform, error) {
	e.log.Info("---LoginPlatformType.GetLoginPlatformTypeByPK--->", logger.Any("req", req))

	res, err := e.strg.LoginPlatformType().GetLoginPlatformType(
		ctx,
		req,
	)

	if err != nil {
		e.log.Error("!!!---LoginPlatformType.GetLoginPlatformTypeByPK--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}

func (e *loginPlatformType) DeleteLoginPlatformType(ctx context.Context, req *pb.LoginPlatformTypePrimaryKey) (*emptypb.Empty, error) {
	e.log.Info("---LoginPlatformTypeService.DeleteLoginPlatformType--->", logger.Any("req", req))

	res, err := e.strg.LoginPlatformType().DeleteLoginSettings(
		ctx,
		req,
	)

	if err != nil {
		e.log.Error("!!!LoginPlatformTypeService.DeleteLoginId--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return res, nil
}
