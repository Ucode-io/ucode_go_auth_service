package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	secure "ucode/ucode_go_auth_service/pkg/security"

	"github.com/google/uuid"
	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/saidamir98/udevs_pkg/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *sessionService) V2Login(ctx context.Context, req *pb.V2LoginRequest) (*pb.V2LoginResponse, error) { //todo

	if len(req.Username) < 6 {
		err := errors.New("invalid username")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(req.Password) < 6 {
		err := errors.New("invalid password")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	data, err := s.services.LoginService().Login(
		ctx,
		&pbObject.LoginRequest{
			Password:      req.Password,
			Login:         req.Username,
			ClientType:    req.ClientType,
			LoginStrategy: req.LoginStrategy,
			ProjectId:     config.UcodeDefaultProjectID,
		},
	)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	expiresAt, err := time.Parse(config.DatabaseTimeLayout, "2023-01-28T06:23:33.952Z")
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if expiresAt.Unix() < time.Now().Unix() {
		err := errors.New("User has been expired")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if !data.UserFound {
		customError := errors.New("User not found")
		s.log.Error("!!!Login--->", logger.Error(customError))
		return nil, status.Error(codes.NotFound, customError.Error())
	}

	projectID := req.ProjectId

	_, err = uuid.Parse(req.GetProjectId())
	if err == nil {
		project, err := s.services.ProjectServiceClient().GetById(
			ctx,
			&company_service.GetProjectByIdRequest{
				ProjectId: req.ProjectId,
			},
		)
		if err != nil {
			s.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		company, err := s.services.CompanyServiceClient().GetById(
			ctx,
			&company_service.GetCompanyByIdRequest{
				Id: project.CompanyId,
			},
		)
		if err != nil {
			s.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		// if user is not company owner
		if company.Company.GetOwnerId() != data.GetUserId() {
			// if user has no access to project
			if req.GetProjectId() != data.Role.GetProjectId() {
				s.log.Error("!!!Login--->", logger.Any("msg", "user has no access to this project"))
				return nil, status.Error(codes.Internal, "user has no access to this project")
			}
		}
	}

	res := helper.ConvertPbToAnotherPb(data)

	resp, err := s.SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
		LoginData: res,
		Tables:    req.Tables,
		ProjectId: projectID,
	})
	if resp == nil {
		err := errors.New("User Not Found")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return resp, nil
}

func (s *sessionService) V2LoginSuperAdmin(ctx context.Context, req *pb.V2LoginSuperAdminReq) (*pb.V2LoginSuperAdminRes, error) {
	if len(req.Username) < 6 {
		err := errors.New("invalid username")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(req.Password) < 6 {
		err := errors.New("invalid password")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user, err := s.strg.User().GetByUsername(ctx, req.GetUsername())
	if err != nil {
		s.log.Error("!!!SuperAdminLogin--->", logger.Error(err))
		if err == sql.ErrNoRows {
			customError := errors.New("User not found")
			return nil, status.Error(codes.NotFound, customError.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	match, err := security.ComparePassword(user.Password, req.Password)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !match {
		err := errors.New("username or password is wrong")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	//fmt.Println(":::::::::user.GetExpiresAt():::::::::", user.GetExpiresAt())

	expiresAt, err := time.Parse(config.DatabaseTimeLayout, user.GetExpiresAt())
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if expiresAt.Unix() < time.Now().Unix() {
		err := errors.New("User has been expired")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.SessionAndTokenGeneratorSuperAdmin(ctx, &pb.SessionAndTokenRequest{
		LoginData: &pb.V2LoginResponse{
			UserFound:      true,
			ClientPlatform: &pb.ClientPlatform{},
			ClientType:     &pb.ClientType{},
			Role:           &pb.Role{},
			UserId:         user.GetId(),
			Permissions:    []*pb.RecordPermission{},
			Sessions:       []*pb.Session{},
			LoginTableSlug: "",
			AppPermissions: []*pb.RecordPermission{},
		},
		ProjectId: "",
	})
	if resp == nil {
		err := errors.New("User Not Found")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.V2LoginSuperAdminRes{
		UserFound: true,
		UserId:    user.GetId(),
		Token:     resp.Token,
		Sessions:  resp.Sessions,
	}, nil
}

func (s *sessionService) V2HasAccess(ctx context.Context, req *pb.HasAccessRequest) (*pb.HasAccessResponse, error) {

	tokenInfo, err := secure.ParseClaims(req.AccessToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if session.IsChanged {
		err := errors.New("permision update")
		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	expiresAt, err := time.Parse(config.DatabaseTimeLayout, session.ExpiresAt)
	if err != nil {
		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if expiresAt.Unix() < time.Now().Unix() {
		err := errors.New("User has been expired")
		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	_, err = s.strg.Scope().Upsert(ctx, &pb.UpsertScopeRequest{
		ClientPlatformId: session.ClientPlatformId,
		Path:             req.Path,
		Method:           req.Method,
	})
	if err != nil {
		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	var methodField string
	switch req.Method {
	case "GET":
		methodField = "read"
	case "POST":
		methodField = "write"
	case "PUT":
		methodField = "update"
	case "DELETE":
		methodField = "delete"
	}

	splitedPath := strings.Split(req.Path, "/")
	splitedPath = splitedPath[1:]

	var tableSlug string
	tableSlug = splitedPath[len(splitedPath)-1]
	if tableSlug[len(tableSlug)-2:] == "id" {
		tableSlug = splitedPath[len(splitedPath)-2]
	}

	if _, ok := config.ObjectBuilderTableSlugs[tableSlug]; ok {
		tableSlug = "app"
	}

	request := make(map[string]interface{})
	request["client_type_id"] = session.ClientTypeId
	request[methodField] = "Yes"
	request["table_slug"] = tableSlug

	clientType, err := s.services.ClientService().V2GetClientTypeByID(ctx, &pb.ClientTypePrimaryKey{
		Id: session.ClientTypeId,
	})
	if err != nil {
		s.log.Error("!!!V2HasAccess.ClientService.V2GetClientTypeByID--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	convertedClientType, err := helper.ConvertStructToResponse(clientType.Data)
	if err != nil {
		s.log.Error("!!!V2HasAccess.ConvertStructToResponse--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	clientName, ok := convertedClientType["response"].(map[string]interface{})["name"]
	if !ok {
		res := make(map[string]interface{})
		resp := &pbObject.CommonMessage{}

		if clientName == nil {
			err := errors.New("Wrong client type")
			s.log.Error("!!!V2HasAccess--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		structPb, err := helper.ConvertMapToStruct(request)
		if err != nil {
			s.log.Error("!!!V2HasAccess--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		if session.ClientTypeId != config.AdminClientPlatformID || clientName.(string) != config.AdminClientName {
			resp, err = s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
				TableSlug: "record_permission",
				Data:      structPb,
				ProjectId: config.UcodeDefaultProjectID,
			})
			if err != nil {
				s.log.Error("!!!V2HasAccess.ObjectBuilderService.GetList--->", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			res, err = helper.ConvertStructToResponse(resp.Data)
			if err != nil {
				s.log.Error("!!!V2HasAccess.ConvertStructToResponse--->", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			if len(res["response"].([]interface{})) == 0 {
				err := errors.New("Permission denied")
				s.log.Error("!!!V2HasAccess--->", logger.Error(err))
				return nil, status.Error(codes.PermissionDenied, err.Error())
			}
		}
	}

	// DONT FORGET TO UNCOMMENT THIS!!!

	// hasAccess, err := s.strg.PermissionScope().HasAccess(ctx, user.RoleId, req.ClientPlatformId, req.Path, req.Method)
	// if err != nil {
	// 	s.log.Error("!!!V2HasAccess--->", logger.Error(err))
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	// if !hasAccess {
	// 	err = errors.New("access denied")
	// 	s.log.Error("!!!V2HasAccess--->", logger.Error(err))
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	var authTables []*pb.TableBody
	for _, table := range tokenInfo.Tables {
		authTable := &pb.TableBody{
			TableSlug: table.TableSlug,
			ObjectId:  table.ObjectID,
		}
		authTables = append(authTables, authTable)
	}

	return &pb.HasAccessResponse{
		Id:               session.Id,
		ProjectId:        session.ProjectId,
		ClientPlatformId: session.ClientPlatformId,
		ClientTypeId:     session.ClientTypeId,
		UserId:           session.UserId,
		RoleId:           session.RoleId,
		Ip:               session.Ip,
		Data:             session.Data,
		ExpiresAt:        session.ExpiresAt,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		Tables:           authTables,
	}, nil
}

func (s *sessionService) V2RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2RefreshTokenResponse, error) {

	tokenInfo, err := secure.ParseClaims(req.RefreshToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, err = s.strg.Session().UpdateByRoleId(ctx, &pb.UpdateSessionByRoleIdRequest{
		RoleId:    tokenInfo.RoleID,
		IsChanged: false,
	})
	if err != nil {
		s.log.Error("!!!RefreshToken.UpdateByRoleId--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userData, err := s.services.LoginService().GetUserUpdatedPermission(ctx, &pbObject.GetUserUpdatedPermissionRequest{
		ClientTypeId: session.ClientTypeId,
		UserId:       session.UserId,
	})
	if err != nil {
		s.log.Error("!!!V2HasAccess.SessionService().GetUserUpdatedPermission--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	convertedData := helper.ConvertPbToAnotherPb(userData)

	authTables := []*pb.TableBody{}
	if tokenInfo.Tables != nil {
		for _, table := range tokenInfo.Tables {
			authTable := &pb.TableBody{
				TableSlug: table.TableSlug,
				ObjectId:  table.ObjectID,
			}
			authTables = append(authTables, authTable)
		}
	}

	// TODO - wrap in a function
	m := map[string]interface{}{
		"id":                 session.Id,
		"project_id":         session.ProjectId,
		"client_platform_id": session.ClientPlatformId,
		"client_type_id":     session.ClientTypeId,
		"user_id":            session.UserId,
		"role_id":            session.RoleId,
		"ip":                 session.Data,
		"data":               session.Data,
		"tables":             authTables,
	}

	accessToken, err := security.GenerateJWT(m, config.AccessTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken, err := security.GenerateJWT(m, config.RefreshTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	token := &pb.Token{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		ExpiresAt:        session.ExpiresAt,
		RefreshInSeconds: int32(config.AccessTokenExpiresInTime.Seconds()),
	}
	res := &pb.V2RefreshTokenResponse{
		Token:       token,
		Permissions: convertedData.Permissions,
	}

	return res, nil
}

func (s *sessionService) V2RefreshTokenSuperAdmin(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2RefreshTokenSuperAdminResponse, error) {

	tokenInfo, err := secure.ParseClaims(req.RefreshToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	_, err = s.strg.Session().UpdateBySessionId(ctx, &pb.UpdateSessionBySessionIdRequest{
		Id:        tokenInfo.ID,
		IsChanged: false,
	})
	if err != nil {
		s.log.Error("!!!RefreshToken.UpdateByRoleId--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	//userData, err := s.services.LoginService().GetUserUpdatedPermission(ctx, &pbObject.GetUserUpdatedPermissionRequest{
	//	ClientTypeId: session.ClientTypeId,
	//	UserId:       session.UserId,
	//})
	//if err != nil {
	//	s.log.Error("!!!V2HasAccess.SessionService().GetUserUpdatedPermission--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//convertedData := helper.ConvertPbToAnotherPb(userData)

	authTables := []*pb.TableBody{}
	if tokenInfo.Tables != nil {
		for _, table := range tokenInfo.Tables {
			authTable := &pb.TableBody{
				TableSlug: table.TableSlug,
				ObjectId:  table.ObjectID,
			}
			authTables = append(authTables, authTable)
		}
	}

	// TODO - wrap in a function
	m := map[string]interface{}{
		"id":                 session.Id,
		"project_id":         session.ProjectId,
		"client_platform_id": session.ClientPlatformId,
		"client_type_id":     session.ClientTypeId,
		"user_id":            session.UserId,
		"role_id":            session.RoleId,
		"ip":                 session.Data,
		"data":               session.Data,
		"tables":             authTables,
	}

	accessToken, err := security.GenerateJWT(m, config.AccessTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken, err := security.GenerateJWT(m, config.RefreshTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	token := &pb.Token{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		ExpiresAt:        session.ExpiresAt,
		RefreshInSeconds: int32(config.AccessTokenExpiresInTime.Seconds()),
	}
	res := &pb.V2RefreshTokenSuperAdminResponse{
		Token: token,
	}

	return res, nil
}

func (s *sessionService) SessionAndTokenGenerator(ctx context.Context, input *pb.SessionAndTokenRequest) (*pb.V2LoginResponse, error) {

	// // TODO - Delete all old sessions & refresh token has this function too
	rowsAffected, err := s.strg.Session().DeleteExpiredUserSessions(ctx, input.LoginData.UserId)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.log.Info("Login--->DeleteExpiredUserSessions", logger.Any("rowsAffected", rowsAffected))
	userSessionList, err := s.strg.Session().GetSessionListByUserID(ctx, input.LoginData.UserId)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	input.LoginData.Sessions = userSessionList.Sessions

	_, err = uuid.Parse(input.ProjectId)
	if err != nil {
		input.ProjectId = "f5955c82-f264-4655-aeb4-86fd1c642cb6"
	}

	sessionPKey, err := s.strg.Session().Create(ctx, &pb.CreateSessionRequest{
		ProjectId:        input.ProjectId,
		ClientPlatformId: input.LoginData.ClientPlatform.Id,
		ClientTypeId:     input.LoginData.ClientType.Id,
		UserId:           input.LoginData.UserId,
		RoleId:           input.LoginData.Role.Id,
		Ip:               "0.0.0.0",
		Data:             "additional json data",
		ExpiresAt:        time.Now().Add(config.RefreshTokenExpiresInTime).Format(config.DatabaseTimeLayout),
	})
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, sessionPKey)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	if input.Tables == nil {
		input.Tables = []*pb.Object{}
	}

	// // TODO - wrap in a function
	m := map[string]interface{}{
		"id":                 session.Id,
		"project_id":         session.ProjectId,
		"client_platform_id": session.ClientPlatformId,
		"client_type_id":     session.ClientTypeId,
		"user_id":            session.UserId,
		"role_id":            session.RoleId,
		"ip":                 session.Data,
		"data":               session.Data,
		"tables":             input.Tables,
	}

	accessToken, err := security.GenerateJWT(m, config.AccessTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken, err := security.GenerateJWT(m, config.RefreshTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	input.LoginData.Token = &pb.Token{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		ExpiresAt:        session.ExpiresAt,
		RefreshInSeconds: int32(config.AccessTokenExpiresInTime.Seconds()),
	}

	return input.LoginData, nil
}

func (s *sessionService) SessionAndTokenGeneratorSuperAdmin(ctx context.Context, input *pb.SessionAndTokenRequest) (*pb.V2LoginResponse, error) {

	// // TODO - Delete all old sessions & refresh token has this function too
	rowsAffected, err := s.strg.Session().DeleteExpiredUserSessions(ctx, input.LoginData.UserId)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.log.Info("Login--->DeleteExpiredUserSessions", logger.Any("rowsAffected", rowsAffected))
	userSessionList, err := s.strg.Session().GetSessionListByUserID(ctx, input.LoginData.UserId)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	input.LoginData.Sessions = userSessionList.Sessions

	_, err = uuid.Parse(input.ProjectId)
	if err != nil {
		input.ProjectId = "f5955c82-f264-4655-aeb4-86fd1c642cb6"
	}

	sessionPKey, err := s.strg.Session().CreateSuperAdmin(ctx, &pb.CreateSessionRequest{
		UserId:    input.LoginData.UserId,
		Ip:        "0.0.0.0",
		Data:      "additional json data",
		ExpiresAt: time.Now().Add(config.RefreshTokenExpiresInTime).Format(config.DatabaseTimeLayout),
	})
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, sessionPKey)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	if input.Tables == nil {
		input.Tables = []*pb.Object{}
	}

	// // TODO - wrap in a function
	m := map[string]interface{}{
		"id":                 session.Id,
		"project_id":         session.ProjectId,
		"client_platform_id": session.ClientPlatformId,
		"client_type_id":     session.ClientTypeId,
		"user_id":            session.UserId,
		"role_id":            session.RoleId,
		"ip":                 session.Data,
		"data":               session.Data,
		"tables":             input.Tables,
	}

	accessToken, err := security.GenerateJWT(m, config.AccessTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken, err := security.GenerateJWT(m, config.RefreshTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	input.LoginData.Token = &pb.Token{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		ExpiresAt:        session.ExpiresAt,
		RefreshInSeconds: int32(config.AccessTokenExpiresInTime.Seconds()),
	}

	return input.LoginData, nil
}

func (s *sessionService) UpdateSessionsByRoleId(ctx context.Context, input *pb.UpdateSessionByRoleIdRequest) (*emptypb.Empty, error) {

	rowsAffected, err := s.strg.Session().UpdateByRoleId(ctx, input)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.log.Info("UpdateByRoleId--->UpdateByRoleId", logger.Any("rowsAffected", rowsAffected))

	return &emptypb.Empty{}, nil
}

func (s *sessionService) MultiCompanyLogin(ctx context.Context, req *pb.MultiCompanyLoginRequest) (*pb.MultiCompanyLoginResponse, error) {
	now := time.Now()
	fmt.Println("TIME1", time.Since(now))
	resp := &pb.MultiCompanyLoginResponse{}

	if len(req.Username) < 6 {
		err := errors.New("invalid username")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if len(req.Password) < 6 {
		err := errors.New("invalid password")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	//userReq, err := helper.ConvertMapToStruct(map[string]interface{}{
	//	"password": req.Password,
	//	"login":    req.Username,
	//})
	//if err != nil {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	//userResp, err := s.services.ObjectBuilderService().GetList(
	//	ctx,
	//	&pbObject.CommonMessage{
	//		TableSlug: "user",
	//		Data:      userReq,
	//		ProjectId: config.UcodeDefaultProjectID,
	//	},
	//)
	//if err != nil {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	user, err := s.strg.User().GetByUsername(ctx, req.GetUsername())
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	match, err := security.ComparePassword(user.Password, req.Password)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !match {
		err := errors.New("username or password is wrong")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	fmt.Println("TIME2", time.Since(now))

	//userDatas, ok := userResp.Data.AsMap()["response"].([]interface{})
	//if !ok {
	//	err := errors.New("invalid assertion")
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	//if len(userDatas) < 1 {
	//	err := errors.New("user not found")
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.PermissionDenied, err.Error())
	//} else if len(userDatas) > 1 {
	//	err := errors.New("many users found")
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.PermissionDenied, err.Error())
	//}

	//userData, ok := userDatas[0].(map[string]interface{})
	//if !ok {
	//	err := errors.New("invalid assertion")
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//clientTypeId, ok := userData["client_type_id"].(string)
	//if !ok {
	//	err := errors.New("invalid assertion")
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	fmt.Println("TIME3", time.Since(now))

	//clientTypeReq, err := helper.ConvertMapToStruct(map[string]interface{}{
	//	"id": clientTypeId,
	//})
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	//fmt.Println("Client Type Request ================>: ", clientTypeReq)
	//clientTypeResp, err := s.services.ObjectBuilderService().GetSingle(
	//	ctx,
	//	&pbObject.CommonMessage{
	//		TableSlug: "client_type",
	//		Data:      clientTypeReq,
	//		ProjectId: config.UcodeDefaultProjectID,
	//	},
	//)
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("TIME4", time.Since(now))

	//clientTypeData, ok := clientTypeResp.Data.AsMap()["response"].(map[string]interface{})
	//if !ok {
	//	err := errors.New("invalid assertion")
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	//err = errors.New("invalid assertion")
	//id, ok := clientTypeData["guid"].(string)
	//if !ok {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//name, ok := clientTypeData["name"].(string)
	//if !ok {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//selfRegister, ok := clientTypeData["self_register"].(bool)
	//if !ok {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//selfRecover, ok := clientTypeData["self_recover"].(bool)
	//if !ok {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//projectId, ok := clientTypeData["project_id"].(string)
	//if !ok {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//fmt.Println("TIME5", time.Since(now))
	//
	//// confirmBy, ok := clientTypeData["confirm_by"].(string)
	//// if !ok {
	//// 	err := errors.New("invalid assertion")
	//// 	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//// 	return nil, status.Error(codes.Internal, err.Error())
	//// }
	//
	//// pb.ConfirmStrategies(pb.ConfirmStrategies_value[confirmBy])
	//
	//clientType := &pb.ClientType{
	//	Id:           id,
	//	Name:         name,
	//	SelfRegister: selfRegister,
	//	SelfRecover:  selfRecover,
	//	ProjectId:    projectId,
	//	// ConfirmBy:    confirmBy,
	//}
	//
	//resp.ClientTypes = append(resp.ClientTypes, clientType)
	//
	//userId, ok := userData["guid"].(string)
	//if !ok {
	//	err := errors.New("invalid assertion")
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//userCompanyProjects, err := s.services.CompanyServiceClient().GetListWithProjects(ctx,
	//	&company_service.GetListWithProjectsRequest{
	//		OwnerId: userId,
	//	})
	//
	//if err != nil {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//fmt.Println("TIME6", time.Since(now))
	//
	//bytes, err := json.Marshal(userCompanyProjects.GetCompanies())
	//if err != nil {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//err = json.Unmarshal(bytes, &resp.Companies)
	//if err != nil {
	//	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}
	//
	//fmt.Println("TIME7", time.Since(now))

	return resp, nil
}
