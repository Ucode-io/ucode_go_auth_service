package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"ucode/ucode_go_auth_service/config"

	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	nb "ucode/ucode_go_auth_service/genproto/new_object_builder_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/saidamir98/udevs_pkg/security"
	"github.com/spf13/cast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *userService) RegisterWithGoogle(ctx context.Context, req *pb.RegisterWithGoogleRequest) (resp *pb.User, err error) {

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := emailRegex.MatchString(req.Email)
	if !email {
		err = fmt.Errorf("email is not valid")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, err
	}

	foundUser, err := s.strg.User().GetByUsername(ctx, req.Email)
	if err != nil {
		s.log.Error("!!!Get User by name--->", logger.Error(err))
		return nil, err
	}

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	if foundUser.Id == "" {
		pKey, err := s.strg.User().Create(ctx, &pb.CreateUserRequest{
			Login:                 "",
			Password:              "",
			Email:                 req.Email,
			Phone:                 "",
			Name:                  req.GetName(),
			CompanyId:             req.GetCompanyId(),
			ProjectId:             req.GetProjectId(),
			ResourceEnvironmentId: req.GetResourceEnvironmentId(),
			RoleId:                "",
			ClientTypeId:          req.GetClientTypeId(),
			ClientPlatformId:      "",
			Active:                -1,
		})
		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
			"guid":           pKey.GetId(),
			"project_id":     req.GetProjectId(),
			"role_id":        "",
			"client_type_id": req.GetClientTypeId(),
			"active":         "",
			"expires_at":     "",
			"email":          req.GetEmail(),
			"phone":          "",
			"name":           req.GetName(),
			"login":          "",
		})
		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		switch req.ResourceType {
		case 1:
			_, err = services.GetObjectBuilderServiceByType(req.NodeType).Create(ctx, &pbObject.CommonMessage{
				TableSlug: "user",
				Data:      structData,
				ProjectId: req.GetResourceEnvironmentId(),
			})
			if err != nil {
				s.log.Error("!!!ObjectBuilderService.CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		case 3:
			_, err = services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
				TableSlug: "user",
				Data:      structData,
				ProjectId: req.GetResourceEnvironmentId(),
			})
			if err != nil {
				s.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		_, err = s.strg.User().AddUserToProject(ctx, &pb.AddUserToProjectReq{
			UserId:    pKey.Id,
			ProjectId: req.GetProjectId(),
			CompanyId: req.GetCompanyId(),
		})
		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		resp, err = s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
			Id: pKey.Id,
		})
		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		return resp, err
	} else {
		var objUser *pbObject.V2LoginResponse

		if req.Email != "" {
			switch req.ResourceType {
			case 1:
				objUser, err = services.GetLoginServiceByType(req.NodeType).LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

					Email:      req.Email,
					ClientType: "WEB_USER",
					ProjectId:  req.GetResourceEnvironmentId(),
					TableSlug:  "user",
				})
				if err != nil {
					s.log.Error("!!!Found user from obj--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			case 3:
				objUser, err = services.PostgresLoginService().LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

					Email:      req.Email,
					ClientType: "WEB_USER",
					ProjectId:  req.GetResourceEnvironmentId(),
					TableSlug:  "user",
				})
				if err != nil {
					s.log.Error("!!!Found user from obj--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			}

		}

		if objUser.UserFound {
			s.log.Error("!!!Found user from obj--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, "User already exists")
		} else {
			structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
				"guid":           foundUser.Id,
				"project_id":     req.GetProjectId(),
				"role_id":        "",
				"client_type_id": req.GetClientTypeId(),
				"active":         "",
				"expires_at":     "",
				"email":          req.GetEmail(),
				"phone":          "",
				"name":           req.GetName(),
				"login":          "",
			})
			if err != nil {
				s.log.Error("!!!CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}

			switch req.ResourceType {
			case 1:
				_, err = services.GetObjectBuilderServiceByType(req.NodeType).Create(ctx, &pbObject.CommonMessage{
					TableSlug: "user",
					Data:      structData,
					ProjectId: req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!CreateUser--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			case 3:
				_, err = services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
					TableSlug: "user",
					Data:      structData,
					ProjectId: req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}

			}

			resp, err = s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
				Id: foundUser.Id,
			})
			if err != nil {
				s.log.Error("!!!CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			return resp, err
		}
	}
}

func (s *userService) RegisterUserViaEmail(ctx context.Context, req *pb.CreateUserRequest) (resp *pb.User, err error) {

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	req.Password = hashedPassword

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := emailRegex.MatchString(req.Email)
	if !email {
		err = fmt.Errorf("email is not valid")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, err
	}

	foundUser, _ := s.strg.User().GetByUsername(ctx, req.Email)
	if foundUser.Id == "" {
		foundUser, _ = s.strg.User().GetByUsername(ctx, req.Phone)
	}

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	if foundUser.Id == "" {
		pKey, err := s.strg.User().Create(ctx, &pb.CreateUserRequest{
			Login:    req.GetLogin(),
			Password: req.GetPassword(),
			//	Email:                 req.GetEmail(),
			Phone:                 req.GetPhone(),
			Name:                  req.GetName(),
			CompanyId:             req.GetCompanyId(),
			ProjectId:             req.GetProjectId(),
			ResourceEnvironmentId: req.GetResourceEnvironmentId(),
			RoleId:                req.GetRoleId(),
			ClientTypeId:          req.GetClientTypeId(),
			ClientPlatformId:      req.GetClientPlatformId(),
			Active:                req.GetActive(),
		})
		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
			"guid":           pKey.GetId(),
			"project_id":     req.GetProjectId(),
			"role_id":        req.GetRoleId(),
			"client_type_id": req.GetClientTypeId(),
			"active":         req.GetActive(),
			"expires_at":     req.GetExpiresAt(),
			"email":          req.GetEmail(),
			"phone":          req.GetPhone(),
			"name":           req.GetName(),
			"login":          req.GetLogin(),
		})
		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		switch req.ResourceType {
		case 1:
			_, err = services.GetObjectBuilderServiceByType(req.NodeType).Create(ctx, &pbObject.CommonMessage{
				TableSlug: "user",
				Data:      structData,
				ProjectId: req.GetResourceEnvironmentId(),
			})
			if err != nil {
				s.log.Error("!!!CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		case 3:
			_, err = services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
				TableSlug: "user",
				Data:      structData,
				ProjectId: req.GetResourceEnvironmentId(),
			})
			if err != nil {
				s.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}

		_, err = s.strg.User().AddUserToProject(ctx, &pb.AddUserToProjectReq{
			UserId:    pKey.Id,
			ProjectId: req.GetProjectId(),
			CompanyId: req.GetCompanyId(),
		})
		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		resp, err = s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
			Id: pKey.Id,
		})
		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		return resp, err
	} else {
		var objUser *pbObject.V2LoginResponse

		if req.Email != "" {
			switch req.ResourceType {
			case 1:
				objUser, err = services.GetLoginServiceByType(req.NodeType).LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

					Email:      req.Email,
					ClientType: "WEB_USER",
					ProjectId:  req.GetResourceEnvironmentId(),
					TableSlug:  "user",
				})
				if err != nil {
					s.log.Error("!!!Found user from obj--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			case 3:
				objUser, err = services.PostgresLoginService().LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

					Email:      req.Email,
					ClientType: "WEB_USER",
					ProjectId:  req.GetResourceEnvironmentId(),
					TableSlug:  "user",
				})
				if err != nil {
					s.log.Error("!!!Found user from obj--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			}

		}
		if objUser != nil && req.Phone != "" && !objUser.UserFound {
			switch req.ResourceType {
			case 1:
				objUser, err = services.GetLoginServiceByType(req.NodeType).LoginWithOtp(context.Background(), &pbObject.PhoneOtpRequst{

					PhoneNumber: req.Phone,
					ClientType:  "WEB_USER",
					ProjectId:   req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!Found user from obj--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			case 3:
				objUser, err = services.PostgresLoginService().LoginWithOtp(context.Background(), &pbObject.PhoneOtpRequst{

					PhoneNumber: req.Phone,
					ClientType:  "WEB_USER",
					ProjectId:   req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!Found user from obj--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			}

		}

		if objUser != nil && objUser.UserFound {
			s.log.Error("!!!Found user from obj--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, "User already exists")
		} else {
			structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
				"guid":           foundUser.Id,
				"project_id":     req.GetProjectId(),
				"role_id":        req.GetRoleId(),
				"client_type_id": req.GetClientTypeId(),
				"active":         req.GetActive(),
				"expires_at":     req.GetExpiresAt(),
				"email":          req.GetEmail(),
				"phone":          req.GetPhone(),
				"name":           req.GetName(),
				"login":          req.GetLogin(),
			})
			if err != nil {
				s.log.Error("!!!CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}

			switch req.ResourceType {
			case 1:
				_, err = services.GetObjectBuilderServiceByType(req.NodeType).Create(ctx, &pbObject.CommonMessage{
					TableSlug: "user",
					Data:      structData,
					ProjectId: req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!CreateUser--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			case 3:
				_, err = services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
					TableSlug: "user",
					Data:      structData,
					ProjectId: req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}

			}

			resp, err = s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
				Id: foundUser.Id,
			})
			if err != nil {
				s.log.Error("!!!CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			return resp, err
		}
	}

}

func (s *userService) V2CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	s.log.Info("\n\n\n\n---V2CreateUser--->", logger.Any("req", req))

	unHashedPassword := req.Password

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		s.log.Error("!!!V2CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	req.Password = hashedPassword

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := emailRegex.MatchString(req.Email)
	if !email && req.Email != "" {
		err = fmt.Errorf("email is not valid")
		s.log.Error("!!!V2CreateUser--->", logger.Error(err))
		return nil, err
	}

	pKey, err := s.strg.User().Create(ctx, req)

	if err != nil {
		s.log.Error("!!!V2CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// objectBuilder -> auth service
	structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
		"guid":               pKey.GetId(),
		"project_id":         req.GetProjectId(),
		"role_id":            req.GetRoleId(),
		"client_type_id":     req.GetClientTypeId(),
		"client_platform_id": req.GetClientPlatformId(),
		"active":             req.GetActive(),
		"expires_at":         req.GetExpiresAt(),
		"name":               req.GetName(),
		"email":              req.GetEmail(),
		"photo":              req.GetPhotoUrl(),
		"password":           req.GetPassword(),
		"login":              req.GetLogin(),
		"birth_day":          req.GetYearOfBirth(),
		"phone":              req.GetPhone(),
		"from_auth_service":  true,
	})
	if err != nil {
		s.log.Error("!!!V2CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var tableSlug = "user"

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	switch req.ResourceType {
	case 1:
		clientType, err := services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(context.Background(), &pbObject.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		_, err = services.GetObjectBuilderServiceByType(req.NodeType).Create(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})

		if err != nil {
			s.log.Error("!!!V2CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		clientType, err := services.PostgresObjectBuilderService().GetSingle(context.Background(), &pbObject.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		_, err = services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})

		if err != nil {
			s.log.Error("!!!PostgresObjectBuilderService.V2CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	_, err = s.strg.User().AddUserToProject(ctx, &pb.AddUserToProjectReq{
		UserId:       pKey.Id,
		ProjectId:    req.GetProjectId(),
		CompanyId:    req.GetCompanyId(),
		ClientTypeId: req.GetClientTypeId(),
		RoleId:       req.GetRoleId(),
		EnvId:        req.GetEnvironmentId(),
	})

	if err != nil {
		s.log.Error("!!!V2CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	if req.GetInvite() {
		emailSettings, err := s.strg.Email().GetListEmailSettings(ctx, &pb.GetListEmailSettingsRequest{
			ProjectId: req.GetProjectId(),
		})
		if err != nil {
			return nil, err
		}
		var devEmail string
		var devEmailPassword string
		if len(emailSettings.GetItems()) > 0 {
			devEmail = emailSettings.GetItems()[0].GetEmail()
			devEmailPassword = emailSettings.GetItems()[0].GetPassword()
		}
		err = helper.SendInviteMessageToEmail(helper.SendMessageToEmailRequest{
			Subject:       "Invite message",
			To:            req.GetEmail(),
			UserId:        pKey.Id,
			Email:         devEmail,
			Password:      devEmailPassword,
			Username:      req.GetLogin(),
			TempPassword:  unHashedPassword,
			EnvironmentId: req.GetEnvironmentId(),
			ClientTypeId:  req.GetClientTypeId(),
			ProjectId:     req.GetProjectId(),
		})
		if err != nil {
			s.log.Error("Error while sending message to invite")
			s.log.Error(err.Error())
		}
	}

	return s.strg.User().GetByPK(ctx, pKey)
}

func (s *userService) V2GetUserByID(ctx context.Context, req *pb.UserPrimaryKey) (*pb.User, error) {
	s.log.Info("---V2GetUserByID--->", logger.Any("req", req))

	var (
		result   *pbObject.CommonMessage
		resultGo *nb.CommonMessage
		userData map[string]interface{}
		ok       bool
	)
	user, err := s.strg.User().GetByPK(ctx, req)

	if err != nil {
		s.log.Error("!!!GetUserByID--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
		"id": req.Id,
	})
	if err != nil {
		s.log.Error("!!!GetUserByID--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	var tableSlug = "user"
	switch req.ResourceType {
	case 1:
		clientType, err := services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(context.Background(), &pbObject.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetUserByID--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		result, err = services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		clientType, err := services.GoItemService().GetSingle(context.Background(), &nb.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetUserByID--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		resultGo, err = services.GoItemService().GetSingle(ctx, &nb.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetUserByID.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if result != nil {
		userData, ok = result.Data.AsMap()["response"].(map[string]interface{})
	} else {
		userData, ok = resultGo.Data.AsMap()["response"].(map[string]interface{})
	}

	if !ok {
		err := errors.New("userData is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	roleId, ok := userData["role_id"].(string)
	if !ok {
		err := errors.New("role_id is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	user.RoleId = roleId

	clientTypeId, ok := userData["client_type_id"].(string)
	if !ok {
		err := errors.New("client_type_id is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	user.ClientTypeId = clientTypeId

	active, ok := userData["active"].(float64)
	if !ok {
		active = 0
	}

	user.Active = int32(active)

	projectId, ok := userData["project_id"].(string)
	if !ok {
		// err := errors.New("projectId is nil")
		// s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		// return nil, status.Error(codes.Internal, err.Error())
		projectId = ""
	}
	name, ok := userData["name"].(string)
	if ok {
		// err := errors.New("projectId is nil")
		// s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		// return nil, status.Error(codes.Internal, err.Error())
		user.Name = name
	}

	user.ProjectId = projectId

	return user, nil
}

func (s *userService) V2GetUserList(ctx context.Context, req *pb.GetUserListRequest) (*pb.GetUserListResponse, error) {
	s.log.Info("---V2GetUserList--->", logger.Any("req", req))

	resp := &pb.GetUserListResponse{}
	var (
		usersResp *pbObject.CommonMessage
	)

	userIds, err := s.strg.User().GetUserIds(ctx, req)
	if err != nil {
		s.log.Error("!!!V2GetUserList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	users, err := s.strg.User().GetListByPKs(ctx, &pb.UserPrimaryKeyList{
		Ids: *userIds,
	})
	if err != nil {
		s.log.Error("!!!V2GetUserList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	usersMap := make(map[string]*pb.User, users.Count)

	for _, user := range users.Users {
		usersMap[user.Id] = user
	}

	structReq := map[string]interface{}{
		"guid": map[string]interface{}{
			"$in": userIds,
		},
	}

	if util.IsValidUUID(req.ClientTypeId) {
		structReq["client_type_id"] = req.ClientTypeId
	}

	if util.IsValidUUID(req.ClientPlatformId) {
		structReq["client_platform_id"] = req.ClientPlatformId
	}

	structData, err := helper.ConvertRequestToSturct(structReq)
	if err != nil {
		s.log.Error("!!!V2GetUserList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	var tableSlug = "user"
	switch req.ResourceType {
	case 1:
		fmt.Println("aaaa:", userIds)
		clientType, err := services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(context.Background(), &pbObject.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2GetUserList--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		usersResp, err = services.GetObjectBuilderServiceByType(req.NodeType).GetList(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		clientType, err := services.GoItemService().GetSingle(context.Background(), &nb.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2GetUserList--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		goUsersResp, err := services.GoObjectBuilderService().GetList2(ctx, &nb.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetUserList.PostgresObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		if err := helper.MarshalToStruct(goUsersResp, &usersResp); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	userCount, ok := usersResp.Data.AsMap()["count"].(float64)
	if !ok {
		err := errors.New("usersData is nil")
		s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	usersData, ok := usersResp.Data.AsMap()["response"].([]interface{})
	if !ok {
		err := errors.New("usersData is nil")
		s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp.Users = make([]*pb.User, 0, int(userCount))
	resp.Count = int32(userCount)

	for _, userData := range usersData {
		userItem, ok := userData.(map[string]interface{})
		if !ok {
			err := errors.New("userItem is nil")
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		userId, ok := userItem["guid"].(string)
		if !ok {
			err := errors.New("userId is nil")
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		roleId, ok := userItem["role_id"].(string)
		if !ok {
			err := errors.New("roleId is nil")
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		clientTypeId, ok := userItem["client_type_id"].(string)
		if !ok {
			err := errors.New("clientTypeId is nil")
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		// clientPlatformId, ok := userItem["client_platform_id"].(string)
		// if !ok {
		// 	err := errors.New("clientPlatformId is nil")
		// 	s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
		// 	return nil, status.Error(codes.Internal, err.Error())
		// }

		projectId, ok := userItem["project_id"].(string)
		if !ok {
			// err := errors.New("projectId is nil")
			// s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			// return nil, status.Error(codes.Internal, err.Error())
			projectId = ""
		}

		active, ok := userItem["active"].(float64)
		if !ok {
			// err := errors.New("active is nil")
			// s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			// return nil, status.Error(codes.Internal, err.Error())
			active = 0
		}

		user, ok := usersMap[userId]
		if !ok {
			err := errors.New("user is nil")
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		name, ok := userItem["name"].(string)
		if ok {
			user.Name = name
		}
		user.Active = int32(active)
		user.ProjectId = projectId
		user.RoleId = roleId
		user.ClientTypeId = clientTypeId

		resp.Users = append(resp.Users, user)
	}

	return resp, nil

}

func (s *userService) V2UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	s.log.Info("---V2UpdateUser--->", logger.Any("req", req))

	//emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	//email := emailRegex.MatchString(req.Email)
	//if !email {
	//	err = fmt.Errorf("email is not valid")
	//	s.log.Error("!!!UpdateUser--->", logger.Error(err))
	//	return nil, err
	//}
	//
	//phoneRegex := regexp.MustCompile(`^[+]?(\d{1,2})?[\s.-]?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$`)
	//phone := phoneRegex.MatchString(req.Phone)
	//if !phone {
	//	err = fmt.Errorf("phone number is not valid")
	//	s.log.Error("!!!UpdateUser--->", logger.Error(err))
	//	return nil, err
	//}
	//
	//if err != nil {
	//	s.log.Error("!!!UpdateUser--->", logger.Error(err))
	//	return nil, status.Error(codes.InvalidArgument, err.Error())
	//}
	//
	//result, err := s.services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(ctx, &pbObject.CommonMessage{
	//	TableSlug: "user",
	//	Data:      structData,
	//	ProjectId: config.UcodeDefaultProjectID,
	//})
	//if err != nil {
	//	s.log.Error("!!!UpdateUser.ObjectBuilderService.GetSingle--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//return &pb.CommonMessage{
	//	TableSlug: result.TableSlug,
	//	Data:      result.Data,
	//}, nil

	rowsAffected, err := s.strg.User().Update(ctx, req)

	if err != nil {
		s.log.Error("!!!V2UpdateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, "no rows were affected")
	}

	userProject, err := s.strg.User().UpdateUserToProject(
		ctx,
		&pb.AddUserToProjectReq{
			UserId:       req.Id,
			CompanyId:    req.CompanyId,
			ProjectId:    req.ProjectId,
			ClientTypeId: req.ClientTypeId,
			RoleId:       req.RoleId,
			EnvId:        req.EnvironmentId,
		},
	)
	if err != nil {
		s.log.Error("!!!V2UpdateUser Update user project--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if userProject.UserId == "" {
		s.log.Error("!!!V2UpdateUser user project not update", logger.Error(err))
	}

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	structData, err := helper.ConvertRequestToSturct(req)
	if err != nil {
		s.log.Error("!!!V2UpdateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	structData.Fields["from_auth_service"] = structpb.NewBoolValue(true)
	var tableSlug = "user"
	switch req.GetResourceType() {
	case 1:
		clientType, err := services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(context.Background(), &pbObject.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2UpdateUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		_, err = services.GetObjectBuilderServiceByType(req.NodeType).Update(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!UpdateUser.ObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		clientType, err := services.GoItemService().GetSingle(context.Background(), &nb.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2GetUserSingle--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		_, err = services.GoItemService().Update(ctx, &nb.CommonMessage{
			TableSlug: tableSlug,
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2UpdateUser.PostgresObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	res, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{Id: req.Id})
	if err != nil {
		s.log.Error("!!!V2UpdateUser.GetByPK--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return res, nil
}

func (s *userService) V2DeleteUser(ctx context.Context, req *pb.UserPrimaryKey) (*emptypb.Empty, error) {
	s.log.Info("---V2DeleteUser--->", logger.Any("req", req))

	res := &emptypb.Empty{}
	responseFromDeleteUser := &pbObject.CommonMessage{}

	// _, err := s.strg.User().Delete(ctx, req)
	// if err != nil {
	// 	s.log.Error("!!!V2DeleteUser--->", logger.Error(err))
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }
	// return res, nil

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	var tableSlug = "user"
	switch req.GetResourceType() {
	case 1:
		clientType, err := services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(context.Background(), &pbObject.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2DeleteUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug := cast.ToString(response["table_slug"])
			if clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		responseFromDeleteUser, err = services.GetObjectBuilderServiceByType(req.NodeType).Delete(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id":                structpb.NewStringValue(req.Id),
					"from_auth_service": structpb.NewBoolValue(true),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2DeleteUser.ObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		_, err = s.strg.User().DeleteUserFromProject(context.Background(), &pb.DeleteSyncUserRequest{
			UserId:       req.GetId(),
			ProjectId:    req.GetProjectId(),
			CompanyId:    req.GetCompanyId(),
			ClientTypeId: req.GetClientTypeId(),
			RoleId:       responseFromDeleteUser.Data.AsMap()["role_id"].(string),
		})
		if err != nil {
			s.log.Error("!!!V2DeleteUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		_, err = s.strg.User().Delete(ctx, req)
		if err != nil {
			s.log.Error("!!!V2DeleteUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		clientType, err := services.PostgresObjectBuilderService().GetSingle(context.Background(), &pbObject.CommonMessage{
			TableSlug: "client_type",
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id": structpb.NewStringValue(req.GetClientTypeId()),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2DeleteUser--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		response, ok := clientType.Data.AsMap()["response"].(map[string]interface{})
		if ok {
			clientTypeTableSlug, ok := response["table_slug"].(string)
			if ok && clientTypeTableSlug != "" {
				tableSlug = clientTypeTableSlug
			}
		}
		responseFromDeleteUser, err = services.PostgresObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
			TableSlug: tableSlug,
			Data: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"id":                structpb.NewStringValue(req.Id),
					"from_auth_service": structpb.NewBoolValue(true),
				},
			},
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!V2DeleteUser.PostgresObjectBuilderService.Update--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		_, err = s.strg.User().DeleteUserFromProject(context.Background(), &pb.DeleteSyncUserRequest{
			UserId:       req.GetId(),
			ProjectId:    req.GetProjectId(),
			CompanyId:    req.GetCompanyId(),
			ClientTypeId: req.GetClientTypeId(),
			RoleId:       responseFromDeleteUser.Data.AsMap()["role_id"].(string),
		})
		if err != nil {
			s.log.Error("!!!V2DeleteUser.PostgresObjectBuilderService.DeleteUserProject--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	_, err = s.strg.User().Delete(ctx, req)
	if err != nil {
		s.log.Error("!!!V2DeleteUser--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return res, nil
}

func (s *userService) AddUserToProject(ctx context.Context, req *pb.AddUserToProjectReq) (*pb.AddUserToProjectRes, error) {
	s.log.Info("AddUserToProject", logger.Any("req", req))

	res, err := s.strg.User().AddUserToProject(ctx, req)
	if err != nil {
		errUserAdd := config.ErrUserAlradyMember
		s.log.Error("cant add project to user", logger.Error(err))
		return nil, status.Error(codes.Internal, errUserAdd.Error())
	}

	return res, nil
}

func (s *userService) GetProjectsByUserId(ctx context.Context, req *pb.GetProjectsByUserIdReq) (*pb.GetProjectsByUserIdRes, error) {
	s.log.Info("GetProjectsByUserId", logger.Any("req", req))

	res, err := s.strg.User().GetProjectsByUserId(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *userService) V2GetUserByLoginTypes(ctx context.Context, req *pb.GetUserByLoginTypesRequest) (*pb.GetUserByLoginTypesResponse, error) {
	s.log.Info("GetProjectsByUserId", logger.Any("req", req))
	fmt.Println("coming here to V2GetUserByLoginTypes >>> ", req)
	res, err := s.strg.User().GetUserByLoginType(ctx, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *userService) GetUserByUsername(ctx context.Context, req *pb.GetUserByUsernameRequest) (*pb.User, error) {
	s.log.Info("GetUserByUsername -> ", logger.Any("req: ", req))
	res, err := s.strg.User().GetByUsername(ctx, req.GetUsername())
	if err != nil {
		return nil, err
	}
	s.log.Info("GetUserByUsername <- ", logger.Any("res: ", res))
	return res, nil
}
func (s *userService) GetUserProjects(ctx context.Context, req *pb.UserPrimaryKey) (*pb.GetUserProjectsRes, error) {

	userProjects, err := s.strg.User().GetUserProjects(ctx, req.Id)
	if err != nil {
		errGetProjects := errors.New("cant get user projects")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, errGetProjects.Error())
	}

	return userProjects, nil
}

func (s *userService) V2ResetPassword(ctx context.Context, req *pb.V2UserResetPasswordRequest) (*pb.User, error) {

	var (
		user             = &pb.User{}
		err              error
		unHashedPassword = req.GetPassword()
	)
	if len(req.GetPassword()) > 6 {
		user, err = s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
			Id: req.UserId,
		})
		if err != nil {
			return nil, err
		}
		match, err := security.ComparePassword(user.GetPassword(), req.OldPassword)
		if err != nil {
			return nil, err
		}
		if !match {
			err := errors.New("wrong old password")
			s.log.Error("!!!V2UserResetPassword--->", logger.Error(err))
			return nil, err
		}
		hashedPassword, err := security.HashPassword(req.Password)
		if err != nil {
			s.log.Error("!!!V2UserResetPassword--->", logger.Error(err))
			return nil, err
		}
		req.Password = hashedPassword
		rowsAffected, err := s.strg.User().V2ResetPassword(ctx, &pb.V2ResetPasswordRequest{
			UserId:   req.UserId,
			Password: req.Password,
		})
		if err != nil {
			s.log.Error("!!!V2UserResetPassword--->", logger.Error(err))
			return nil, err
		}
		if rowsAffected <= 0 {
			return nil, status.Error(codes.InvalidArgument, "no rows were affected")
		}
		user.Password = hashedPassword

		services, err := s.serviceNode.GetByNodeType(
			req.ProjectId,
			req.NodeType,
		)
		if err != nil {
			return nil, err
		}

		if req.GetClientTypeId() != "" && req.GetEnvironmentId() != "" && req.GetProjectId() != "" {
			resource, err := s.services.ServiceResource().GetSingle(ctx, &company_service.GetSingleServiceResourceReq{
				ProjectId:     req.GetProjectId(),
				EnvironmentId: req.GetEnvironmentId(),
				ServiceType:   company_service.ServiceType_BUILDER_SERVICE,
			})
			if err != nil {
				err = errors.New("password updated in auth but not found resource in this project")
				s.log.Error("!!!V2UserResetPassword--->", logger.Error(err))
				return nil, err
			}
			switch resource.ResourceType {
			case 1:
				_, err = services.GetLoginServiceByType(resource.NodeType).UpdateUserPassword(ctx, &pbObject.UpdateUserPasswordRequest{
					Guid:                  req.UserId,
					ResourceEnvironmentId: resource.ResourceEnvironmentId,
					Password:              unHashedPassword,
					ClientTypeId:          req.ClientTypeId,
				})
				if err != nil {
					err = errors.New("password updated in auth but failed to update in object builder")
					s.log.Error("!!!V2UserResetPassword.GetLoginServiceByType(resource.NodeType).UpdateUserPassword--->", logger.Error(err))
					return nil, err
				}
			case 3:
				_, err = services.PostgresLoginService().UpdateUserPassword(ctx, &pbObject.UpdateUserPasswordRequest{
					Guid:                  req.UserId,
					ResourceEnvironmentId: resource.ResourceEnvironmentId,
					Password:              unHashedPassword,
					ClientTypeId:          req.ClientTypeId,
				})
				if err != nil {
					err = errors.New("password updated in auth but failed to update in object builder")
					s.log.Error("!!!V2UserResetPassword.PostgresLoginService().UpdateUserPassword--->", logger.Error(err))
					return nil, err
				}
			}
		}
	} else {
		err := fmt.Errorf("password must not be less than 6 characters")
		s.log.Error("!!!V2UserResetPassword--->", logger.Error(err))
		return nil, err
	}
	return user, nil
}
