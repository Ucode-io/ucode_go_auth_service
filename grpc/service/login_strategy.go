package service

import (
	"context"
	"runtime"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/saidamir98/udevs_pkg/logger"
)

type loginStrategyService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedLoginStrategyServiceServer
}

func NewLoginStrategyService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *loginStrategyService {
	return &loginStrategyService{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
	}
}

func (ls *loginStrategyService) GetList(ctx context.Context, req *pb.GetListRequest) (*pb.GetListResponse, error) {
	ls.log.Info("GetList: ", logger.Any("request: --> ", req))

	res, err := ls.strg.LoginStrategy().GetList(ctx, req)
	if err != nil {
		ls.log.Error("! GetList:", logger.Error(err))
		return nil, err
	}

	return res, nil
}

func (ls *loginStrategyService) GetByID(ctx context.Context, req *pb.LoginStrategyPrimaryKey) (*pb.LoginStrategy, error) {
	ls.log.Info("GetByID: ", logger.Any("request: --> ", req))

	res, err := ls.strg.LoginStrategy().GetByID(ctx, req)
	if err != nil {
		ls.log.Error("! GetByID:", logger.Error(err))
		return nil, err
	}

	return res, nil
}

func (ls *loginStrategyService) Upsert(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	ls.log.Info("Upsert: ", logger.Any("request: --> ", req))

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		ls.log.Info("Memory used by the UpsertLoginStrategy", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			ls.log.Info("Memory used over 300 mb", logger.Any("UpsertLoginStrategy", memoryUsed))
		}
	}()

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
