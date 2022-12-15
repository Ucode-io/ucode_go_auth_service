package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/google/uuid"
	"github.com/saidamir98/udevs_pkg/logger"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"

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

	companyId, err := uuid.NewRandom()
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	companyPKey, err := s.services.CompanyServiceClient().CreateCompany(ctx, &company_service.CreateCompanyRequest{
		Title:       req.Name,
		Logo:        "",
		Description: "",
		OwnerId:     companyId.String(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	_, err = s.services.ProjectServiceClient().CreateProject(ctx, &company_service.CreateProjectRequest{
		CompanyId:    companyPKey.GetCompanyId(),
		K8SNamespace: "",
		Title:        req.GetName(),
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	// PROJECT
	createProjectReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"company_id": companyPKey.GetCompanyId(),
		"name":       req.GetName(),
		"domain":     config.UcodeTestAdminDomain,
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	createProjectResp, err := s.services.ObjectBuilderService().Create(
		ctx,
		&object_builder_service.CommonMessage{
			TableSlug: "project",
			Data:      createProjectReq,
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	if bytes, err := json.Marshal(createProjectResp); err == nil {
		fmt.Println("createProjectResp", string(bytes))
	}

	projectData, ok := createProjectResp.Data.AsMap()["data"].(map[string]interface{})
	if !ok || projectData == nil {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "project is nil"))
		return nil, err
	}

	projectID, ok := projectData["guid"].(string)
	if !ok {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "project_id is nil"))
		return nil, err
	}

	fmt.Println("projectID", projectID)

	// CLIENT_TYPE
	createClientTypeReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"name":          strings.ToUpper(req.Name) + " ADMIN",
		"confirm_by":    "UNDECIDED",
		"self_register": true,
		"self_recover":  true,
		"project_id":    projectID,
		// "client_platform_ids": []string{},
	})

	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	createClientTypeResp, err := s.services.ObjectBuilderService().Create(
		ctx,
		&object_builder_service.CommonMessage{
			TableSlug: "client_type",
			Data:      createClientTypeReq,
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	clientTypeData, ok := createClientTypeResp.Data.AsMap()["data"].(map[string]interface{})
	if !ok || clientTypeData == nil {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "clientType is nil"))
		return nil, err
	}

	clientTypeID, ok := clientTypeData["guid"].(string)
	if !ok {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "clientType_id is nil"))
		return nil, err
	}

	fmt.Println("clientTypeID", clientTypeID)

	// client_platform
	createClientPlatformReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"name":            "ADMIN PLATFORM",
		"subdomain":       config.UcodeTestAdminDomain,
		"project_id":      projectID,
		"client_type_ids": []string{clientTypeID},
	})

	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	createClientPlatformResp, err := s.services.ObjectBuilderService().Create(
		ctx,
		&object_builder_service.CommonMessage{
			TableSlug: "client_platform",
			Data:      createClientPlatformReq,
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	clientPlatformData, ok := createClientPlatformResp.Data.AsMap()["data"].(map[string]interface{})
	if !ok || clientPlatformData == nil {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "clientPlatform is nil"))
		return nil, err
	}

	clientPlatformID, ok := clientPlatformData["guid"].(string)
	if !ok {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "clientPlatform_id is nil"))
		return nil, err
	}

	fmt.Println("clientPlatformID", clientPlatformID)

	// TEST_LOGIN
	createTestLoginReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"login_strategy": "Login with password",
		"table_slug":     "user",
		"login_view":     "login",
		"login_label":    "Логин",
		"password_view":  "password",
		"object_id":      "2546e042-af2f-4cef-be7c-834e6bde951c",
		"password_label": "",
		"client_type_id": clientTypeID,
	})

	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	createTestLoginResp, err := s.services.ObjectBuilderService().Create(
		ctx,
		&object_builder_service.CommonMessage{
			TableSlug: "test_login",
			Data:      createTestLoginReq,
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	testLoginData, ok := createTestLoginResp.Data.AsMap()["data"].(map[string]interface{})
	if !ok || testLoginData == nil {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "testLogin is nil"))
		return nil, err
	}

	// ROLE
	createRoleReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"name":               "ADMIN",
		"project_id":         projectID,
		"client_platform_id": clientPlatformID,
		"client_type_id":     clientTypeID,
	})

	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	createRoleResp, err := s.services.ObjectBuilderService().Create(
		ctx,
		&object_builder_service.CommonMessage{
			TableSlug: "role",
			Data:      createRoleReq,
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	roleData, ok := createRoleResp.Data.AsMap()["data"].(map[string]interface{})
	if !ok || roleData == nil {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "role is nil"))
		return nil, err
	}

	roleID, ok := roleData["guid"].(string)
	if !ok {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "role_id is nil"))
		return nil, err
	}

	fmt.Println("roleID", roleID)

	// record_permission
	recordPermissionTableSlugs := []string{"app", "record_permission"}

	for _, recordPermission := range recordPermissionTableSlugs {
		createRecordPermissionReq, err := helper.ConvertMapToStruct(map[string]interface{}{
			"table_slug":        recordPermission,
			"update":            "Yes",
			"write":             "Yes",
			"read":              "Yes",
			"delete":            "Yes",
			"role_id":           roleID,
			"is_have_condition": false,
		})

		if err != nil {
			s.log.Error("---RegisterCompany--->", logger.Error(err))
			return nil, err
		}

		if err != nil {
			s.log.Error("---RegisterCompany--->", logger.Error(err))
			return nil, err
		}

		_, err = s.services.ObjectBuilderService().Create(
			ctx,
			&object_builder_service.CommonMessage{
				TableSlug: "record_permission",
				Data:      createRecordPermissionReq,
			},
		)
		if err != nil {
			s.log.Error("---RegisterCompany--->", logger.Error(err))
			return nil, err
		}
	}

	// USER
	createUserReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"phone":              req.UserInfo.Phone,
		"active":             1,
		"password":           req.UserInfo.Password,
		"login":              req.UserInfo.Login,
		"name":               "",
		"photo_url":          "",
		"salary":             0,
		"role_id":            roleID,
		"client_type_id":     clientTypeID,
		"client_platform_id": clientPlatformID,
		"project_id":         projectID,
	})

	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	createUserResp, err := s.services.ObjectBuilderService().Create(
		ctx,
		&object_builder_service.CommonMessage{
			TableSlug: "user",
			Data:      createUserReq,
		},
	)
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	userData, ok := createUserResp.Data.AsMap()["data"].(map[string]interface{})
	if !ok || userData == nil {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "user is nil"))
		return nil, err
	}

	userID, ok := userData["guid"].(string)
	if !ok {
		s.log.Error("---RegisterCompany--->", logger.Any("msg", "user_id is nil"))
		return nil, err
	}

	fmt.Println("userID", userID)

	_, err = s.services.CompanyServiceClient().UpdateCompany(ctx, &company_service.Company{
		CompanyId:   companyPKey.CompanyId,
		Name:        req.Name,
		Logo:        "",
		Description: "",
		OwnerId:     userID,
	})
	if err != nil {
		s.log.Error("---RegisterCompany--->", logger.Error(err))
		return nil, err
	}

	// //@DONE:: create company
	// //@DONE:: create project
	// //@DONE:: create client_platform
	// //@DONE:: create client_type
	// //@DONE:: create client
	// //@DONE:: create role
	// //@DONE:: permission
	// //@TODO:: scope
	// //@TODO:: permission_scope
	// //@DONE:: role_permission
	// //@DONE:: create user

	return &pb.CompanyPrimaryKey{Id: companyPKey.GetCompanyId()}, nil
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
