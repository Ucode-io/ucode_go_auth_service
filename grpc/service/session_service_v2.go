package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"
	nb "ucode/ucode_go_auth_service/genproto/new_object_builder_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/genproto/sms_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/security"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/saidamir98/udevs_pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *sessionService) V2Login(ctx context.Context, req *pb.V2LoginRequest) (*pb.V2LoginResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2Login")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2Login", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2Login", memoryUsed))
		}
	}()

	user := &pb.User{}
	var err error

	switch req.Type {
	case config.Default:
		{
			if len(req.Username) < 6 {
				err := errors.New("invalid username")
				s.log.Error("!!!MultiCompanyLogin--->InvalidUsername", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}

			if len(req.Password) < 6 {
				err := errors.New("invalid password")
				s.log.Error("!!!MultiCompanyLogin--->InvalidPassword", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}

			user, err = s.strg.User().GetByUsername(ctx, req.GetUsername())
			if err != nil {
				s.log.Error("!!!MultiCompanyLogin--->GetByUsername", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			match, err := security.ComparePassword(user.Password, req.Password)
			if err != nil {
				s.log.Error("!!!MultiCompanyLogin--->ComparePassword", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			if !match {
				err := errors.New("username or password is wrong")
				s.log.Error("!!!MultiCompanyLogin Default--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	case config.WithPhone:
		{
			if config.DefaultOtp != req.Otp {
				_, err := s.services.SmsService().ConfirmOtp(
					ctx,
					&sms_service.ConfirmOtpRequest{
						SmsId: req.GetSmsId(),
						Otp:   req.GetOtp(),
					},
				)
				if err != nil {
					return nil, err
				}
			}

			user, err = s.strg.User().GetByUsername(ctx, req.GetPhone())
			if err != nil {
				s.log.Error("!!!MultiCompanyLogin Phone--->GetByUsername", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	case config.WithEmail:
		{
			if config.DefaultOtp != req.Otp {
				_, err := s.services.SmsService().ConfirmOtp(
					ctx,
					&sms_service.ConfirmOtpRequest{
						SmsId: req.GetSmsId(),
						Otp:   req.GetOtp(),
					},
				)
				if err != nil {
					return nil, err
				}
			}

			user, err = s.strg.User().GetByUsername(ctx, req.GetEmail())
			if err != nil {
				s.log.Error("!!!MultiCompanyLogin Email--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	reqLoginData := &pbObject.LoginDataReq{
		UserId:                user.GetId(),
		ClientType:            req.GetClientType(),
		ProjectId:             req.GetProjectId(),
		ResourceEnvironmentId: req.GetResourceEnvironmentId(),
	}
	var data *pbObject.LoginDataRes

	services, err := s.serviceNode.GetByNodeType(
		req.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	switch req.ResourceType {
	case 1:
		data, err = services.GetLoginServiceByType(req.NodeType).LoginData(
			ctx,
			reqLoginData,
		)

		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			s.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}
	case 3:

		newReq := nb.LoginDataReq{}

		err = helper.MarshalToStruct(&reqLoginData, &newReq)
		if err != nil {
			s.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(400, err.Error())
		}

		newReq.ProjectId = newReq.ResourceEnvironmentId

		newData, err := services.GoLoginService().LoginData(
			ctx,
			&newReq,
		)

		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			s.log.Error("!!!PostgresBuilder.Login--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}

		err = helper.MarshalToStruct(&newData, &data)
		if err != nil {
			s.log.Error("!!!Login--->", logger.Error(err))
			return nil, status.Error(400, err.Error())
		}
	}

	if !data.UserFound {
		customError := errors.New("user not found")
		s.log.Error("!!!Login--->", logger.Error(customError))
		return nil, status.Error(codes.NotFound, customError.Error())
	}

	userData, err := helper.ConvertStructToResponse(data.UserData)
	if err != nil {
		return nil, status.Error(400, err.Error())
	}

	delete(userData, "password")

	data.UserData, err = helper.ConvertMapToStruct(userData)
	if err != nil {
		return nil, status.Error(400, err.Error())
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
		UserData:         data.GetUserData(),
	})

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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2LoginWithOption")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2LoginWithOption", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2LoginWithOption", memoryUsed))
		}
	}()

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
			s.log.Error("!!!V2LoginWithOption--->NoSMSId", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		otp, ok := req.GetData()["otp"]
		if !ok {
			err := errors.New("otp is empty")
			s.log.Error("!!!V2LoginWithOption--->NoOTP", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		phone, ok := req.GetData()["phone"]
		if !ok {
			err := errors.New("phone is empty")
			s.log.Error("!!!V2LoginWithOption--->NoPhone", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		smsOtpSettings, err := s.services.ResourceService().GetProjectResourceList(
			context.Background(),
			&pbCompany.GetProjectResourceListRequest{
				EnvironmentId: req.Data["environment_id"],
				ProjectId:     req.Data["project_id"],
				Type:          pbCompany.ResourceType_SMS,
			})
		if err != nil {
			s.log.Error("!!!V2LoginWithOption.SmsOtpSettingsService().GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
		var defaultOtp string
		if len(smsOtpSettings.GetResources()) > 0 {
			if smsOtpSettings.GetResources()[0].GetSettings().GetSms().GetDefaultOtp() != "" {
				defaultOtp = smsOtpSettings.GetResources()[0].GetSettings().GetSms().GetDefaultOtp()
			}
		}

		if defaultOtp != otp {
			_, err = s.services.SmsService().ConfirmOtp(
				ctx,
				&sms_service.ConfirmOtpRequest{SmsId: sms_id, Otp: otp},
			)
			if err != nil {
				s.log.Error("!!!V2LoginWithOption--->ConfirmOTP", logger.Error(err))
				return nil, err
			}
		}
		verified = true
		user, err := s.strg.User().GetByUsername(ctx, phone)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption--->GetUserByUsername", logger.Error(err))
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

		if otp != "4554" {
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
		user, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
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
		user, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
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
				err = errors.New("invalid arguments google auth")
				s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
			if userInfo["error"] != nil || !(userInfo["email_verified"].(bool)) {
				err = errors.New("invalid arguments google auth")
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
			s.log.Error("!!!V2LoginWithOption--->UserNotFound", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("!!!V2LoginWithOption--->LoginMiddleware", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return data, nil
}

func (s *sessionService) LoginMiddleware(ctx context.Context, req models.LoginMiddlewareReq) (*pb.V2LoginWithOptionsResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.LoginMiddleware")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the LoginMiddleware", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("LoginMiddleware", memoryUsed))
		}
	}()

	var res *pb.V2LoginResponse

	if req.Data["project_id"] != "" && req.Data["environment_id"] != "" {
		serviceResource, err := s.services.ServiceResource().GetSingle(ctx, &pbCompany.GetSingleServiceResourceReq{
			EnvironmentId: req.Data["environment_id"],
			ProjectId:     req.Data["project_id"],
			ServiceType:   pbCompany.ServiceType_BUILDER_SERVICE,
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
			NodeType:              serviceResource.GetNodeType(),
		}
		var data *pbObject.LoginDataRes

		services, err := s.serviceNode.GetByNodeType(
			req.Data["project_id"],
			req.NodeType,
		)
		if err != nil {
			s.log.Error("!!!LoginMiddleware--->GetByNodeType", logger.Error(err))
			return nil, err
		}

		switch serviceResource.ResourceType {
		case 1:
			data, err = services.GetLoginServiceByType(req.NodeType).LoginData(
				ctx,
				reqLoginData,
			)

			if err != nil {
				errGetUserProjectData := errors.New("invalid user project data")
				s.log.Error("!!!LoginMiddleware--->LoginService()", logger.Error(err))
				return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
			}
		case 3:
			goReq := &nb.LoginDataReq{}

			err = helper.MarshalToStruct(reqLoginData, &goReq)
			if err != nil {
				s.log.Error("!!!LoginMiddleware--->PostgresMarshal2Struct", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			goData, err := services.GoLoginService().LoginData(ctx, goReq)
			if err != nil {
				errGetUserProjectData := errors.New("invalid user project data")
				s.log.Error("!!!LoginMiddleware--->PostgresLoginService", logger.Error(err))
				return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
			}

			err = helper.MarshalToStruct(goData, &data)
			if err != nil {
				s.log.Error("!!!LoginMiddleware--->LoginDataMarshal", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		if !data.UserFound {
			customError := errors.New("user not found")
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
			UserData:       data.GetUserData(),
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

	// companies, err := s.services.CompanyServiceClient().GetList(ctx, &pbCompany.GetCompanyListRequest{
	// 	Offset:  0,
	// 	Limit:   128,
	// 	OwnerId: req.Data["user_id"],
	// })
	// if err != nil {
	// 	return nil, err
	// }

	// companiesResp := []*pb.Company{}

	// if len(companies.Companies) < 1 {
	// 	companiesById := make([]*pbCompany.Company, 0)
	// 	user, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
	// 		Id: resp.GetUserId(),
	// 	})
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	company, err := s.services.CompanyServiceClient().GetById(ctx, &pbCompany.GetCompanyByIdRequest{
	// 		Id: user.GetCompanyId(),
	// 	})
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	companiesById = append(companiesById, company.Company)
	// 	companies.Companies = companiesById
	// 	companies.Count = int32(len(companiesById))
	// }

	// bytes, err := json.Marshal(companies.GetCompanies())
	// if err != nil {
	// 	return nil, err
	// }

	// err = json.Unmarshal(bytes, &companiesResp)
	// if err != nil {
	// 	return nil, err
	// }

	return &pb.V2LoginWithOptionsResponse{
		UserFound:       true,
		UserId:          req.Data["user_id"],
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
		UserData:        res.GetUserData(),
		// Companies:       companiesResp,
	}, nil
}

func (s *sessionService) V2LoginSuperAdmin(ctx context.Context, req *pb.V2LoginSuperAdminReq) (*pb.V2LoginSuperAdminRes, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2LoginSuperAdmin")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2LoginSuperAdmin", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2LoginSuperAdmin", memoryUsed))
		}
	}()

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
			customError := errors.New("user not found")
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
		err := errors.New("user Not Found")
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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2HasAccess")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2HasAccess", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2HasAccess", memoryUsed))
		}
	}()

	tokenInfo, err := security.ParseClaims(req.AccessToken, s.cfg.SecretKey)
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
		err := errors.New("user has been expired")
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

	services, err := s.serviceNode.GetByNodeType(
		session.ProjectId,
		req.NodeType,
	)
	if err != nil {
		return nil, err
	}

	clientName, ok := convertedClientType["response"].(map[string]interface{})["name"]
	if !ok {
		if clientName == nil {
			err := errors.New("wrong client type")
			s.log.Error("!!!V2HasAccess--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		structPb, err := helper.ConvertMapToStruct(request)
		if err != nil {
			s.log.Error("!!!V2HasAccess--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		if session.ClientTypeId != config.AdminClientPlatformID || clientName.(string) != config.AdminClientName {
			resp, err := services.GetObjectBuilderServiceByType("").GetList(ctx, &pbObject.CommonMessage{
				TableSlug: "record_permission",
				Data:      structPb,
				ProjectId: session.ProjectId,
			})
			if err != nil {
				s.log.Error("!!!V2HasAccess.ObjectBuilderService.GetList--->", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			res, err := helper.ConvertStructToResponse(resp.Data)
			if err != nil {
				s.log.Error("!!!V2HasAccess.ConvertStructToResponse--->", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}

			if len(res["response"].([]interface{})) == 0 {
				err := errors.New("permission denied")
				s.log.Error("!!!V2HasAccess--->", logger.Error(err))
				return nil, status.Error(codes.PermissionDenied, err.Error())
			}

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

func (s *sessionService) V2RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2LoginResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2RefreshToken")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2RefreshToken", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2RefreshToken", memoryUsed))
		}
	}()

	tokenInfo, err := security.ParseClaims(req.RefreshToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	s.log.Info("PROJECT_ID->" + req.ProjectId)
	s.log.Info("REFRESH TOKEN->" + req.RefreshToken)
	s.log.Info(fmt.Sprintf("TOKEN INFO->%+v", tokenInfo))
	s.log.Info(fmt.Sprintf("SECRET KEY->%v", s.cfg.SecretKey))

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.Code(http.Unauthorized.Code), err.Error())
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

	_, err = s.strg.User().GetByUsername(ctx, session.GetUserId())
	if err != nil {
		s.log.Error("!!!V2Login--->", logger.Error(err))
		if err == sql.ErrNoRows {
			errNoRows := errors.New("no user found")
			return nil, status.Error(codes.Internal, errNoRows.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.strg.Session().Update(ctx, &pb.UpdateSessionRequest{
		Id:               session.Id,
		ProjectId:        session.ProjectId,
		ClientPlatformId: session.ClientPlatformId,
		ClientTypeId:     session.ClientTypeId,
		UserId:           session.UserId,
		RoleId:           session.RoleId,
		Ip:               session.Ip,
		Data:             session.Data,
		ExpiresAt:        time.Now().Add(24 * time.Hour).Format(config.DatabaseTimeLayout),
		IsChanged:        session.IsChanged,
		EnvId:            session.EnvId,
	})
	if err != nil {
		s.log.Error("!!!V2RefreshToken.SessionUpdate--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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
		ExpiresAt:        time.Now().Add(24 * time.Hour).Format(config.DatabaseTimeLayout),
		RefreshInSeconds: int32(config.AccessTokenExpiresInTime.Seconds()),
	}

	res := &pb.V2LoginResponse{}
	res.Token = token

	return res, nil
}

func (s *sessionService) V2RefreshTokenSuperAdmin(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2RefreshTokenSuperAdminResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2RefreshTokenSuperAdmin")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2RefreshTokenSuperAdmin", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2RefreshTokenSuperAdmin", memoryUsed))
		}
	}()

	tokenInfo, err := security.ParseClaims(req.RefreshToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!RefreshToken--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.SessionAndTokenGenerator")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the SessionAndTokenGenerator", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("SessionAndTokenGenerator", memoryUsed))
		}
	}()

	if _, err := uuid.Parse(input.GetLoginData().GetUserId()); err != nil {
		err := errors.New("INVALID USER_ID(UUID)" + err.Error())
		s.log.Error("---ERR->GetLoginData().GetUserId-->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO - Delete all old sessions & refresh token has this function too
	_, err := s.strg.Session().DeleteExpiredUserSessions(ctx, input.GetLoginData().GetUserId())
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userSessionList, err := s.strg.Session().GetSessionListByUserID(ctx, input.GetLoginData().GetUserId())
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	input.LoginData.Sessions = userSessionList.GetSessions()

	_, err = uuid.Parse(input.GetProjectId())
	if err != nil {
		input.ProjectId = "8104fded-dfdf-4ee3-87a4-4fa1edca1f68"
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

	// TODO - wrap in a function
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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.UpdateSessionsByRoleId")
	defer dbSpan.Finish()

	rowsAffected, err := s.strg.Session().UpdateByRoleId(ctx, input)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	s.log.Info("UpdateByRoleId--->UpdateByRoleId", logger.Any("rowsAffected", rowsAffected))

	return &emptypb.Empty{}, nil
}

func (s *sessionService) V2MultiCompanyLogin(ctx context.Context, req *pb.V2MultiCompanyLoginReq) (*pb.V2MultiCompanyLoginRes, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2MultiCompanyLogin")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2MultiCompanyLogin", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2MultiCompanyLogin", memoryUsed))
		}
	}()

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
			&pbCompany.GetCompanyByIdRequest{
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
				&pbCompany.GetProjectByIdRequest{
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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2HasAccessUser")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2HasAccessUser", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2HasAccessUser", memoryUsed))
		}
	}()

	s.log.Info("\n!!!V2HasAccessUser--->", logger.Any("req", req))

	arr_path := strings.Split(req.Path, "/")

	tokenInfo, err := security.ParseClaims(req.AccessToken, s.cfg.SecretKey)
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

	if expiresAt.Unix() < time.Now().Add(5*time.Hour).Unix() {
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
	// this condition need our object/get-list-group-by and object/get-group-by-field api because this api's method is post we change it to get
	if ((strings.Contains(req.GetPath(), "object/get-list")) || (strings.Contains(req.GetPath(), "object/get-list-group-by")) || (strings.Contains(req.GetPath(), "object/get-group-by-field"))) && req.GetMethod() != "GET" {
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
	for _, item := range projects.GetUserProjects() {
		if item.ProjectId == session.GetProjectId() {
			exist = true
			break
		}
	}
	if !exist {
		err = errors.New("---V2HasAccessUser->Access denied")
		s.log.Error("---V2HasAccessUser--->AccessDenied--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.EnvironmentId != "" {
		exist = false
		for _, item := range projects.GetUserProjects() {
			if item.EnvId == req.EnvironmentId {
				exist = true
				break
			}
		}

		if !exist {
			err = errors.New("user not access environment")
			s.log.Error("---V2HasAccessUser--->AccessNotEnvironment--->", logger.Error(err))
			return nil, status.Error(codes.Unavailable, err.Error())
		}
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
		if strings.Contains(arr_path[len(arr_path)-1], ":") {
			tableSlug = arr_path[len(arr_path)-2]
		} else {
			tableSlug = arr_path[len(arr_path)-1]
		}

		resource, err := s.services.ServiceResource().GetSingle(
			ctx,
			&pbCompany.GetSingleServiceResourceReq{
				ProjectId:     session.ProjectId,
				EnvironmentId: session.EnvId,
				ServiceType:   pbCompany.ServiceType_BUILDER_SERVICE,
			},
		)
		if err != nil {
			return nil, err
		}

		switch resource.ResourceType {
		case pbCompany.ResourceType_MONGODB:
			services, err := s.serviceNode.GetByNodeType(
				resource.ProjectId,
				resource.NodeType,
			)
			if err != nil {
				return nil, err
			}

			resp, err := services.GetBuilderPermissionServiceByType(resource.NodeType).GetTablePermission(
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
		case pbCompany.ResourceType_POSTGRESQL:

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
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2MultiCompanyOneLogin")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2MultiCompanyOneLogin", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2MultiCompanyOneLogin", memoryUsed))
		}
	}()

	resp := pb.V2MultiCompanyOneLoginRes{
		Companies: []*pb.Company2{},
	}

	user := &pb.User{}
	var err error

	switch req.Type {
	case config.Default:
		{
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

			user, err = s.strg.User().GetByUsername(ctx, req.GetUsername())
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
				s.log.Error("!!!MultiCompanyLogin Default--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	case config.WithPhone:
		{
			if config.DefaultOtp != req.Otp {
				_, err := s.services.SmsService().ConfirmOtp(
					ctx,
					&sms_service.ConfirmOtpRequest{
						SmsId: req.GetSmsId(),
						Otp:   req.GetOtp(),
					},
				)
				if err != nil {
					return nil, err
				}
			}

			user, err = s.strg.User().GetByUsername(ctx, req.GetPhone())
			if err != nil {
				s.log.Error("!!!MultiCompanyLogin Phone--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	case config.WithEmail:
		{
			if config.DefaultOtp != req.Otp {
				_, err := s.services.SmsService().ConfirmOtp(
					ctx,
					&sms_service.ConfirmOtpRequest{
						SmsId: req.GetSmsId(),
						Otp:   req.GetOtp(),
					},
				)
				if err != nil {
					return nil, err
				}
			}

			user, err = s.strg.User().GetByUsername(ctx, req.GetEmail())
			if err != nil {
				s.log.Error("!!!MultiCompanyLogin Email--->", logger.Error(err))
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
	}

	userProjects, err := s.strg.User().GetUserProjects(ctx, user.GetId())
	if err != nil {
		errGetProjects := errors.New("cant get user projects")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, errGetProjects.Error())
	}

	userEnvProject, err := s.strg.User().GetUserEnvProjects(ctx, user.GetId())
	if err != nil {
		errGetEnvProjects := errors.New("cant get user env projects")
		s.log.Error("!!!MultiCompanyLogin--->", logger.Error(err))
		return nil, status.Error(codes.NotFound, errGetEnvProjects.Error())
	}

	for _, item := range userProjects.Companies {
		projects := make([]*pb.Project2, 0, 20)
		company, err := s.services.CompanyServiceClient().GetById(ctx,
			&pbCompany.GetCompanyByIdRequest{
				Id: item.Id,
			})

		if err != nil {
			errGetProjects := errors.New("cant get user projects")
			s.log.Error("!!!MultiCompanyLogin--->GetById", logger.Error(err))
			return nil, status.Error(codes.NotFound, errGetProjects.Error())
		}

		for _, projectId := range item.ProjectIds {

			clientType, _ := s.strg.User().GetUserProjectClientTypes(
				ctx,
				&models.UserProjectClientTypeRequest{
					UserId:    user.GetId(),
					ProjectId: projectId,
				},
			)

			projectInfo, err := s.services.ProjectServiceClient().GetById(
				ctx,
				&pbCompany.GetProjectByIdRequest{
					ProjectId: projectId,
					CompanyId: item.Id,
				})
			if err != nil {
				errGetProjects := errors.New("cant get user projects")
				s.log.Error("!!!MultiCompanyLogin---->ProjectInfo", logger.Error(err))
				return nil, status.Error(codes.NotFound, errGetProjects.Error())
			}

			resProject := &pb.Project2{
				Id:        projectInfo.GetProjectId(),
				CompanyId: projectInfo.GetCompanyId(),
				Name:      projectInfo.GetTitle(),
				Domain:    projectInfo.GetK8SNamespace(),
			}

			environments, err := s.services.EnvironmentService().GetList(
				ctx,
				&pbCompany.GetEnvironmentListRequest{
					ProjectId: projectId,
					Limit:     1000,
					Ids:       userEnvProject.EnvProjects[projectId],
				},
			)
			if err != nil {
				errGetProjects := errors.New("cant get environments")
				s.log.Error("!!!MultiCompanyLogin--->EnvironmentsList", logger.Error(err))
				return nil, status.Error(codes.NotFound, errGetProjects.Error())
			}

			for _, en := range environments.Environments {
				resourceEnv, err := s.services.ServiceResource().GetList(
					ctx,
					&pbCompany.GetListServiceResourceReq{
						ProjectId:     projectId,
						EnvironmentId: en.Id,
					},
				)
				if err != nil {
					errGetProjects := errors.New("cant get resourse environments")
					s.log.Error("!!!MultiCompanyLogin--->ServiceResourceList", logger.Error(err))
					return nil, status.Error(codes.NotFound, errGetProjects.Error())
				}

				respResourceEnvironment := &pb.ResourceEnvironmentV2MultiCompany{
					Id:            resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceEnvironmentId,
					Name:          en.Name,
					ProjectId:     en.ProjectId,
					ResourceId:    resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceId,
					EnvironmentId: en.Id,
					IsConfigured:  true,
					ResourceType:  int32(resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceType.Number()),
					ServiceType:   int32(resourceEnv.ServiceResources[config.ObjectBuilderService].ServiceType.Number()),
					DisplayColor:  en.DisplayColor,
					Description:   en.Description,
					AccessType:    en.AccessType,
				}

				if resourceEnv.ServiceResources[config.ObjectBuilderService] == nil || resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceEnvironmentId == "" {
					continue
				}

				if clientType == nil || len(clientType.ClientTypeIds) == 0 {

					clientTypes, err := s.services.ClientService().V2GetClientTypeList(
						ctx,
						&pb.V2GetClientTypeListRequest{
							ProjectId:              resourceEnv.ServiceResources[config.ObjectBuilderService].ProjectId,
							ResourceType:           int32(resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceType.Number()),
							NodeType:               resourceEnv.ServiceResources[config.ObjectBuilderService].NodeType,
							ResourceEnvrironmentId: resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceEnvironmentId,
						},
					)
					if err != nil {
						errGetProjects := errors.New("cant get client types")
						s.log.Error("!!!MultiCompanyLogin--->ClientTypes", logger.Error(err))
						return nil, status.Error(codes.NotFound, errGetProjects.Error())
					}
					respResourceEnvironment.ClientTypes = clientTypes.Data

				} else if len(clientType.ClientTypeIds) > 0 {
					clientTypes, err := s.services.ClientService().V2GetClientTypeList(
						ctx,
						&pb.V2GetClientTypeListRequest{
							ProjectId:              resourceEnv.ServiceResources[config.ObjectBuilderService].ProjectId,
							ResourceType:           int32(resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceType.Number()),
							Guids:                  clientType.ClientTypeIds,
							NodeType:               resourceEnv.ServiceResources[config.ObjectBuilderService].NodeType,
							ResourceEnvrironmentId: resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceEnvironmentId,
						},
					)
					if err != nil {
						errGetProjects := errors.New("cant get client types")
						s.log.Error("!!!MultiCompanyLogin--->ClientTypes2", logger.Error(err))
						return nil, status.Error(codes.NotFound, errGetProjects.Error())
					}
					respResourceEnvironment.ClientTypes = clientTypes.Data
				}

				resProject.ResourceEnvironments = append(resProject.ResourceEnvironments, respResourceEnvironment)
			}

			projects = append(projects, resProject)
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
	resp.UserId = user.Id

	return &resp, nil
}

func (s *sessionService) V2ResetPassword(ctx context.Context, req *pb.V2ResetPasswordRequest) (*pb.User, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2ResetPassword")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2ResetPassword", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2ResetPassword", memoryUsed))
		}
	}()

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

	return s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{Id: req.GetUserId()})
}

func (s *sessionService) V2RefreshTokenForEnv(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2LoginResponse, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.V2RefreshTokenForEnv")
	defer dbSpan.Finish()

	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	defer func() {
		var after runtime.MemStats
		runtime.ReadMemStats(&after)
		memoryUsed := (after.TotalAlloc - before.TotalAlloc) / (1024 * 1024)
		s.log.Info("Memory used by the V2RefreshTokenForEnv", logger.Any("memoryUsed", memoryUsed))
		if memoryUsed > 300 {
			s.log.Info("Memory used over 300 mb", logger.Any("V2RefreshTokenForEnv", memoryUsed))
		}
	}()

	tokenInfo, err := security.ParseClaims(req.RefreshToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(err))
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

	user, err := s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
		Id: session.GetUserId(),
	})
	if err != nil {
		s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(err))
		if err == sql.ErrNoRows {
			errNoRows := errors.New("no user found")
			return nil, status.Error(codes.Internal, errNoRows.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	resource, err := s.services.ServiceResource().GetSingle(
		ctx,
		&pbCompany.GetSingleServiceResourceReq{
			ProjectId:     session.ProjectId,
			EnvironmentId: req.EnvId,
			ServiceType:   pbCompany.ServiceType_BUILDER_SERVICE,
		},
	)
	if err != nil {
		s.log.Error("!!!V2Refresh.SessionService().GetServiceResource--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	services, err := s.serviceNode.GetByNodeType(
		resource.ProjectId,
		resource.NodeType,
	)
	if err != nil {
		return nil, err
	}

	reqLoginData := &pbObject.LoginDataReq{
		UserId:                user.GetId(),
		ClientType:            session.GetClientTypeId(),
		ProjectId:             session.GetProjectId(),
		ResourceEnvironmentId: resource.GetResourceEnvironmentId(),
	}

	var data *pbObject.LoginDataRes

	switch resource.ResourceType {
	case 1:
		data, err = services.GetLoginServiceByType(resource.NodeType).LoginData(
			ctx,
			reqLoginData,
		)

		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(err))
			return nil, status.Error(codes.Internal, errGetUserProjectData.Error())
		}
	case 3:
		data, err = services.PostgresLoginService().LoginData(
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
		customError := fmt.Errorf("user not found with env_id %s, user_id %s, client_type_id %s", req.GetEnvId(), user.Id, session.ClientTypeId)
		s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(customError))
		return nil, status.Error(codes.NotFound, customError.Error())
	}

	userResp := helper.ConvertPbToAnotherPb(&pbObject.V2LoginResponse{
		ClientPlatform:   data.GetClientPlatform(),
		ClientType:       data.GetClientType(),
		UserFound:        data.GetUserFound(),
		UserId:           data.GetUserId(),
		Role:             data.GetRole(),
		Permissions:      data.GetPermissions(),
		LoginTableSlug:   data.GetLoginTableSlug(),
		AppPermissions:   data.GetAppPermissions(),
		GlobalPermission: data.GetGlobalPermission(),
		UserData:         data.GetUserData(),
	})

	roleId := ""
	if userRole, ok := userResp.UserData.Fields["role_id"].GetKind().(*structpb.Value_StringValue); ok {

		roleId = userRole.StringValue
	}

	_, err = s.strg.Session().Update(ctx, &pb.UpdateSessionRequest{
		Id:               session.Id,
		ProjectId:        session.ProjectId,
		ClientPlatformId: session.ClientPlatformId,
		ClientTypeId:     session.ClientTypeId,
		UserId:           session.UserId,
		RoleId:           roleId,
		Ip:               session.Ip,
		Data:             session.Data,
		ExpiresAt:        session.ExpiresAt,
		IsChanged:        session.IsChanged,
		EnvId:            req.EnvId,
	})
	if err != nil {
		s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

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
		s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken, err := security.GenerateJWT(m, config.RefreshTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(err))
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

	res := &pb.V2LoginResponse{}
	res.Token = token

	return res, nil
}

func (s *sessionService) ExpireSessions(ctx context.Context, req *pb.ExpireSessionsRequest) (*emptypb.Empty, error) {
	dbSpan, ctx := opentracing.StartSpanFromContext(ctx, "grpc_session_v2.ExpireSessions")
	defer dbSpan.Finish()

	s.log.Info("---ExpireSessions--->>>", logger.Any("req", req.SessionIds))

	err := s.strg.Session().ExpireSessions(ctx, req)
	if err != nil {
		s.log.Error("!!!ExpireSessiona--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}
