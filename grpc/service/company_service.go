package service

import (
	"context"
	"runtime"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/security"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/saidamir98/udevs_pkg/logger"
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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_company.Register")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the CompanyRegister", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("CompanyRegister", memoryUsed))
		}
	}()
	//@TODO:: refactor later
	tempOwnerId, err := uuid.NewRandom()
	if err != nil {
		s.log.Error("---RegisterCompany-->NewUUID", logger.Error(err))
		return nil, err
	}

	companyPKey, err := s.services.CompanyServiceClient().Create(ctx, &company_service.CreateCompanyRequest{
		Title:       req.Name,
		Logo:        "",
		Description: "",
		OwnerId:     tempOwnerId.String(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->CreateCompanyServiceClient", logger.Error(err))
		return nil, err
	}

	project, err := s.services.ProjectServiceClient().Create(ctx, &company_service.CreateProjectRequest{
		CompanyId:    companyPKey.GetId(),
		K8SNamespace: "cp-region-type-id",
		Title:        req.GetName(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->CreateProject", logger.Error(err))
		return nil, err
	}

	_, _ = s.services.ProjectServiceClient().Update(
		ctx,
		&company_service.Project{
			CompanyId:    companyPKey.GetId(),
			K8SNamespace: "cp-region-type-id",
			ProjectId:    project.GetProjectId(),
			Title:        req.Name,
			Language: []*company_service.Language{{
				Id:         "e2d68f08-8587-4136-8cd4-c26bf1b9cda1",
				Name:       "English",
				NativeName: "English",
				ShortName:  "en",
			}},
		},
	)

	environment, err := s.services.EnvironmentService().Create(ctx,
		&company_service.CreateEnvironmentRequest{
			ProjectId:    project.ProjectId,
			Name:         "Production",
			DisplayColor: "#00FF00",
			Description:  "Production Environment",
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany-->CreateEnvironment", logger.Error(err))
		return nil, err
	}

	hashedPassword, err := security.HashPasswordBcrypt(req.GetUserInfo().GetPassword())
	if err != nil {
		s.log.Error("!!!RegisterCompany-->HashPassword", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	createUserRes, err := s.strg.User().Create(ctx,
		&pb.CreateUserRequest{
			Phone:     req.GetUserInfo().GetPhone(),
			Email:     req.GetUserInfo().GetEmail(),
			Login:     req.GetUserInfo().GetLogin(),
			Password:  hashedPassword,
			Active:    1,
			PhotoUrl:  "",
			CompanyId: companyPKey.GetId(),
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->CreateUser", logger.Error(err))
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
		s.log.Error("---RegisterCompany-->UpdateCompany", logger.Error(err))
		return nil, err
	}

	_, err = s.services.UserService().AddUserToProject(ctx, &pb.AddUserToProjectReq{
		CompanyId: companyPKey.GetId(),
		ProjectId: project.GetProjectId(),
		UserId:    createUserRes.GetId(),
		EnvId:     environment.GetId(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany-->AddUser2Project", logger.Error(err))
		return nil, err
	}

	resource, err := s.services.ResourceService().CreateResource(ctx,
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

	_, err = s.services.ServiceResource().Update(ctx,
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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_company.Update")
	defer dbSpan.Finish()

	_, err := s.strg.Company().Update(ctx, req)
	if err != nil {
		s.log.Error("---UpdateCompany--->", logger.Error(err))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *companyService) Remove(ctx context.Context, req *pb.CompanyPrimaryKey) (*emptypb.Empty, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_company.Remove")
	defer dbSpan.Finish()

	_, err := s.strg.Company().Remove(ctx, req)
	if err != nil {
		s.log.Error("---RemoveCompany--->", logger.Error(err))
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *companyService) GetList(ctx context.Context, req *pb.GetComapnyListRequest) (*pb.GetListCompanyResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_company.GetList")
	defer dbSpan.Finish()

	resp, err := s.strg.Company().GetList(ctx, req)
	if err != nil {
		s.log.Error("---RemoveCompany--->", logger.Error(err))
		return nil, err
	}

	return resp, nil
}

func (s *companyService) GetByID(ctx context.Context, pKey *pb.CompanyPrimaryKey) (*pb.Company, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_company.GetByID")
	defer dbSpan.Finish()

	resp, err := s.strg.Company().GetByID(ctx, pKey)
	if err != nil {
		s.log.Error("---RemoveCompany--->", logger.Error(err))
		return nil, err
	}

	return resp, nil
}
