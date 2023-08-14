package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/genproto/sms_service"
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

	// expiresAt, err := time.Parse(config.DatabaseTimeLayout, time.Now().Add(time.Hour).String())
	// if err != nil {
	// 	s.log.Error("!!!Login--->", logger.Error(err))
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }
	//
	// if expiresAt.Unix() < time.Now().Unix() {
	// 	err := errors.New("User has been expired")
	// 	s.log.Error("!!!Login--->", logger.Error(err))
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }

	reqLoginData := &pbObject.LoginDataReq{
		UserId:                user.GetId(),
		ClientType:            req.GetClientType(),
		ProjectId:             req.GetProjectId(),
		ResourceEnvironmentId: req.GetResourceEnvironmentId(),
	}
	log.Println("reqLoginData--->", reqLoginData)
	var data *pbObject.LoginDataRes

	switch req.ResourceType {
	case 1:
		data, err = s.services.LoginService().LoginData(
			ctx,
			reqLoginData,
		)

		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			s.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}
	case 3:
		data, err = s.services.PostgresLoginService().LoginData(
			ctx,
			reqLoginData,
		)

		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			s.log.Error("!!!PostgresBuilder.Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}

	}

	if !data.UserFound {
		customError := errors.New("User not found")
		s.log.Error("!!!Login--->", logger.Error(customError))
		return nil, status.Error(codes.NotFound, customError.Error())
	}

	res := helper.ConvertPbToAnotherPb(&pbObject.V2LoginResponse{
		ClientPlatform:   data.GetClientPlatform(),
		ClientType:       data.GetClientType(),
		UserFound:        data.GetUserFound(),
		UserId:           data.GetUserId(),
		Role:             data.GetRole(),
		Permissions:      data.GetPermissions(),
		LoginTableSlug:   data.GetLoginTableSlug(),
		AppPermissions:   data.GetAppPermissions(),
		GlobalPermission: data.GetGlobalPermission(),
	})
	//if bytes, err := json.MarshalIndent(res, "", "  "); err == nil {
	//
	//}

	resp, err := s.SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
		LoginData:     res,
		Tables:        req.Tables,
		ProjectId:     req.GetProjectId(),
		EnvironmentId: req.GetEnvironmentId(),
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

	if req.Tables != nil {
		res.Tables = req.Tables
	}

	return res, nil
}

func (s *sessionService) V2LoginWithOption(ctx context.Context, req *pb.V2LoginWithOptionRequest) (*pb.V2LoginWithOptionsResponse, error) {
	s.log.Info("V2LoginWithOption --> ", logger.Any("request: ", req))
	var (
		userId   string
		verified bool
	)
pwd:
	switch strings.ToUpper(req.GetLoginStrategy()) {
	case "LOGIN_PWD":
		username, ok := req.GetData()["username"]
		if ok {
			if len(username) < 6 {
				err := errors.New("invalid username")
				s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		} else {
			err := errors.New("username is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		password, ok := req.GetData()["password"]
		if ok {
			if len(password) < 6 {
				err := errors.New("invalid password")
				s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		} else {
			err := errors.New("password is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		user, err := s.strg.User().GetByUsername(ctx, username)
		if err != nil {
			s.log.Error("!!!V2V2LoginWithOption--->", logger.Error(err))
			if err == sql.ErrNoRows {
				errNoRows := errors.New("no user found")
				return nil, status.Error(codes.Internal, errNoRows.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}
		match, err := security.ComparePassword(user.Password, password)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !match {
			err := errors.New("username or password is wrong")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = user.Id
	case "PHONE":
		phone, ok := req.GetData()["phone"]

		if !ok {
			err := errors.New("phone is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		user, err := s.strg.User().GetByUsername(ctx, phone)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = user.GetId()

	case "EMAIL":
		email, ok := req.GetData()["email"]
		if !ok {
			err := errors.New("email is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		user, err := s.strg.User().GetByUsername(ctx, email)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = user.GetId()
	case "LOGIN":
		username, ok := req.GetData()["username"]
		if !ok {
			err := errors.New("username is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		user, err := s.strg.User().GetByUsername(ctx, username)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = user.GetId()
	case "PHONE_OTP":
		sms_id, ok := req.GetData()["sms_id"]
		if !ok {
			err := errors.New("sms_id is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		otp, ok := req.GetData()["otp"]
		if !ok {
			err := errors.New("otp is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		phone, ok := req.GetData()["phone"]
		if !ok {
			err := errors.New("phone is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		smsOtpSettings, err := s.services.SmsOtpSettingsService().GetList(ctx, &pb.GetListSmsOtpSettingsRequest{
			EnvironmentId: req.Data["environment_id"],
			ProjectId:     req.Data["project_id"],
		})
		if err != nil {
			s.log.Error("!!!V2LoginWithOption.SmsOtpSettingsService().GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		var defaultOtp string
		if len(smsOtpSettings.GetItems()) > 0 {
			if smsOtpSettings.GetItems()[0].GetDefaultOtp() != "" {
				defaultOtp = smsOtpSettings.GetItems()[0].GetDefaultOtp()
			}
		}
		if defaultOtp != otp {
			_, err = s.services.SmsService().ConfirmOtp(
				ctx,
				&sms_service.ConfirmOtpRequest{
					SmsId: sms_id,
					Otp:   otp,
				},
			)
			if err != nil {
				return nil, err
			}
		}
		verified = true
		user, err := s.strg.User().GetByUsername(ctx, phone)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		userId = user.GetId()
	case "EMAIL_OTP":
		sms_id, ok := req.GetData()["sms_id"]
		if !ok {
			err := errors.New("sms_id is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		otp, ok := req.GetData()["otp"]
		if !ok {
			err := errors.New("otp is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		email, ok := req.GetData()["email"]
		if !ok {
			err := errors.New("email is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		_, err := s.services.SmsService().ConfirmOtp(
			ctx,
			&sms_service.ConfirmOtpRequest{
				SmsId: sms_id,
				Otp:   otp,
			},
		)
		if err != nil {
			return nil, err
		}
		verified = true
		user, err := s.strg.User().GetByUsername(ctx, email)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = user.GetId()
	case "PHONE_PWD":
		phone, ok := req.GetData()["phone"]
		if !ok {
			err := errors.New("phone is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		password, ok := req.GetData()["password"]
		if ok {
			if len(password) < 6 {
				err := errors.New("invalid password")
				s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		} else {
			err := errors.New("password is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userIdRes, err := s.strg.User().GetByUsername(ctx, phone)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		user, err := s.strg.User().GetByPK(ctx, &auth_service.UserPrimaryKey{
			Id: userIdRes.GetId(),
		})
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		match, err := security.ComparePassword(user.Password, password)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !match {
			err := errors.New("username or password is wrong")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = user.GetId()
	case "EMAIL_PWD":
		email, ok := req.GetData()["email"]
		if !ok {
			err := errors.New("email is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		password, ok := req.GetData()["password"]
		if ok {
			if len(password) < 6 {
				err := errors.New("invalid password")
				s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		} else {
			err := errors.New("password is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userIdRes, err := s.strg.User().GetByUsername(ctx, email)

		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		user, err := s.strg.User().GetByPK(ctx, &auth_service.UserPrimaryKey{
			Id: userIdRes.GetId(),
		})
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		match, err := security.ComparePassword(user.Password, password)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		if !match {
			err := errors.New("username or password is wrong")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = userIdRes.GetId()
	case "GOOGLE_AUTH":
		email, ok := req.GetData()["email"]
		if !ok {
			err := errors.New("email is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		gooleToken, ok := req.GetData()["google_token"]
		if ok {
			userInfo, err := helper.GetGoogleUserInfo(gooleToken)
			if err != nil {
				err = errors.New("Invalid arguments google auth")
				s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			if userInfo["error"] != nil || !(userInfo["email_verified"].(bool)) {
				err = errors.New("Invalid arguments google auth")
				s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		} else {
			err := errors.New("google token is required when login type is google auth")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userIdRes, err := s.strg.User().GetUserByLoginType(ctx, &pb.GetUserByLoginTypesRequest{
			Email: email,
		})
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		userId = userIdRes.GetUserId()
	case "APPLE_AUTH":

		//
		err := errors.New("not implemented")
		return nil, status.Error(codes.InvalidArgument, err.Error())

	default:
		req.LoginStrategy = "LOGIN_PWD"
		goto pwd
	}
	req.Data["user_id"] = userId
	data, err := s.LoginMiddleware(ctx, models.LoginMiddlewareReq{
		Data:   req.Data,
		Tables: req.Tables,
	})
	if err != nil {
		httpErrorStr := ""

		httpErrorStr = strings.Split(err.Error(), "=")[len(strings.Split(err.Error(), "="))-1][1:]
		httpErrorStr = strings.ToLower(httpErrorStr)

		if httpErrorStr == "user not found" && verified {
			err := errors.New("user verified but not found")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.log.Info("Login By " + req.GetLoginStrategy() + " done!")
	return data, nil
}

func (s *sessionService) LoginMiddleware(ctx context.Context, req models.LoginMiddlewareReq) (*pb.V2LoginWithOptionsResponse, error) {
	log.Println("reqLoginData--->", req)

	var res *pb.V2LoginResponse

	if req.Data["project_id"] != "" && req.Data["environment_id"] != "" {
		serviceResource, err := s.services.ServiceResource().GetSingle(ctx, &company_service.GetSingleServiceResourceReq{
			EnvironmentId: req.Data["environment_id"],
			ProjectId:     req.Data["project_id"],
			ServiceType:   company_service.ServiceType_BUILDER_SERVICE,
		})
		if err != nil {
			errGetUserProjectData := errors.New("unable to get resource")
			s.log.Error("!!!LoginMiddleware--->LoginService()", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}

		reqLoginData := &pbObject.LoginDataReq{
			UserId:                req.Data["user_id"],
			ProjectId:             req.Data["project_id"],
			ResourceEnvironmentId: serviceResource.GetResourceEnvironmentId(),
			ClientType:            req.Data["client_type_id"],
		}
		log.Println("reqLoginData--->", reqLoginData)
		var data *pbObject.LoginDataRes

		switch serviceResource.ResourceType {
		case 1:
			data, err = s.services.LoginService().LoginData(
				ctx,
				reqLoginData,
			)

			if err != nil {
				errGetUserProjectData := errors.New("invalid user project data")
				s.log.Error("!!!LoginMiddleware--->LoginService()", logger.Error(err))
				return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
			}
		case 3:
			data, err = s.services.PostgresLoginService().LoginData(
				ctx,
				reqLoginData,
			)

			if err != nil {
				errGetUserProjectData := errors.New("invalid user project data")
				s.log.Error("!!!LoginMiddleware--->PostgresLoginService", logger.Error(err))
				return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
			}

		}

		if !data.UserFound {
			customError := errors.New("User not found")
			s.log.Error("!!!LoginMiddleware--->", logger.Error(customError))
			return nil, status.Error(codes.NotFound, customError.Error())
		}

		res = helper.ConvertPbToAnotherPb(&pbObject.V2LoginResponse{
			ClientPlatform: data.GetClientPlatform(),
			ClientType:     data.GetClientType(),
			UserFound:      data.GetUserFound(),
			UserId:         data.GetUserId(),
			Role:           data.GetRole(),
			Permissions:    data.GetPermissions(),
			LoginTableSlug: data.GetLoginTableSlug(),
		})

	}
	if req.Tables == nil {
		req.Tables = []*pb.Object{}
	}

	resp, err := s.SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
		LoginData:     res,
		ProjectId:     req.Data["project_id"],
		Tables:        req.Tables,
		EnvironmentId: req.Data["environment_id"],
	})

	if resp == nil {
		err := errors.New("error while generating token")
		s.log.Error("!!!LoginMiddleware--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		s.log.Error("!!!LoginMiddleware--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	companies, err := s.services.CompanyServiceClient().GetList(ctx, &company_service.GetCompanyListRequest{
		Offset:  0,
		Limit:   128,
		OwnerId: req.Data["user_id"],
	})
	if err != nil {
		return nil, err
	}

	companiesResp := []*auth_service.Company{}

	if len(companies.Companies) < 1 {
		companiesById := make([]*company_service.Company, 0)
		user, err := s.strg.User().GetByPK(ctx, &auth_service.UserPrimaryKey{
			Id: resp.GetUserId(),
		})
		if err != nil {
			return nil, err
		}
		company, err := s.services.CompanyServiceClient().GetById(ctx, &company_service.GetCompanyByIdRequest{
			Id: user.GetCompanyId(),
		})
		if err != nil {
			return nil, err
		}
		companiesById = append(companiesById, company.Company)
		companies.Companies = companiesById
		companies.Count = int32(len(companiesById))

	}
	bytes, err := json.Marshal(companies.GetCompanies())
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &companiesResp)
	if err != nil {
		return nil, err
	}

	return &pb.V2LoginWithOptionsResponse{
		UserFound:       true,
		UserId:          req.Data["user_id"],
		Token:           resp.GetToken(),
		Sessions:        resp.GetSessions(),
		Companies:       companiesResp,
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

func (s *sessionService) V2LoginSuperAdmin(ctx context.Context, req *pb.V2LoginSuperAdminReq) (*pb.V2LoginSuperAdminRes, error) {
	if len(req.Username) < 6 {
		err := errors.New("invalid username")
		s.log.Error("!!!V2LoginSuperAdmin--->", logger.Error(err))
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

	//

	// @TODO:: get user expires from builder
	// expiresAt, err := time.Parse(config.DatabaseTimeLayout, time.Now().Add(time.Hour).String())
	// if err != nil {
	// 	s.log.Error("!!!Login--->", logger.Error(err))
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	// if expiresAt.Unix() < time.Now().Unix() {
	// 	err := errors.New("User has been expired")
	// 	s.log.Error("!!!Login--->", logger.Error(err))
	// 	return nil, status.Error(codes.InvalidArgument, err.Error())
	// }
	resp, err := s.SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
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
		ProjectId: user.ProjectId,
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
	// this is for object get list api because our object/get-list api is post method.
	if strings.Contains(req.GetPath(), "object/get-list/") || strings.Contains(req.GetPath(), "object-slim/get-list") {
		methodField = "read"
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

	clientType, err := s.services.ClientService().V2GetClientTypeByID(ctx, &pb.V2ClientTypePrimaryKey{
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
				ProjectId: session.ProjectId,
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
		LoginTableSlug:   tokenInfo.LoginTableSlug,
		EnvId:            session.EnvId,
	}, nil
}

func (s *sessionService) V2RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2RefreshTokenResponse, error) {

	tokenInfo, err := secure.ParseClaims(req.RefreshToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	fmt.Println("\n\n tokenInfo >>>>>>>>>>>> #1", tokenInfo.RoleID)
	// _, err = s.strg.Session().UpdateByRoleId(ctx, &pb.UpdateSessionByRoleIdRequest{
	// 	RoleId:    tokenInfo.RoleID,
	// 	IsChanged: false,
	// })
	// if err != nil {
	// 	s.log.Error("!!!RefreshToken.UpdateByRoleId--->", logger.Error(err))
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if req.ClientTypeId != "" {
		session.ClientTypeId = req.ClientTypeId
	}
	if req.ProjectId != "" {
		session.ProjectId = req.ProjectId
	}
	if req.RoleId != "" {
		session.RoleId = req.RoleId
	}
	if req.EnvId != "" {
		session.EnvId = req.EnvId
	}
	fmt.Println("\n\n session >>>>>>>>>>>> #2", session)
	_, err = s.strg.Session().Update(ctx, &pb.UpdateSessionRequest{
		Id:               session.Id,
		ProjectId:        session.ProjectId,
		ClientPlatformId: session.ClientPlatformId,
		ClientTypeId:     session.ClientTypeId,
		UserId:           session.UserId,
		RoleId:           session.RoleId,
		Ip:               session.Ip,
		Data:             session.Data,
		ExpiresAt:        session.ExpiresAt,
		IsChanged:        session.IsChanged,
		EnvId:            session.EnvId,
	})
	if err != nil {
		s.log.Error("!!!V2RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// userData, err := s.services.LoginService().GetUserUpdatedPermission(ctx, &pbObject.GetUserUpdatedPermissionRequest{
	// 	ClientTypeId: session.ClientTypeId,
	// 	UserId:       session.UserId,
	// 	ProjectId:    session.GetProjectId(),
	// })
	// if err != nil {
	// 	s.log.Error("!!!V2HasAccess.SessionService().GetUserUpdatedPermission--->", logger.Error(err))
	// 	return nil, status.Error(codes.Internal, err.Error())
	// }
	// convertedData := helper.ConvertPbToAnotherPb(userData)

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
		"login_table_slug":   tokenInfo.LoginTableSlug,
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
		Token: token,
	}

	return res, nil
}

func (s *sessionService) V2RefreshTokenSuperAdmin(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2RefreshTokenSuperAdminResponse, error) {

	tokenInfo, err := secure.ParseClaims(req.RefreshToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	//_, err = s.strg.Session().UpdateBySessionId(ctx, &pb.UpdateSessionBySessionIdRequest{
	//	Id:        tokenInfo.ID,
	//	IsChanged: false,
	//})
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
	s.log.Info("--->SessionAndTokenGenerator--->", logger.Any("req", input))
	fmt.Print("user id:::", input.GetLoginData().GetUserId())

	if _, err := uuid.Parse(input.GetLoginData().GetUserId()); err != nil {
		err := errors.New("INVALID USER_ID(UUID)" + err.Error())
		s.log.Error("---ERR->GetLoginData().GetUserId-->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO - Delete all old sessions & refresh token has this function too
	rowsAffected, err := s.strg.Session().DeleteExpiredUserSessions(ctx, input.GetLoginData().GetUserId())
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.log.Info("Login--->DeleteExpiredUserSessions", logger.Any("rowsAffected", rowsAffected))
	userSessionList, err := s.strg.Session().GetSessionListByUserID(ctx, input.GetLoginData().GetUserId())
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	input.LoginData.Sessions = userSessionList.GetSessions()

	_, err = uuid.Parse(input.GetProjectId())
	if err != nil {
		input.ProjectId = "f5955c82-f264-4655-aeb4-86fd1c642cb6"
	}

	sessionPKey, err := s.strg.Session().Create(ctx, &pb.CreateSessionRequest{
		ProjectId:        input.GetProjectId(),
		ClientPlatformId: input.GetLoginData().GetClientPlatform().GetId(),
		ClientTypeId:     input.GetLoginData().GetClientType().GetId(),
		UserId:           input.GetLoginData().GetUserId(),
		RoleId:           input.GetLoginData().GetRole().GetId(),
		Ip:               "0.0.0.0",
		Data:             "additional json data",
		ExpiresAt:        time.Now().Add(config.RefreshTokenExpiresInTime).Format(config.DatabaseTimeLayout),
		EnvId:            input.GetEnvironmentId(),
	})
	if err != nil {
		s.log.Error("!!!Create--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, sessionPKey)
	if err != nil {
		s.log.Error("!!!GetByPK--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	if input.Tables == nil {
		input.Tables = []*pb.Object{}
	}

	userData, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
		ProjectId: input.GetProjectId(),
		Id:        input.GetLoginData().GetUserId(),
	})
	if err != nil {
		s.log.Error("!!!Login->GetByPK--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// // TODO - wrap in a function
	m := map[string]interface{}{
		"id":                 session.GetId(),
		"project_id":         session.GetProjectId(),
		"client_platform_id": session.GetClientPlatformId(),
		"client_type_id":     session.GetClientTypeId(),
		"user_id":            session.GetUserId(),
		"role_id":            session.GetRoleId(),
		"ip":                 session.GetData(),
		"data":               session.GetData(),
		"tables":             input.GetTables(),
		"login_table_slug":   input.GetLoginData().GetLoginTableSlug(),
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
		CreatedAt:        session.GetCreatedAt(),
		UpdatedAt:        session.GetUpdatedAt(),
		ExpiresAt:        session.GetExpiresAt(),
		RefreshInSeconds: int32(config.AccessTokenExpiresInTime.Seconds()),
	}
	input.LoginData.User = userData

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

	clientTypeReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"id": clientTypeId,
	})
	if err != nil {
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

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

		for _, projectId := range item.ProjectIds {

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
			UserId:      user.GetId(),
		})
	}

	return &resp, nil
}

func (s *sessionService) V2HasAccessUser(ctx context.Context, req *pb.V2HasAccessUserReq) (*pb.V2HasAccessUserRes, error) {
	fmt.Println("has access user begin::", time.Now())
	s.log.Info("\n!!!V2HasAccessUser--->", logger.Any("req", req))

	arr_path := strings.Split(req.Path, "/")

	tokenInfo, err := secure.ParseClaims(req.AccessToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!V2HasAccessUser->ParseClaims--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!V2HasAccessUser->GetByPK--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	expiresAt, err := time.Parse(config.DatabaseTimeLayout, session.ExpiresAt)
	if err != nil {
		s.log.Error("!!!V2HasAccessUser->TimeParse--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if expiresAt.Unix() < time.Now().Unix() {
		err := errors.New("user has been expired")
		s.log.Error("!!!V2HasAccessUser->CHeckExpiredToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
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
	// this condition need our object/get-list api because this api's method is post we change it to get
	if (strings.Contains(req.GetPath(), "object/get-list")) && req.GetMethod() != "GET" {
		methodField = "read"
	}

	projects, err := s.services.UserService().GetProjectsByUserId(ctx, &pb.GetProjectsByUserIdReq{
		UserId: session.GetUserId(),
	})
	if err != nil {
		s.log.Error("---V2HasAccessUser->GetProjectsByUserId--->", logger.Error(err))
		return nil, err
	}

	exist := false
	for _, item := range projects.GetProjectIds() {
		if item == session.GetProjectId() {

			exist = true
			break
		}
	}
	if !exist {
		err = errors.New("---V2HasAccessUser->Access denied")
		s.log.Error("---V2HasAccessUser--->AccessDenied--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	var checkPermission bool
	// guess role check
	for _, path := range arr_path {
		if path == "object" || path == "object-slim" {
			checkPermission = true
			break
		}
	}
	if session.RoleId != "027944d2-0460-11ee-be56-0242ac120002" && checkPermission {
		var tableSlug string
		if arr_path[len(arr_path)-1] == ":object_id" {
			tableSlug = arr_path[len(arr_path)-2]
		} else {
			tableSlug = arr_path[len(arr_path)-1]
		}

		resource, err := s.services.ServiceResource().GetSingle(
			ctx,
			&company_service.GetSingleServiceResourceReq{
				ProjectId:     session.ProjectId,
				EnvironmentId: session.EnvId,
				ServiceType:   company_service.ServiceType_BUILDER_SERVICE,
			},
		)
		if err != nil {
			return nil, err
		}

		resp, err := s.services.BuilderPermissionService().GetTablePermission(
			context.Background(),
			&pbObject.GetTablePermissionRequest{
				TableSlug:             tableSlug,
				RoleId:                session.RoleId,
				ResourceEnvironmentId: resource.ResourceEnvironmentId,
				Method:                methodField,
			},
		)
		if err != nil {
			return nil, err
		}
		if !resp.IsHavePermission {
			err := status.Error(codes.PermissionDenied, "Permission denied")
			return nil, err //fmt.Errorf("Permission denied")
		}
	}

	var authTables []*pb.TableBody
	for _, table := range tokenInfo.Tables {
		authTable := &pb.TableBody{
			TableSlug: table.TableSlug,
			ObjectId:  table.ObjectID,
		}
		authTables = append(authTables, authTable)
	}
	fmt.Println("has access user end::", time.Now())

	return &pb.V2HasAccessUserRes{
		Id:               session.Id,
		ProjectId:        session.ProjectId,
		UserId:           session.UserId,
		Ip:               session.Ip,
		Data:             session.Data,
		ExpiresAt:        session.ExpiresAt,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		Tables:           authTables,
		ClientPlatformId: session.ClientPlatformId,
		ClientTypeId:     session.ClientTypeId,
		RoleId:           session.RoleId,
		EnvId:            session.EnvId,
	}, nil
}

func (s *sessionService) V2MultiCompanyOneLogin(ctx context.Context, req *pb.V2MultiCompanyLoginReq) (*pb.V2MultiCompanyOneLoginRes, error) {
	resp := pb.V2MultiCompanyOneLoginRes{
		Companies: []*pb.Company2{},
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

	for _, item := range userProjects.Companies {
		projects := make([]*pb.Project2, 0, 20)
		company, err := s.services.CompanyServiceClient().GetById(ctx,
			&company_service.GetCompanyByIdRequest{
				Id: item.Id,
			})

		if err != nil {
			errGetProjects := errors.New("cant get user projects")
			s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, errGetProjects.Error())
		}

		for _, projectId := range item.ProjectIds {

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

			projects = append(projects, &pb.Project2{
				Id:        projectInfo.GetProjectId(),
				CompanyId: projectInfo.GetCompanyId(),
				Name:      projectInfo.GetTitle(),
				Domain:    projectInfo.GetK8SNamespace(),
			})
		}

		for _, p := range projects {

			// get environments
			environments, err := s.services.ResourceService().GetListConfiguredResourceEnvironment(
				ctx,
				&company_service.GetListConfiguredResourceEnvironmentReq{
					ProjectId: p.Id,
				},
			)
			if err != nil {
				errGetProjects := errors.New("cant get environments")
				s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
				return nil, status.Error(codes.NotFound, errGetProjects.Error())
			}
			for _, en := range environments.Data {
				p.ResourceEnvironments = append(p.ResourceEnvironments, &pb.ResourceEnvironmentV2MultiCompany{
					Id:            en.Id,
					Name:          en.Name,
					ProjectId:     en.ProjectId,
					ResourceId:    en.ResourceId,
					EnvironmentId: en.EnvironmentId,
					IsConfigured:  en.IsConfigured,
					ResourceType:  en.ResourceType,
					ServiceType:   en.ServiceType,
					DisplayColor:  en.DisplayColor,
					Description:   en.Description,
				})

				// get client types
				clientTypes, err := s.services.ClientService().V2GetClientTypeList(
					ctx,
					&pb.V2GetClientTypeListRequest{
						ProjectId: en.Id,
					},
				)
				if err != nil {
					errGetProjects := errors.New("cant get environments")
					s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
					return nil, status.Error(codes.NotFound, errGetProjects.Error())
				}
				p.ClientTypes = clientTypes.Data

			}
		}

		resp.Companies = append(resp.Companies, &pb.Company2{
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

func (s *sessionService) V2ResetPassword(ctx context.Context, req *pb.V2ResetPasswordRequest) (*pb.User, error) {
	s.log.Info("V2ResetPassword -> ", logger.Any("req: ", req))
	if req.GetPassword() != "" {
		if len(req.GetPassword()) < 6 {
			err := fmt.Errorf("password must not be less than 6 characters")
			s.log.Error("!!!ResetPassword--->", logger.Error(err))
			return nil, err
		}

		hashedPassword, err := security.HashPassword(req.GetPassword())
		if err != nil {
			s.log.Error("!!!ResetPassword--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		req.Password = hashedPassword
	}
	rowsAffected, err := s.strg.User().V2ResetPassword(ctx, req)
	if err != nil {
		s.log.Error("!!!ResetPassword--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, "no rows were affected")
	}
	s.log.Info("V2ResetPassword <- ", logger.Any("res: ", rowsAffected))
	return s.strg.User().GetByPK(ctx, &auth_service.UserPrimaryKey{Id: req.GetUserId()})
}
