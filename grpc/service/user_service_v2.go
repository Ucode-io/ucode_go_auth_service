package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"ucode/ucode_go_auth_service/config"
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

func (s *userService) V2CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	s.log.Info("---CreateUser--->", logger.Any("req", req))

	if len(req.Login) < 6 {
		err := fmt.Errorf("login must not be less than 6 characters")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(req.Password) < 6 {
		err := fmt.Errorf("password must not be less than 6 characters")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, err
	}

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

	phoneRegex := regexp.MustCompile(`^[+]?(\d{1,2})?[\s.-]?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$`)
	phone := phoneRegex.MatchString(req.Phone)
	if !phone {
		err = fmt.Errorf("phone number is not valid")
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, err
	}

	pKey, err := s.strg.User().Create(ctx, req)

	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// objectBuilder -> auth service
	structData, err := helper.ConvertRequestToSturct(map[string]interface{}{
		"guid":               pKey.Id,
		"project_id":         req.ProjectId,
		"role_id":            req.RoleId,
		"client_type_id":     req.ClientTypeId,
		"client_platform_id": req.ClientPlatformId,
		"active":             req.Active,
		"expires_at":         req.ExpiresAt,
	})
	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = s.services.ObjectBuilderService().Create(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
		ProjectId: config.UcodeDefaultProjectID,
	})

	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.strg.User().AddUserToProject(ctx, &pb.AddUserToProjectReq{
		UserId:    pKey.Id,
		ProjectId: req.ProjectId,
		CompanyId: req.CompanyId,
	})

	if err != nil {
		s.log.Error("!!!CreateUser--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.strg.User().GetByPK(ctx, pKey)
}

func (s *userService) V2GetUserByID(ctx context.Context, req *pb.UserPrimaryKey) (*pb.User, error) {
	s.log.Info("---GetUserByID--->", logger.Any("req", req))

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

	result, err := s.services.ObjectBuilderService().GetSingle(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
		ProjectId: req.ProjectId,
	})
	if err != nil {
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

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
		err := errors.New("client_platform_id is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
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
		err := errors.New("active is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	user.Active = int32(active)

	projectId, ok := userData["project_id"].(string)
	if !ok {
		err := errors.New("projectId is nil")
		s.log.Error("!!!GetUserByID.ObjectBuilderService.GetSingle--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	user.ProjectId = projectId

	return user, nil
}

func (s *userService) V2GetUserList(ctx context.Context, req *pb.GetUserListRequest) (*pb.GetUserListResponse, error) {
	s.log.Info("---GetUserList--->", logger.Any("req", req))

	resp := &pb.GetUserListResponse{}

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

	usersResp, err := s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
		TableSlug: "user",
		Data:      structData,
		ProjectId: req.ProjectId,
	})
	if err != nil {
		s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
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

		clientPlatformId, ok := userItem["client_platform_id"].(string)
		if !ok {
			err := errors.New("clientPlatformId is nil")
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		projectId, ok := userItem["project_id"].(string)
		if !ok {
			err := errors.New("projectId is nil")
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		active, ok := userItem["active"].(float64)
		if !ok {
			err := errors.New("active is nil")
			s.log.Error("!!!GetUserList.ObjectBuilderService.GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
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
		user.ClientPlatformId = clientPlatformId
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

	emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	email := emailRegex.MatchString(req.Email)
	if !email {
		err = fmt.Errorf("email is not valid")
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, err
	}

	phoneRegex := regexp.MustCompile(`^[+]?(\d{1,2})?[\s.-]?\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}$`)
	phone := phoneRegex.MatchString(req.Phone)
	if !phone {
		err = fmt.Errorf("phone number is not valid")
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, err
	}

	res, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{Id: req.Id})
	if err != nil {
		s.log.Error("!!!UpdateUser--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return res, err
}

func (s *userService) V2DeleteUser(ctx context.Context, req *pb.UserPrimaryKey) (*emptypb.Empty, error) {
	s.log.Info("---DeleteUser--->", logger.Any("req", req))

	//res := &emptypb.Empty{}
	//
	//structData, err := helper.ConvertRequestToSturct(req)
	//if err != nil {
	//	s.log.Error("!!!DeleteUser--->", logger.Error(err))
	//	return nil, status.Error(codes.InvalidArgument, err.Error())
	//}
	//
	//_, err = s.services.ObjectBuilderService().Delete(ctx, &pbObject.CommonMessage{
	//	TableSlug: "user",
	//	Data:      structData,
	//	ProjectId: config.UcodeDefaultProjectID,
	//})
	//if err != nil {
	//	s.log.Error("!!!DeleteUser.ObjectBuilderService.Delete--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//return res, nil

	res := &emptypb.Empty{}

	rowsAffected, err := s.strg.User().Delete(ctx, req)

	if err != nil {
		s.log.Error("!!!DeleteUser--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, "no rows were affected")
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
