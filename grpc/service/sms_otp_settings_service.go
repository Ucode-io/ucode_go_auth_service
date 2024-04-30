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

type smsOtpSettingsService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedSmsOtpSettingsServiceServer
}

func NewSmsOtpSettingsService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *smsOtpSettingsService {
	return &smsOtpSettingsService{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
	}
}

func (s *smsOtpSettingsService) Create(ctx context.Context, req *pb.CreateSmsOtpSettingsRequest) (*pb.SmsOtpSettings, error) {
	res, err := s.strg.SmsOtpSettings().Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (s *smsOtpSettingsService) Update(ctx context.Context, req *pb.SmsOtpSettings) (*pb.SmsOtpSettings, error) {
	rowsAffected, err := s.strg.SmsOtpSettings().Update(ctx, req)
	if err != nil {
		return nil, err
	}
	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, "no rows were affected")
	}
	res, err := s.strg.SmsOtpSettings().GetById(ctx, &pb.SmsOtpSettingsPrimaryKey{
		Id: req.Id,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (s *smsOtpSettingsService) GetById(ctx context.Context, req *pb.SmsOtpSettingsPrimaryKey) (*pb.SmsOtpSettings, error) {
	res, err := s.strg.SmsOtpSettings().GetById(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (s *smsOtpSettingsService) GetList(ctx context.Context, req *pb.GetListSmsOtpSettingsRequest) (*pb.SmsOtpSettingsResponse, error) {
	res, err := s.strg.SmsOtpSettings().GetList(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func (s *smsOtpSettingsService) Delete(ctx context.Context, req *pb.SmsOtpSettingsPrimaryKey) (*emptypb.Empty, error) {
	rowsAffected, err := s.strg.SmsOtpSettings().Delete(ctx, req)
	if err != nil {
		return nil, err
	}
	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &emptypb.Empty{}, nil
}
