package service

import (
	"context"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/saidamir98/udevs_pkg/logger"
)

type loginStrategyService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	pb.UnimplementedLoginStrategyServiceServer
}

func NewLoginStrategyService(cfg config.Config, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI) *loginStrategyService {
	return &loginStrategyService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (ls *loginStrategyService) GetList(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	ls.log.Info("GetList: ", logger.Any("request: --> ", req))
	res, err := ls.strg.LoginStrategy().GetList(ctx, req)
	if err != nil {
		ls.log.Error("! GetList:", logger.Error(err))
		return nil, err
	}
	ls.log.Info("GetList: ", logger.Any("response: <-- ", res))
	return res, nil
}

func (ls *loginStrategyService) GetByID(ctx context.Context, req *pb.LoginStrategyPrimaryKey) (*pb.LoginStrategy, error) {
	ls.log.Info("GetByID: ", logger.Any("request: --> ", req))
	res, err := ls.strg.LoginStrategy().GetByID(ctx, req)
	if err != nil {
		ls.log.Error("! GetByID:", logger.Error(err))
		return nil, err
	}
	ls.log.Info("GetByID: ", logger.Any("response: <-- ", res))
	return res, nil
}

func (ls *loginStrategyService) Upsert(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	ls.log.Info("Upsert: ", logger.Any("request: --> ", req))
	for _, value := range req.LoginStrategies {
		if value.Id == "" {
			uuid, err := uuid.NewRandom()
			if err != nil {
				ls.log.Error("! error while generating uuid:", logger.Error(err))
				return nil, err
			}
			value.Id = uuid.String()
		}
	}
	res, err := ls.strg.LoginStrategy().Upsert(ctx, req)
	if err != nil {
		ls.log.Error("! Upsert:", logger.Error(err))
		return nil, err
	}
	return res, nil
}
