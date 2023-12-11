package service

import (
	"context"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/saidamir98/udevs_pkg/security"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type companyService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedCompanyServiceServer
}

func NewCompanyService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *companyService {
	return &companyService{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
	}
}

func (s *companyService) Register(ctx context.Context, req *pb.RegisterCompanyRequest) (*pb.CompanyPrimaryKey, error) {

	//@TODO:: refactor later
	tempOwnerId, err := uuid.NewRandom()
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	companyPKey, err := s.services.CompanyServiceClient().Create(ctx, &company_service.CreateCompanyRequest{
		Title:       req.Name,
		Logo:        "",
		Description: "",
		OwnerId:     tempOwnerId.String(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	project, err := s.services.ProjectServiceClient().Create(ctx, &company_service.CreateProjectRequest{
		CompanyId:    companyPKey.GetId(),
		K8SNamespace: "cp-region-type-id",
		Title:        req.GetName(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	environment, err := s.services.EnvironmentService().Create(
		ctx,
		&company_service.CreateEnvironmentRequest{
			ProjectId:    project.ProjectId,
			Name:         "Production",
			DisplayColor: "#00FF00",
			Description:  "Production Environment",
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	hashedPassword, err := security.HashPassword(req.GetUserInfo().GetPassword())
	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	createUserRes, err := s.strg.User().Create(
		ctx,
		&pb.CreateUserRequest{
			Phone:     req.GetUserInfo().GetPhone(),
			Email:     req.GetUserInfo().GetEmail(),
			Login:     req.GetUserInfo().GetLogin(),
			Password:  hashedPassword,
			Active:    1, //@TODO:: user must verify himself
			PhotoUrl:  "",
			CompanyId: companyPKey.GetId(),
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	_, err = s.services.CompanyServiceClient().Update(ctx, &company_service.Company{
		Id:          companyPKey.Id,
		Name:        req.Name,
		Logo:        "",
		Description: "",
		OwnerId:     createUserRes.GetId(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	_, err = s.services.UserService().AddUserToProject(ctx, &pb.AddUserToProjectReq{
		CompanyId: companyPKey.GetId(),
		ProjectId: project.GetProjectId(),
		UserId:    createUserRes.GetId(),
		EnvId:     environment.GetId(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	// resource and settings service resource

	resource, err := s.services.ResourceService().CreateResource(
		ctx,
		&company_service.CreateResourceReq{
			CompanyId:     companyPKey.GetId(),
			EnvironmentId: environment.GetId(),
			ProjectId:     project.GetProjectId(),
			Resource: &company_service.Resource{
				ResourceType: 1,
				NodeType:     config.LOW_NODE_TYPE,
			},
			UserId: createUserRes.GetId(),
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany-AutoCreateResource--->", logger.Error(err))
		return nil, err
	}

	_, err = s.services.ServiceResource().Update(
		ctx,
		&company_service.UpdateServiceResourceReq{
			EnvironmentId:    environment.GetId(),
			ProjectId:        project.GetProjectId(),
			ServiceResources: helper.MakeBodyServiceResource(resource.GetId()),
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany-AutoSettingMicroServices--->", logger.Error(err))
		return nil, err
	}

	return &pb.CompanyPrimaryKey{Id: companyPKey.GetId()}, nil
}

func (s *companyService) Update(ctx context.Context, req *pb.UpdateCompanyRequest) (*emptypb.Empty, error) {
	_, err := s.strg.Company().Update(ctx, req)
	if err != nil {
		s.log.Error("---UpdateCompany--->", logger.Error(err))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *companyService) Remove(ctx context.Context, req *pb.CompanyPrimaryKey) (*emptypb.Empty, error) {
	_, err := s.strg.Company().Remove(ctx, req)
	if err != nil {
		s.log.Error("---RemoveCompany--->", logger.Error(err))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *companyService) GetList(ctx context.Context, req *pb.GetComapnyListRequest) (*pb.GetListCompanyResponse, error) {
	resp, err := s.strg.Company().GetList(ctx, req)
	if err != nil {
		s.log.Error("---RemoveCompany--->", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (s *companyService) GetByID(ctx context.Context, pKey *pb.CompanyPrimaryKey) (*pb.Company, error) {
	resp, err := s.strg.Company().GetByID(ctx, pKey)
	if err != nil {
		s.log.Error("---RemoveCompany--->", logger.Error(err))
		return nil, err
	}

	return resp, nil
}
