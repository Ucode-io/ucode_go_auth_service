package service

import (
	"context"
	"fmt"
	"strings"
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
	body := data.Data.AsMap()
	var (
		foundUser *pb.User
		err       error
		userId    string
		// userData  *pbObject.LoginDataRes

		login    = helper.GetStringFromMap(body, "login")
		email    = helper.GetStringFromMap(body, "email")
		password = helper.GetStringFromMap(body, "password")
		phone    = helper.GetStringFromMap(body, "phone")
	)

	switch strings.ToUpper(data.Type) {
	case "EMAIL":
		foundUser, err = rs.strg.User().GetByUsername(ctx, email)
		if err != nil {
			rs.log.Error("!RegisterUserError--->EmailGetByUsername", logger.Error(err))
			return nil, err
		}
	case "PHONE":
		foundUser, err = rs.strg.User().GetByUsername(ctx, phone)
		if err != nil {
			rs.log.Error("!RegisterUserError--->PhoneGetByUsernames", logger.Error(err))
			return nil, err
		}
	}

	userId = foundUser.GetId()

	if len(foundUser.GetId()) == 0 {
		if !helper.EmailValidation(email) && len(email) > 0 {
			err = fmt.Errorf("email is not valid")
			rs.log.Error("!!!CreateUser--->EmailValidation", logger.Error(err))
			return nil, err
		}

		hashedPassword, err := security.HashPassword(password)
		if err != nil {
			rs.log.Error("!!!CreateUser--->HashPassword", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		password = hashedPassword

		pKey, err := rs.strg.User().Create(ctx, &pb.CreateUserRequest{
			Login:     login,
			Password:  password,
			Email:     email,
			Phone:     phone,
			CompanyId: data.GetCompanyId(),
		})
		if err != nil {
			rs.log.Error("!!!CreateUser--->CreateRequest", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = pKey.GetId()
	}

	return &pb.V2LoginResponse{}, nil

	body["guid"] = userId
	body["from_auth_service"] = true
	structData, err := helper.ConvertMapToStruct(body)
	if err != nil {
		rs.log.Error("!!!CreateUser--->ConvertMapToStruct", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.V2LoginResponse{}, nil

	services, err := rs.serviceNode.GetByNodeType(
		data.ProjectId,
		data.NodeType,
	)
	if err != nil {
		rs.log.Error("!!!CreateUser--->GetByNodeType", logger.Error(err))
		return nil, err
	}

	var (
		resourceType = body["resource_type"].(float64)
		tableSlug    = "user"
	)

	return &pb.V2LoginResponse{}, nil

	switch resourceType {
	case 1:
		response, err := services.GetObjectBuilderServiceByType(data.NodeType).GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "client_type",
			ProjectId: data.ResourceEnvironmentId,
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(data.ClientTypeId),
				},
			},
		})
		if err != nil {
			rs.log.Error("!!!CreateUser--->Node GetSingle", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		clientType, ok := response.Data.AsMap()["response"]
		if ok && clientType != nil {
			if clientTypeTableSlug, ok := clientType.(map[string]interface{})["table_slug"]; ok {
				tableSlug = clientTypeTableSlug.(string)
			}
		}

		_, err = services.GetObjectBuilderServiceByType(data.NodeType).Create(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: data.ResourceEnvironmentId,
		})
		if err != nil {
			rs.log.Error("!!!CreateUser--->NodeType Create", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:
		response, err := services.GoItemService().GetSingle(ctx, &new_object_builder_service.CommonMessage{
			TableSlug: "client_type",
			ProjectId: data.ResourceEnvironmentId,
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(data.ClientTypeId),
				},
			},
		})
		if err != nil {
			rs.log.Error("!!!CreateUser--->GetSingle", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		if clientType, ok := response.Data.AsMap()["response"]; ok {
			if clientTypeTableSlug, ok := clientType.(map[string]interface{})["table_slug"]; ok {
				tableSlug = clientTypeTableSlug.(string)
			}
		}

		_, err = services.GoItemService().Create(ctx, &new_object_builder_service.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: data.ResourceEnvironmentId,
		})
		if err != nil {
			rs.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return &pb.V2LoginResponse{}, nil

	// _, err = rs.strg.User().AddUserToProject(ctx, &pb.AddUserToProjectReq{
	// 	UserId:       userId,
	// 	RoleId:       data.RoleId,
	// 	CompanyId:    data.CompanyId,
	// 	ProjectId:    data.ProjectId,
	// 	ClientTypeId: data.ClientTypeId,
	// 	EnvId:        data.EnvironmentId,
	// })
	// if err != nil {
	// 	rs.log.Error("!RegisterUserError--->AddUserToProject", logger.Error(err))
	// 	return nil, err
	// }
	// reqLoginData := &pbObject.LoginDataReq{
	// 	UserId:                userId,
	// 	ProjectId:             data.ProjectId,
	// 	ClientType:            data.ClientTypeId,
	// 	ResourceEnvironmentId: data.ResourceEnvironmentId,
	// }

	// switch resourceType {
	// case 1:
	// 	userData, err = services.GetLoginServiceByType(data.NodeType).LoginData(ctx, reqLoginData)
	// 	if err != nil {
	// 		errGetUserProjectData := errors.New("invalid user project data")
	// 		rs.log.Error("!!!Login--->LoginData", logger.Error(err))
	// 		return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
	// 	}
	// case 3:
	// 	pgLoginData := &new_object_builder_service.LoginDataReq{}

	// 	err := helper.MarshalToStruct(reqLoginData, &pgLoginData)
	// 	if err != nil {
	// 		rs.log.Error("!!!PostgresBuilder.Login--->", logger.Error(err))
	// 		return nil, status.Error(codes.Internal, err.Error())
	// 	}

	// 	pgUserData, err := services.GoLoginService().LoginData(
	// 		ctx,
	// 		pgLoginData,
	// 	)
	// 	if err != nil {
	// 		errGetUserProjectData := errors.New("invalid user project data")
	// 		rs.log.Error("!!!PostgresBuilder.Login--->", logger.Error(err))
	// 		return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
	// 	}

	// 	err = helper.MarshalToStruct(&pgUserData, &userData)
	// 	if err != nil {
	// 		errGetUserProjectData := errors.New("invalid user project data")
	// 		rs.log.Error("!!!Login--->", logger.Error(err))
	// 		return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
	// 	}
	// }

	// if !userData.UserFound {
	// 	customError := errors.New("user not found")
	// 	rs.log.Error("!!!Login--->", logger.Error(customError))
	// 	return nil, status.Error(codes.NotFound, customError.Error())
	// }

	// res := helper.ConvertPbToAnotherPb(&pbObject.V2LoginResponse{
	// 	ClientPlatform: userData.GetClientPlatform(),
	// 	ClientType:     userData.GetClientType(),
	// 	UserFound:      userData.GetUserFound(),
	// 	UserId:         userData.GetUserId(),
	// 	Role:           userData.GetRole(),
	// 	Permissions:    userData.GetPermissions(),
	// 	LoginTableSlug: userData.GetLoginTableSlug(),
	// })

	// res, err = rs.services.SessionService().SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
	// 	LoginData:     res,
	// 	ProjectId:     data.ProjectId,
	// 	Tables:        []*pb.Object{},
	// 	EnvironmentId: data.EnvironmentId,
	// })
	// if res == nil {
	// 	err := errors.New("user not found")
	// 	rs.log.Error("!!!Login--->SessionAndTokenGenerator", logger.Error(err))
	// 	return nil, status.Error(codes.NotFound, err.Error())
	// }
	// if err != nil {
	// 	rs.log.Error("!!!Login--->SessionAndTokenGenerator", logger.Error(err))
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// res.GlobalPermission = nil
	// res.UserData = nil
	// res.UserFound = true
	// res.UserId = userId

	// return res, nil
}
