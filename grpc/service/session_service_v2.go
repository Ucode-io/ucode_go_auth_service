package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	secure "ucode/ucode_go_auth_service/pkg/security"

	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/saidamir98/udevs_pkg/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *sessionService) V2Login(ctx context.Context, req *pb.V2LoginRequest) (*pb.V2LoginResponse, error) {

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

	var projectID string
	projectID = req.ProjectId

	project, err := s.strg.Project().GetByPK(ctx, &pb.ProjectPrimaryKey{Id: req.ProjectId})
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	company, err := s.strg.Company().GetByID(ctx, &pb.CompanyPrimaryKey{Id: project.CompanyId})
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	user, err := s.strg.User().GetByUsername(ctx, req.Username)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err == nil {
		if company.OwnerId == user.Id {
			projectID = project.Id
		}
	}

	data, err := s.services.LoginService().Login(ctx, &object_builder_service.LoginRequest{
		Password:      req.Password,
		Login:         req.Username,
		ClientType:    req.ClientType,
		LoginStrategy: req.LoginStrategy,
	})
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	fmt.Println("appPermission ::::", data.AppPermissions)

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
	methodField := ""
	tableSlug := ""
	slugs := make(map[string]int)
	request := make(map[string]interface{})
	res := make(map[string]interface{})
	resp := &pbObject.CommonMessage{}

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

	tableSlug = splitedPath[len(splitedPath)-1]
	if tableSlug[len(tableSlug)-2:] == "id" {
		tableSlug = splitedPath[len(splitedPath)-2]
	}

	// these apis also manage by app's permission

	slugs["field"] = 1
	slugs["view"] = 1
	slugs["table"] = 1
	slugs["relation"] = 1
	slugs["section"] = 1
	slugs["view"] = 1
	slugs["view_relation"] = 1
	slugs["html-template"] = 1
	slugs["variable"] = 1
	slugs["dashboard"] = 1
	slugs["panel"] = 1
	slugs["html-to-pdf"] = 1
	slugs["document"] = 1
	slugs["template-to-html"] = 1
	slugs["many-to-many"] = 1
	slugs["upload"] = 1
	slugs["upload-file"] = 1
	slugs["close-cashbox"] = 1
	slugs["open-cashbox"] = 1
	slugs["cashbox_transaction"] = 1
	slugs["query"] = 1
	slugs["event"] = 1
	slugs["event-log"] = 1
	slugs["permission-upsert"] = 1
	slugs["custom-event"] = 1
	slugs["excel"] = 1
	slugs["field-permission"] = 1
	slugs["function"] = 1
	slugs["invoke_function"] = 1

	if _, ok := slugs[tableSlug]; ok {
		tableSlug = "app"
	}

	request["client_type_id"] = session.ClientTypeId
	request[methodField] = "Yes"
	request["table_slug"] = tableSlug
	structPb, err := helper.ConvertMapToStruct(request)
	if err != nil {
		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

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
	clientName := convertedClientType["response"].(map[string]interface{})["name"]

	if clientName == nil {
		err := errors.New("Wrong client type")
		s.log.Error("!!!V2HasAccess--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if session.ClientTypeId != "142e9d0b-d9d3-4f71-bde1-5f1dbd70e83d" || clientName.(string) != "ADMIN" {
		resp, err = s.services.ObjectBuilderService().GetList(ctx, &pbObject.CommonMessage{
			TableSlug: "record_permission",
			Data:      structPb,
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
		"tables:":            authTables,
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

func (s *sessionService) UpdateSessionsByRoleId(ctx context.Context, input *pb.UpdateSessionByRoleIdRequest) (*emptypb.Empty, error) {

	rowsAffected, err := s.strg.Session().UpdateByRoleId(ctx, input)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.log.Info("UpdateByRoleId--->UpdateByRoleId", logger.Any("rowsAffected", rowsAffected))

	return &emptypb.Empty{}, nil
}
