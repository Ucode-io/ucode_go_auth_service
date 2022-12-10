package service

import (
	"context"
	"strings"
	"time"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/saidamir98/udevs_pkg/security"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type companyService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	pb.UnimplementedCompanyServiceServer
}

func NewCompanyService(cfg config.Config, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI) *companyService {
	return &companyService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (s *companyService) Register(ctx context.Context, req *pb.RegisterCompanyRequest) (*pb.CompanyPrimaryKey, error) {

	companyPKey, err := s.strg.Company().Register(ctx, req)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	projectPKey, err := s.strg.Project().Create(ctx, &pb.CreateProjectRequest{
		CompanyId: companyPKey.Id,
		Name:      req.Name,
		Domain:    "test.admin.u-code.io", //@TODO:: get domain
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	clientPlatformPKey, err := s.strg.ClientPlatform().Create(ctx, &pb.CreateClientPlatformRequest{
		ProjectId: projectPKey.Id,
		Name:      strings.ToUpper(req.Name),
		Subdomain: "test.admin.u-code.io", //@TODO:: get subdomain
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	clientTypePKey, err := s.strg.ClientType().Create(ctx, &pb.CreateClientTypeRequest{
		ProjectId:    projectPKey.Id,
		Name:         "ADMIN",
		ConfirmBy:    pb.ConfirmStrategies_UNDECIDED,
		SelfRegister: false,
		SelfRecover:  false,
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	err = s.strg.Client().Add(ctx, projectPKey.Id, &pb.AddClientRequest{
		ClientPlatformId: clientPlatformPKey.Id,
		ClientTypeId:     clientTypePKey.Id,
		LoginStrategy:    pb.LoginStrategies_STANDARD,
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	rolePKey, err := s.strg.Role().Add(ctx, &pb.AddRoleRequest{
		ProjectId:        projectPKey.Id,
		ClientPlatformId: clientPlatformPKey.Id,
		ClientTypeId:     clientTypePKey.Id,
		Name:             "DEFAULT",
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	permissionPkey, err := s.strg.Permission().Create(ctx, &pb.CreatePermissionRequest{
		ClientPlatformId: clientPlatformPKey.Id,
		ParentId:         "ffffffff-ffff-4fff-8fff-ffffffffffff",
		Name:             "/root",
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	_, err = s.strg.RolePermission().Add(ctx, &pb.AddRolePermissionRequest{
		RoleId:       rolePKey.Id,
		PermissionId: permissionPkey.Id,
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	hashedPassword, err := security.HashPassword(req.UserInfo.Password)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	expiresAt := time.Now().Add(time.Hour * 24 * 7).Format(config.DatabaseTimeLayout)

	createUserReq := &pb.CreateUserRequest{
		ProjectId:        projectPKey.Id,
		ClientPlatformId: clientPlatformPKey.GetId(),
		ClientTypeId:     clientTypePKey.GetId(),
		RoleId:           rolePKey.GetId(),
		Phone:            req.UserInfo.Phone,
		Email:            req.UserInfo.Email,
		Login:            req.UserInfo.Login,
		Password:         hashedPassword,
		Active:           1, //@TODO:: user must be activated by phone or email
		ExpiresAt:        expiresAt,
	}
	user, err := s.strg.User().Create(ctx, createUserReq)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.strg.Company().TransferOwnership(ctx, companyPKey.Id, user.Id)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// sync auth companies with company service companies
	_, err = s.services.CompanyServiceClient().CreateCompany(
		ctx,
		&company_service.CreateCompanyRequest{
			Title:       req.Name,
			Logo:        "",
			Description: "",
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// @TODO:: remove when auth is independent from object builder
	structData, err := helper.ConvertRequestToSturct(createUserReq)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	//@DONE:: create company
	//@DONE:: create project
	//@DONE:: create client_platform
	//@DONE:: create client_type
	//@DONE:: create client
	//@DONE:: create role
	//@DONE:: permission
	//@TODO:: scope
	//@TODO:: permission_scope
	//@DONE:: role_permission
	//@DONE:: create user

	return companyPKey, nil
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
