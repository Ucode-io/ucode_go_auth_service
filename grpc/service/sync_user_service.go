package service

import (
	"context"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"
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

func (sus *syncUserService) CreateUser(ctx context.Context, req *pb.CreateSyncUserRequest) (*pb.SyncUserResponse, error) {
	var (
		response = pb.SyncUserResponse{}
		user     *pb.User
	)
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
	project, err := sus.services.ProjectService().GetByPK(context.Background(), &pb.ProjectPrimaryKey{
		Id: req.GetProjectId(),
	})
	if err != nil {
		return nil, err
	}
	resEnv, err := sus.services.ResourceService().GetResourceByResEnvironId(context.Background(), &pbCompany.GetResourceRequest{
		Id: req.GetResourceEnvironmentId(),
	})
	if err != nil {
		return nil, err
	}

	if userId == "" {
		// if user not found in auth service db we have to create it
		user, err = sus.services.UserService().V2CreateUser(context.Background(), &pb.CreateUserRequest{
			Login:                 req.GetLogin(),
			Password:              req.GetPassword(),
			Phone:                 req.GetPhone(),
			Email:                 req.GetEmail(),
			ResourceEnvironmentId: req.GetResourceEnvironmentId(),
			ProjectId:             req.GetProjectId(),
			RoleId:                req.GetRoleId(),
			ClientTypeId:          req.GetClientTypeId(),
			CompanyId:             project.GetCompanyId(),
			ResourceType:          int32(resEnv.GetResourceType()),
		})
		if err != nil {
			return nil, err
		}
	} else {
		exists, err := sus.strg.User().GetUserProjectByAllFields(context.Background(), models.GetUserProjectByAllFieldsReq{
			ClientTypeId: req.GetClientTypeId(),
			RoleId:       req.GetRoleId(),
			UserId:       userId,
			CompanyId:    project.GetCompanyId(),
			ProjectId:    req.GetProjectId(),
		})
		if err != nil {
			return nil, err
		}
		if !exists {
			_, err = sus.strg.User().AddUserToProject(context.Background(), &pb.AddUserToProjectReq{
				UserId:       userId,
				CompanyId:    project.GetCompanyId(),
				RoleId:       req.GetRoleId(),
				ProjectId:    req.GetProjectId(),
				ClientTypeId: req.GetClientTypeId(),
			})
			if err != nil {
				return nil, err
			}
		}
	}
	response.UserId = user.GetId()
	return &response, nil
}
