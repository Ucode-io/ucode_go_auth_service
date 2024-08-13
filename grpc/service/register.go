package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/security"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/new_object_builder_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
)

type registerService struct {
	cfg         config.BaseConfig
	log         logger.LoggerI
	strg        storage.StorageI
	services    client.ServiceManagerI
	serviceNode ServiceNodesI
	pb.UnimplementedRegisterServiceServer
}

func NewRegisterService(cfg config.BaseConfig, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI, projectServiceNodes ServiceNodesI) *registerService {
	return &registerService{
		cfg:         cfg,
		log:         log,
		strg:        strg,
		services:    svcs,
		serviceNode: projectServiceNodes,
	}
}

func (rs *registerService) RegisterUser(ctx context.Context, data *pb.RegisterUserRequest) (*pb.V2LoginResponse, error) {
	rs.log.Info("--RegisterUser invoked--", logger.Any("data", data))
	body := make(map[string]interface{})
	var (
		foundUser *pb.User
		err       error
		userId    string
		userData  *pbObject.LoginDataRes
	)

	switch strings.ToUpper(body["type"].(string)) {
	case "EMAIL":
		{
			foundUser, err = rs.strg.User().GetByUsername(ctx, body["email"].(string))
			if err != nil {
				rs.log.Error("!RegisterUserError--->EmailGetByUsername", logger.Error(err))
				return nil, err
			}
		}
	case "PHONE":
		{
			foundUser, err = rs.strg.User().GetByUsername(ctx, body["phone"].(string))
			if err != nil {
				rs.log.Error("!RegisterUserError--->PhoneGetByUsernames", logger.Error(err))
				return nil, err
			}
		}
	}

	userId = foundUser.GetId()

	if foundUser.Id == "" {
		// create user in auth service
		var login, email, password, phone string

		if _, ok := body["login"]; ok {
			login = body["login"].(string)
		}
		if _, ok := body["password"].(string); ok {
			password = body["password"].(string)
		}
		if _, ok := body["phone"].(string); ok {
			phone = body["phone"].(string)
		}
		if _, ok := body["email"].(string); ok {
			email = body["email"].(string)
		}
		emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
		checkEmail := emailRegex.MatchString(email)
		if !checkEmail && email != "" {
			err = fmt.Errorf("email is not valid")
			rs.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, err
		}
		hashedPassword, err := security.HashPassword(password)
		if err != nil {
			rs.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		password = hashedPassword

		pKey, err := rs.strg.User().Create(ctx, &pb.CreateUserRequest{
			Login:     login,
			Password:  password,
			Email:     email,
			Phone:     phone,
			CompanyId: body["company_id"].(string),
		})

		if err != nil {
			rs.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = pKey.GetId()
	}

	body["guid"] = userId
	body["from_auth_service"] = true
	structData, err := helper.ConvertMapToStruct(body)

	if err != nil {
		rs.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	services, err := rs.serviceNode.GetByNodeType(
		body["project_id"].(string),
		data.NodeType,
	)
	if err != nil {
		return nil, err
	}

	var tableSlug = "user"
	switch body["resource_type"].(float64) {
	case 1:
		response, err := services.GetObjectBuilderServiceByType(data.NodeType).GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(body["client_type_id"].(string)),
				},
			},
			ProjectId: body["resource_environment_id"].(string),
		})
		if err != nil {
			rs.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		clientType, ok := response.Data.AsMap()["response"]
		if ok && clientType != nil {
			if clientTypeTableSlug, ok := clientType.(map[string]interface{})["table_slug"]; ok {
				tableSlug = clientTypeTableSlug.(string)
			}
		}
	case 3:
		response, err := services.GoItemService().GetSingle(ctx, &new_object_builder_service.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(body["client_type_id"].(string)),
				},
			},
			ProjectId: body["resource_environment_id"].(string),
		})
		if err != nil {
			rs.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		clientType, ok := response.Data.AsMap()["response"]
		if ok {
			if clientTypeTableSlug, ok := clientType.(map[string]interface{})["table_slug"]; ok {
				tableSlug = clientTypeTableSlug.(string)
			}
		}
	}

	// create user in object builder service
	switch body["resource_type"].(float64) {
	case 1:
		_, err = services.GetObjectBuilderServiceByType(data.NodeType).Create(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: body["resource_environment_id"].(string),
		})
		if err != nil {
			rs.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:
		_, err = services.GoItemService().Create(ctx, &new_object_builder_service.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: body["resource_environment_id"].(string),
		})
		if err != nil {
			rs.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	_, err = rs.strg.User().AddUserToProject(ctx, &pb.AddUserToProjectReq{
		UserId:       userId,
		ProjectId:    body["project_id"].(string),
		CompanyId:    body["company_id"].(string),
		ClientTypeId: body["client_type_id"].(string),
		RoleId:       body["role_id"].(string),
		EnvId:        body["environment_id"].(string),
	})
	if err != nil {
		rs.log.Error("!RegisterUserError--->AddUserToProject", logger.Error(err))
		return nil, err
	}
	reqLoginData := &pbObject.LoginDataReq{
		UserId:                userId,
		ClientType:            body["client_type_id"].(string),
		ProjectId:             body["project_id"].(string),
		ResourceEnvironmentId: body["resource_environment_id"].(string),
	}

	t2 := time.Now()
	switch body["resource_type"].(float64) {
	case 1:
		userData, err = services.GetLoginServiceByType(data.NodeType).LoginData(
			ctx,
			reqLoginData,
		)

		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			rs.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}
	case 3:
		pgLoginData := &new_object_builder_service.LoginDataReq{}
		err := helper.MarshalToStruct(reqLoginData, &pgLoginData)
		if err != nil {
			rs.log.Error("!!!PostgresBuilder.Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		pgUserData, err := services.GoLoginService().LoginData(
			ctx,
			pgLoginData,
		)
		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			rs.log.Error("!!!PostgresBuilder.Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}

		err = helper.MarshalToStruct(&pgUserData, &userData)
		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			rs.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}

	}

	fmt.Println("SINCE2", time.Since(t2))

	if !userData.UserFound {
		customError := errors.New("user not found")
		rs.log.Error("!!!Login--->", logger.Error(customError))
		return nil, status.Error(codes.NotFound, customError.Error())
	}

	t := time.Now()
	res := helper.ConvertPbToAnotherPb(&pbObject.V2LoginResponse{
		ClientPlatform: userData.GetClientPlatform(),
		ClientType:     userData.GetClientType(),
		UserFound:      userData.GetUserFound(),
		UserId:         userData.GetUserId(),
		Role:           userData.GetRole(),
		Permissions:    userData.GetPermissions(),
		LoginTableSlug: userData.GetLoginTableSlug(),
	})
	fmt.Println("SINCE", time.Since(t))

	t1 := time.Now()
	resp, err := rs.services.SessionService().SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
		LoginData:     res,
		ProjectId:     body["project_id"].(string),
		Tables:        []*pb.Object{},
		EnvironmentId: body["environment_id"].(string),
	})
	fmt.Println("SINCE1", time.Since(t1))
	if resp == nil {
		err := errors.New("user Not Found")
		rs.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		rs.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.V2LoginResponse{
		UserFound:       true,
		UserId:          userId,
		Token:           resp.GetToken(),
		Sessions:        resp.GetSessions(),
		ClientPlatform:  resp.GetClientPlatform(),
		ClientType:      resp.GetClientType(),
		Role:            resp.GetRole(),
		Permissions:     resp.GetPermissions(),
		AppPermissions:  resp.GetAppPermissions(),
		Tables:          resp.GetTables(),
		LoginTableSlug:  resp.GetLoginTableSlug(),
		AddationalTable: resp.GetAddationalTable(),
		ResourceId:      resp.GetResourceId(),
		EnvironmentId:   resp.GetEnvironmentId(),
		User:            resp.GetUser(),
	}, nil
}
