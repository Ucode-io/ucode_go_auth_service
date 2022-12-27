package service

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (s *sessionService) V2Login(ctx context.Context, req *pb.V2LoginRequest) (*pb.V2LoginResponse, error) {
	fmt.Println("TEST::::1")
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
		s.log.Error("!!!V2Login--->", logger.Error(err))
		if err == sql.ErrNoRows {
			errNoRows := errors.New("no user found")
			return nil, status.Error(codes.Internal, errNoRows.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	fmt.Println("TEST::::2")
	match, err := security.ComparePassword(user.Password, req.Password)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	fmt.Println("TEST::::3")
	if !match {
		err := errors.New("username or password is wrong")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fmt.Println("TEST::::4")
	expiresAt, err := time.Parse(config.DatabaseTimeLayout, user.GetExpiresAt())
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	fmt.Println("TEST::::5")
	if expiresAt.Unix() < time.Now().Unix() {
		err := errors.New("User has been expired")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fmt.Println("TEST::::6")
	data, err := s.services.LoginService().LoginData(
		ctx,
		&pbObject.LoginDataReq{
			UserId:     user.GetId(),
			ClientType: req.ClientType,
			ProjectId:  config.UcodeDefaultProjectID,
		},
	)
	if err != nil {
		errGetUserProjectData := errors.New("invalid user project data")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
	}
	fmt.Println("TEST::::7")
	if !data.UserFound {
		customError := errors.New("User not found")
		s.log.Error("!!!Login--->", logger.Error(customError))
		return nil, status.Error(codes.NotFound, customError.Error())
	}

	res := helper.ConvertPbToAnotherPb(&pbObject.V2LoginResponse{
		ClientPlatform: data.GetClientPlatform(),
		ClientType:     data.GetClientType(),
		UserFound:      data.GetUserFound(),
		UserId:         data.GetUserId(),
		Role:           data.GetRole(),
		Permissions:    data.GetPermissions(),
		LoginTableSlug: data.GetLoginTableSlug(),
		AppPermissions: data.GetAppPermissions(),
	})
	fmt.Println("TEST::::8")
	resp, err := s.SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
		LoginData: res,
		Tables:    req.Tables,
		ProjectId: req.GetProjectId(),
	})
	if resp == nil {
		errGenerateToken := errors.New("unable to generate token")
		s.log.Error("!!!Login--->", logger.Error(errGenerateToken))
		return nil, status.Error(codes.NotFound, errGenerateToken.Error())
	}
	if err != nil {
		errGenerateToken := errors.New("unable to generate token")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, errGenerateToken.Error())
	}
	fmt.Println("TEST::::9")
	return res, nil
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

	userReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"password": req.Password,
		"login":    req.Username,
	})
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	userResp, err := s.services.ObjectBuilderService().GetList(
		ctx,
		&pbObject.CommonMessage{
			TableSlug: "user",
			Data:      userReq,
			ProjectId: config.UcodeDefaultProjectID,
		},
	)
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

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

	userDatas, ok := userResp.Data.AsMap()["response"].([]interface{})
	if !ok {
		err := errors.New("invalid assertion")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(userDatas) < 1 {
		err := errors.New("user not found")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.PermissionDenied, err.Error())
	} else if len(userDatas) > 1 {
		err := errors.New("many users found")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}

	userData, ok := userDatas[0].(map[string]interface{})
	if !ok {
		err := errors.New("invalid assertion")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	clientTypeId, ok := userData["client_type_id"].(string)
	if !ok {
		err := errors.New("invalid assertion")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("TIME3", time.Since(now))

	clientTypeReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"id": clientTypeId,
	})
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("Client Type Request ================>: ", clientTypeReq)
	clientTypeResp, err := s.services.ObjectBuilderService().GetSingle(
		ctx,
		&pbObject.CommonMessage{
			TableSlug: "client_type",
			Data:      clientTypeReq,
			ProjectId: config.UcodeDefaultProjectID,
		},
	)
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("TIME4", time.Since(now))

	clientTypeData, ok := clientTypeResp.Data.AsMap()["response"].(map[string]interface{})
	if !ok {
		err := errors.New("invalid assertion")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = errors.New("invalid assertion")
	id, ok := clientTypeData["guid"].(string)
	if !ok {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	name, ok := clientTypeData["name"].(string)
	if !ok {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	selfRegister, ok := clientTypeData["self_register"].(bool)
	if !ok {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	selfRecover, ok := clientTypeData["self_recover"].(bool)
	if !ok {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	projectId, ok := clientTypeData["project_id"].(string)
	if !ok {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("TIME5", time.Since(now))

	// confirmBy, ok := clientTypeData["confirm_by"].(string)
	// if !ok {
	// 	err := errors.New("invalid assertion")
	// 	s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// pb.ConfirmStrategies(pb.ConfirmStrategies_value[confirmBy])

	clientType := &pb.ClientType{
		Id:           id,
		Name:         name,
		SelfRegister: selfRegister,
		SelfRecover:  selfRecover,
		ProjectId:    projectId,
		// ConfirmBy:    confirmBy,
	}

	resp.ClientTypes = append(resp.ClientTypes, clientType)

	userId, ok := userData["guid"].(string)
	if !ok {
		err := errors.New("invalid assertion")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	userCompanyProjects, err := s.services.CompanyServiceClient().GetListWithProjects(ctx,
		&company_service.GetListWithProjectsRequest{
			OwnerId: userId,
		})

	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("TIME6", time.Since(now))

	bytes, err := json.Marshal(userCompanyProjects.GetCompanies())
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	err = json.Unmarshal(bytes, &resp.Companies)
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("TIME7", time.Since(now))

	return resp, nil
}

func (s *sessionService) V2MultiCompanyLogin(ctx context.Context, req *pb.V2MultiCompanyLoginReq) (*pb.V2MultiCompanyLoginRes, error) {
	resp := pb.V2MultiCompanyLoginRes{
		Companies: []*pb.V2MultiCompanyLoginRes_Company{},
	}

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

	user, err := s.strg.User().GetByUsername(ctx, req.GetUsername())
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	match, err := security.ComparePassword(user.Password, req.Password)
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if !match {
		err := errors.New("username or password is wrong")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userProjects, err := s.strg.User().GetUserProjects(ctx, user.GetId())
	if err != nil {
		errGetProjects := errors.New("cant get user projects")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, errGetProjects.Error())
	}

	fmt.Println("userProjects", userProjects)

	for _, item := range userProjects.Companies {
		projects := make([]*pb.V2MultiCompanyLoginRes_Company_Project, 0, 20)
		company, err := s.services.CompanyServiceClient().GetById(ctx,
			&company_service.GetCompanyByIdRequest{
				Id: item.Id,
			})

		if err != nil {
			errGetProjects := errors.New("cant get user projects")
			s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, errGetProjects.Error())
		}

		for _, projectId := range item.Projects {
			fmt.Println("hello")
			projectInfo, err := s.services.ProjectServiceClient().GetById(
				ctx,
				&company_service.GetProjectByIdRequest{
					ProjectId: projectId,
					CompanyId: item.Id,
				})

			if err != nil {
				errGetProjects := errors.New("cant get user projects")
				s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
				return nil, status.Error(codes.NotFound, errGetProjects.Error())
			}

			projects = append(projects, &pb.V2MultiCompanyLoginRes_Company_Project{
				Id:        projectInfo.GetProjectId(),
				CompanyId: projectInfo.GetCompanyId(),
				Name:      projectInfo.GetTitle(),
				Domain:    projectInfo.GetK8SNamespace(),
			})
		}

		resp.Companies = append(resp.Companies, &pb.V2MultiCompanyLoginRes_Company{
			Id:          company.GetCompany().GetId(),
			Name:        company.GetCompany().GetName(),
			Logo:        company.GetCompany().GetLogo(),
			Description: company.GetCompany().GetLogo(),
			OwnerId:     company.GetCompany().GetOwnerId(),
			Projects:    projects,
		})
	}

	return &resp, nil
}

func (s *sessionService) V2HasAccessUser(ctx context.Context, req *pb.V2HasAccessUserReq) (*pb.V2HasAccessUserRes, error) {

	tokenInfo, err := secure.ParseClaims(req.AccessToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!V2HasAccessUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!V2HasAccessUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if session.IsChanged {
		err := errors.New("permission update")
		s.log.Error("!!!V2HasAccessUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	expiresAt, err := time.Parse(config.DatabaseTimeLayout, session.ExpiresAt)
	if err != nil {
		s.log.Error("!!!V2HasAccessUser--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if expiresAt.Unix() < time.Now().Unix() {
		err := errors.New("user has been expired")
		s.log.Error("!!!V2HasAccessUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	//_, err = s.strg.Scope().Upsert(ctx, &pb.UpsertScopeRequest{
	//	ClientPlatformId: session.ClientPlatformId,
	//	Path:             req.Path,
	//	Method:           req.Method,
	//})
	//if err != nil {
	//	s.log.Error("!!!V2HasAccess--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	//var methodField string
	//switch req.Method {
	//case "GET":
	//	methodField = "read"
	//case "POST":
	//	methodField = "write"
	//case "PUT":
	//	methodField = "update"
	//case "DELETE":
	//	methodField = "delete"
	//}
	//
	//splitedPath := strings.Split(req.Path, "/")
	//splitedPath = splitedPath[1:]
	//
	//var tableSlug string
	//tableSlug = splitedPath[len(splitedPath)-1]
	//if tableSlug[len(tableSlug)-2:] == "id" {
	//	tableSlug = splitedPath[len(splitedPath)-2]
	//}
	//
	//if _, ok := config.ObjectBuilderTableSlugs[tableSlug]; ok {
	//	tableSlug = "app"
	//}
	//
	//request := make(map[string]interface{})
	//request["client_type_id"] = session.ClientTypeId
	//request[methodField] = "Yes"
	//request["table_slug"] = tableSlug

	//clientType, err := s.services.ClientService().V2GetClientTypeByID(ctx, &pb.ClientTypePrimaryKey{
	//	Id: session.ClientTypeId,
	//})
	//if err != nil {
	//	s.log.Error("!!!V2HasAccess.ClientService.V2GetClientTypeByID--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	//convertedClientType, err := helper.ConvertStructToResponse(clientType.Data)
	//if err != nil {
	//	s.log.Error("!!!V2HasAccess.ConvertStructToResponse--->", logger.Error(err))
	//	return nil, status.Error(codes.Internal, err.Error())
	//}

	//clientName, ok := convertedClientType["response"].(map[string]interface{})["name"]
	//if !ok {
	//	res := make(map[string]interface{})
	//	resp := &pbObject.CommonMessage{}
	//
	//	if clientName == nil {
	//		err := errors.New("Wrong client type")
	//		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
	//		return nil, status.Error(codes.Internal, err.Error())
	//	}
	//
	//	structPb, err := helper.ConvertMapToStruct(request)
	//	if err != nil {
	//		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
	//		return nil, status.Error(codes.Internal, err.Error())
	//	}
	//
	//	if session.ClientTypeId != config.AdminClientPlatformID || clientName.(string) != config.AdminClientName {
	//		resp, err = s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
	//			TableSlug: "record_permission",
	//			Data:      structPb,
	//			ProjectId: config.UcodeDefaultProjectID,
	//		})
	//		if err != nil {
	//			s.log.Error("!!!V2HasAccess.ObjectBuilderService.GetList--->", logger.Error(err))
	//			return nil, status.Error(codes.Internal, err.Error())
	//		}
	//
	//		res, err = helper.ConvertStructToResponse(resp.Data)
	//		if err != nil {
	//			s.log.Error("!!!V2HasAccess.ConvertStructToResponse--->", logger.Error(err))
	//			return nil, status.Error(codes.Internal, err.Error())
	//		}
	//
	//		if len(res["response"].([]interface{})) == 0 {
	//			err := errors.New("Permission denied")
	//			s.log.Error("!!!V2HasAccess--->", logger.Error(err))
	//			return nil, status.Error(codes.PermissionDenied, err.Error())
	//		}
	//	}
	//}

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

	projects, err := s.services.UserService().GetProjectsByUserId(ctx, &pb.GetProjectsByUserIdReq{
		UserId: session.GetUserId(),
	})
	if err != nil {
		s.log.Error("V2HasAccessUser", logger.Error(err))
		return nil, err
	}

	exist := false
	for _, item := range projects.GetProjectIds() {
		if item == req.GetProjectId() {
			exist = true
			break
		}
	}

	if !exist {
		err = errors.New("access denied")
		s.log.Error("!!!V2HasAccessUser--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var authTables []*pb.TableBody
	for _, table := range tokenInfo.Tables {
		authTable := &pb.TableBody{
			TableSlug: table.TableSlug,
			ObjectId:  table.ObjectID,
		}
		authTables = append(authTables, authTable)
	}

	return &pb.V2HasAccessUserRes{
		Id:        session.Id,
		ProjectId: session.ProjectId,
		UserId:    session.UserId,
		Ip:        session.Ip,
		Data:      session.Data,
		ExpiresAt: session.ExpiresAt,
		CreatedAt: session.CreatedAt,
		UpdatedAt: session.UpdatedAt,
		Tables:    authTables,
	}, nil
}
