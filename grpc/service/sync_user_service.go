package service

import (
	"context"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"
)

type syncUserService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	pb.UnimplementedSyncUserServiceServer
}

func NewSyncUserService(cfg config.Config, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI) *syncUserService {
	return &syncUserService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (sus *syncUserService) SyncUserWithAuth(ctx context.Context, req *pb.SyncUserWithAuthRequest) (*pb.SyncUserWithAuthResponse, error) {
	var response = pb.SyncUserWithAuthResponse{}
	var username string
	username = req.GetLogin()
	if username == "" {
		username = req.GetEmail()
	}
	if username == "" {
		username = req.GetPhone()
	}

	user, err := sus.strg.User().GetByUsername(context.Background(), username)
	if err != nil {
		return nil, err
	}
	userId := user.GetId()
	if userId == "" {
		// if user not found in auth service db we have to create it
		sus.services.UserService().V2CreateUser(context.Background(), &pb.CreateUserRequest{})
		
	}

	return &response, nil
}
