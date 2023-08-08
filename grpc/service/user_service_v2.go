package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/saidamir98/udevs_pkg/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

	if foundUser.Id == "" {
		pKey, err := s.strg.User().Create(ctx, &auth_service.CreateUserRequest{
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
			_, err = s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
				TableSlug: "user",
				Data:      structData,
				ProjectId: req.GetResourceEnvironmentId(),
			})
			if err != nil {
				s.log.Error("!!!ObjectBuilderService.CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		case 3:
			_, err = s.services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
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
				objUser, err = s.services.LoginService().LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

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
				objUser, err = s.services.PostgresLoginService().LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

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
				_, err = s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
					TableSlug: "user",
					Data:      structData,
					ProjectId: req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!CreateUser--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			case 3:
				_, err = s.services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
					TableSlug: "user",
					Data:      structData,
					ProjectId: req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}

			}

			// _, err = s.strg.User().AddUserToProject(ctx, &pb.AddUserToProjectReq{
			// 	UserId:    foundUser.Id,
			// 	ProjectId: req.GetProjectId(),
			// 	CompanyId: req.GetCompanyId(),
			// })
			// if err != nil {
			// 	s.log.Error("!!!CreateUser--->", logger.Error(err))
			// 	return nil, status.Error(codes.Internal, err.Error())
			// }

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

	foundUser, err := s.strg.User().GetByUsername(ctx, req.Email)
	if foundUser.Id == "" {
		foundUser, err = s.strg.User().GetByUsername(ctx, req.Phone)
	}

	if foundUser.Id == "" {
		pKey, err := s.strg.User().Create(ctx, &auth_service.CreateUserRequest{
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
			_, err = s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
				TableSlug: "user",
				Data:      structData,
				ProjectId: req.GetResourceEnvironmentId(),
			})
			if err != nil {
				s.log.Error("!!!CreateUser--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		case 3:
			_, err = s.services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
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
				objUser, err = s.services.LoginService().LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

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
				objUser, err = s.services.PostgresLoginService().LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

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
				objUser, err = s.services.LoginService().LoginWithOtp(context.Background(), &pbObject.PhoneOtpRequst{

					PhoneNumber: req.Phone,
					ClientType:  "WEB_USER",
					ProjectId:   req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!Found user from obj--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			case 3:
				objUser, err = s.services.PostgresLoginService().LoginWithOtp(context.Background(), &pbObject.PhoneOtpRequst{

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
				_, err = s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
					TableSlug: "user",
					Data:      structData,
					ProjectId: req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!CreateUser--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			case 3:
				_, err = s.services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
					TableSlug: "user",
					Data:      structData,
					ProjectId: req.GetResourceEnvironmentId(),
				})
				if err != nil {
					s.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}

			}

			// _, err = s.strg.User().AddUserToProject(ctx, &pb.AddUserToProjectReq{
			// 	UserId:    foundUser.Id,
			// 	ProjectId: req.GetProjectId(),
			// 	CompanyId: req.GetCompanyId(),
			// })
			// if err != nil {
			// 	s.log.Error("!!!CreateUser--->", logger.Error(err))
			// 	return nil, status.Error(codes.Internal, err.Error())
			// }

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
	s.log.Info("---CreateUser--->", logger.Any("req", req))

	// if len(req.Login) < 6 {
	// 	err := fmt.Errorf("login must not be less than 6 characters")
	// 	s.log.Error("!!!CreateUser--->", logger.Error(err))
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	// if len(req.Password) < 6 {
	// 	err := fmt.Errorf("password must not be less than 6 characters")
	// 	s.log.Error("!!!CreateUser--->", logger.Error(err))
	// 	return nil, err
	// }

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	req.Password = hashedPassword

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := emailRegex.MatchString(req.Email)
	if !email && req.Email != "" {
		err = fmt.Errorf("email is not valid")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, err
	}

	// phoneRegex := regexp.MustCompile(`^[+]?(\d{1,2})?[\s.-]?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$`)
	// phone := phoneRegex.MatchString(req.Phone)
	// if !phone {
	// 	err = fmt.Errorf("phone number is not valid")
	// 	s.log.Error("!!!CreateUser--->", logger.Error(err))
	// 	return nil, err
	// }

	pKey, err := s.strg.User().Create(ctx, req)

	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
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
	})
	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	switch req.ResourceType {
	case 1:
		_, err = s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})

		if err != nil {
			s.log.Error("!!!CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		_, err = s.services.PostgresObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})

		if err != nil {
			s.log.Error("!!!PostgresObjectBuilderService.CreateUser--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
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

	return s.strg.User().GetByPK(ctx, pKey)
}

func (s *userService) V2GetUserByID(ctx context.Context, req *pb.UserPrimaryKey) (*pb.User, error) {
	s.log.Info("---GetUserByID--->", logger.Any("req", req))

	var (
		result *pbObject.CommonMessage
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

	switch req.ResourceType {
	case 1:

		result, err = s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: req.ProjectId,
		})
		if err != nil {
			s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		result, err = s.services.PostgresObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: req.ProjectId,
		})
		if err != nil {
			s.log.Error("!!!GetUserByID.PostgresObjectBuilderService.GetSingle--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

	}
	log.Println("Data:   ", result)
	userData, ok := result.Data.AsMap()["response"].(map[string]interface{})

	if !ok {
		err := errors.New("userData is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if bytes, err := json.Marshal(userData); err == nil {
		fmt.Println("userdata", string(bytes))
	}

	roleId, ok := userData["role_id"].(string)
	if !ok {
		err := errors.New("role_id is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	user.RoleId = roleId

	clientPlatformId, ok := userData["client_platform_id"].(string)
	if !ok {
		// err := errors.New("client_platform_id is nil")
		// s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		// return nil, status.Error(codes.Internal, err.Error())
		clientPlatformId = ""
	}

	user.ClientPlatformId = clientPlatformId

	clientTypeId, ok := userData["client_type_id"].(string)
	if !ok {
		err := errors.New("client_type_id is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	user.ClientTypeId = clientTypeId

	active, ok := userData["active"].(float64)
	if !ok {
		// err := errors.New("active is nil")
		// s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		// return nil, status.Error(codes.Internal, err.Error())
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

	user.ProjectId = projectId

	return user, nil
}

func (s *userService) V2GetUserList(ctx context.Context, req *pb.GetUserListRequest) (*pb.GetUserListResponse, error) {
	s.log.Info("---GetUserList--->", logger.Any("req", req))

	resp := &pb.GetUserListResponse{}
	var (
		usersResp *pbObject.CommonMessage
	)

	// res, err := s.strg.User().GetList(ctx, req)

	// if err != nil {
	// 	s.log.Error("!!!GetUserList--->", logger.Error(err))
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	userIds, err := s.strg.User().GetUserIds(ctx, req)
	if err != nil {
		s.log.Error("!!!GetUserList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	users, err := s.strg.User().GetListByPKs(ctx, &pb.UserPrimaryKeyList{
		Ids: *userIds,
	})
	if err != nil {
		s.log.Error("!!!GetUserList--->", logger.Error(err))
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
		s.log.Error("!!!GetUserList--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	switch req.ResourceType {
	case 1:
		usersResp, err = s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case 3:
		usersResp, err = s.services.PostgresObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: req.GetResourceEnvironmentId(),
		})
		if err != nil {
			s.log.Error("!!!GetUserList.PostgresObjectBuilderService.GetList--->", logger.Error(err))
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

		user.Active = int32(active)
		user.RoleId = roleId
		user.ClientTypeId = clientTypeId
		// user.ClientPlatformId = clientPlatformId
		user.ProjectId = projectId

		resp.Users = append(resp.Users, user)
	}

	return resp, nil

}

func (s *userService) V2UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	s.log.Info("---UpdateUser--->", logger.Any("req", req))

	//structData, err := helper.ConvertRequestToSturct(req)
	//if err != nil {
	//	s.log.Error("!!!UpdateUser--->", logger.Error(err))
	//	return nil, status.Error(codes.InvalidArgument, err.Error())
	//}
	//
	//_, err = s.services.ObjectBuilderService().Update(ctx, &pbObject.CommonMessage{
	//	TableSlug: "user",
	//	Data:      structData,
	//	ProjectId: config.UcodeDefaultProjectID,
	//})
	//if err != nil {
	//	s.log.Error("!!!UpdateUser.ObjectBuilderService.Update--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
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
	//result, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
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
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, "no rows were affected")
	}

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

	res, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{Id: req.Id})
	if err != nil {
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return res, err
}

func (s *userService) V2DeleteUser(ctx context.Context, req *pb.UserPrimaryKey) (*emptypb.Empty, error) {
	s.log.Info("---DeleteUser--->", logger.Any("req", req))

	res := &emptypb.Empty{}
	//structData, err := helper.ConvertRequestToSturct(req)
	//if err != nil {
	//	s.log.Error("!!!DeleteUser--->", logger.Error(err))
	//	return nil, status.Error(codes.InvalidArgument, err.Error())
	//}
	//
	//_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
	//	TableSlug: "user",
	//	Data:      structData,
	//	ProjectId: req.GetResourceEnvironmentId(),
	//})
	//if err != nil {
	//	s.log.Error("!!!DeleteUser.ObjectBuilderService.Delete--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	_, err := s.strg.User().Delete(ctx, req)

	if err != nil {
		s.log.Error("!!!DeleteUser--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// if rowsAffected <= 0 {
	// 	return nil, status.Error(codes.InvalidArgument, "no rows were affected")
	// }

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

	res, err := s.strg.User().GetUserByLoginType(ctx, req)
	if err != nil {
		return nil, err
	}

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

func (s *userService) GetUserByUsername(ctx context.Context, req *auth_service.GetUserByUsernameRequest) (*pb.User, error) {
	s.log.Info("GetUserByUsername -> ", logger.Any("req: ", req))
	res, err := s.strg.User().GetByUsername(ctx, req.GetUsername())
	if err != nil {
		return nil, err
	}
	s.log.Info("GetUserByUsername <- ", logger.Any("res: ", res))
	return res, nil
}

func (s *userService) V2UserResetPassword(ctx context.Context, req *pb.V2UserResetPasswordRequest) (*pb.User, error) {

	var (
		user = &pb.User{}
		err  error
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
	} else {
		err := fmt.Errorf("password must not be less than 6 characters")
		s.log.Error("!!!V2UserResetPassword--->", logger.Error(err))
		return nil, err
	}
	return user, nil
}
