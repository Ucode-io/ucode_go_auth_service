package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/grpc/client"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/storage"

	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"ucode/ucode_go_auth_service/genproto/auth_service"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
)

type registerService struct {
	cfg      config.Config
	log      logger.LoggerI
	strg     storage.StorageI
	services client.ServiceManagerI
	pb.UnimplementedRegisterServiceServer
}

func NewRegisterService(cfg config.Config, log logger.LoggerI, strg storage.StorageI, svcs client.ServiceManagerI) *registerService {
	return &registerService{
		cfg:      cfg,
		log:      log,
		strg:     strg,
		services: svcs,
	}
}

func (rs *registerService) RegisterUser(ctx context.Context, data *pb.RegisterUserRequest) (*pb.V2LoginResponse, error) {
	rs.log.Info("--RegisterUser invoked--", logger.Any("data", data))
	body := data.Data.AsMap()
	var (
		foundUser                    *pb.User
		err, errorInAdditionalObject error
		userId                       string
		userData                     *pbObject.LoginDataRes
	)
	fmt.Println("::::::::::TEST:::::::::::1")
	switch strings.ToUpper(body["type"].(string)) {
	case "EMAIL":
		{
			foundUser, err = rs.strg.User().GetByUsername(ctx, body["email"].(string))
			if err != nil {
				rs.log.Error("!RegisterUser --->", logger.Error(err))
				return nil, err
			}
		}
	case "PHONE":
		{
			foundUser, err = rs.strg.User().GetByUsername(ctx, body["phone"].(string))
			if err != nil {
				rs.log.Error("!RegisterUser --->", logger.Error(err))
				return nil, err
			}
		}
	}
	fmt.Println("::::::::::TEST:::::::::::2")
	fmt.Println("user found::", foundUser)
	if foundUser.Id == "" {
		// create user in auth service
		var name, login, email, password, phone string
		if _, ok := body["name"]; ok {
			name = body["name"].(string)
		}
		if _, ok := body["login"]; ok {
			login = body["login"].(string)
		}
		if _, ok := body["password"]; ok {
			password = body["password"].(string)
		}
		if _, ok := body["phone"]; ok {
			phone = body["phone"].(string)
		}
		if _, ok := body["email"]; ok {
			email = body["email"].(string)
		}
		fmt.Println("::::::::::TEST:::::::::::3")
		pKey, err := rs.services.UserService().V2CreateUser(ctx, &auth_service.CreateUserRequest{
			Login:     login,
			Password:  password,
			Email:     email,
			Phone:     phone,
			Name:      name,
			CompanyId: body["company_id"].(string),
			ProjectId: body["project_id"].(string),
		})

		// pKey, err := rs.strg.User().Create(ctx, &auth_service.CreateUserRequest{
		// 	Login:     login,
		// 	Password:  password,
		// 	Email:     email,
		// 	Phone:     phone,
		// 	Name:      name,
		// 	CompanyId: body["company_id"].(string),
		// })
		fmt.Println("::::::::::TEST:::::::::::4")
		if err != nil {
			rs.log.Error("!!!RegisterUser.User().Create--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = pKey.GetId()
	} else {
		userId = foundUser.GetId()
	}
	fmt.Println("::::::::::TEST:::::::::::5")
	fmt.Println("user id in convert ::", userId)
	body["guid"] = userId
	structData, err := helper.ConvertMapToStruct(body)
	fmt.Println("::::::::::TEST:::::::::::6")
	if err != nil {
		rs.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// create user in object builder service
	fmt.Println("::::::::::TEST:::::::::::7")
	switch body["resource_type"].(float64) {
	case 1:
		_, err = rs.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: body["resource_environment_id"].(string),
		})
		fmt.Println("::::::::::TEST:::::::::::8")
		if err != nil {
			rs.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case 3:
		_, err = rs.services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: body["resource_environment_id"].(string),
		})
		if err != nil {
			rs.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}
	fmt.Println("::::::::::TEST:::::::::::9")
	if body["addational_table"] != nil {
		validRegisterForAddationalTable := map[string]bool{
			"phone": true,
			"email": true,
		}
		fmt.Println("::::::::::TEST:::::::::::10")
		if _, ok := validRegisterForAddationalTable[body["type"].(string)]; ok {
			body["addational_table"].(map[string]interface{})["guid"] = userId
			body["addational_table"].(map[string]interface{})["project_id"] = body["project_id"]

			mapedInterface := body["addational_table"].(map[string]interface{})
			structData, errorInAdditionalObject = helper.ConvertRequestToSturct(mapedInterface)
			if errorInAdditionalObject != nil {
				rs.log.Error("Additional table struct table --->", logger.Error(err))
			}
			fmt.Println("::::::::::TEST:::::::::::11")
			_, errorInAdditionalObject = rs.services.ObjectBuilderService().Create(
				context.Background(),
				&pbObject.CommonMessage{
					TableSlug: mapedInterface["table_slug"].(string),
					Data:      structData,
					ProjectId: body["resource_environment_id"].(string),
				})
			if errorInAdditionalObject != nil {
				rs.log.Error("!!!RegisterUser--->Additional Object create error >>", logger.Error(err))
			}
		}
		fmt.Println("::::::::::TEST:::::::::::12")
		defer func(userId string) {
			// delete user from object builder user table if has any error while create additional object
			if errorInAdditionalObject != nil {
				structData, errorInAdditionalObject = helper.ConvertRequestToSturct(map[string]interface{}{
					"id": userId,
				})
				_, errorInAdditionalObject = rs.services.ObjectBuilderService().Delete(
					context.Background(),
					&pbObject.CommonMessage{
						TableSlug: "user",
						Data:      structData,
						ProjectId: body["resource_environment_id"].(string),
					})
				if errorInAdditionalObject != nil {
					rs.log.Error("!!!RegisterUser--->delete user if have error while create additional user >>", logger.Error(err))
				}
			}
		}(userId)
	}
	reqLoginData := &pbObject.LoginDataReq{
		UserId:                userId,
		ProjectId:             body["project_id"].(string),
		ResourceEnvironmentId: body["resource_environment_id"].(string),
	}
	fmt.Println("::::::::::TEST:::::::::::13")
	switch body["resource_type"].(float64) {
	case 1:
		userData, err = rs.services.LoginService().LoginDataByUserId(
			ctx,
			reqLoginData,
		)
		fmt.Println("::::::::::TEST:::::::::::14")
		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			rs.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}
	case 3:
		userData, err = rs.services.PostgresLoginService().LoginDataByUserId(
			ctx,
			reqLoginData,
		)

		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			rs.log.Error("!!!PostgresBuilder.Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}

	}
	if bytes, err := json.MarshalIndent(data, "", "  "); err == nil {
		fmt.Println("ConvertPbToAnotherPb", string(bytes))
	}
	fmt.Println("::::::::::TEST:::::::::::15")
	if !userData.UserFound {
		customError := errors.New("User not found")
		rs.log.Error("!!!Login--->", logger.Error(customError))
		return nil, status.Error(codes.NotFound, customError.Error())
	}
	fmt.Println("::::::::::TEST:::::::::::16")
	res := helper.ConvertPbToAnotherPb(&pbObject.V2LoginResponse{
		ClientPlatform: userData.GetClientPlatform(),
		ClientType:     userData.GetClientType(),
		UserFound:      userData.GetUserFound(),
		UserId:         userData.GetUserId(),
		Role:           userData.GetRole(),
		Permissions:    userData.GetPermissions(),
		LoginTableSlug: userData.GetLoginTableSlug(),
		AppPermissions: userData.GetAppPermissions(),
	})

	resp, err := rs.services.SessionService().SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
		LoginData:     res,
		ProjectId:     body["project_id"].(string),
		Tables:        []*pb.Object{},
		EnvironmentId: body["environment_id"].(string),
	})
	if resp == nil {
		err := errors.New("User Not Found")
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
