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
	span "ucode/ucode_go_auth_service/pkg/jaeger"
	"ucode/ucode_go_auth_service/pkg/security"

	"github.com/google/uuid"
	"github.com/saidamir98/udevs_pkg/logger"
	"github.com/spf13/cast"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *sessionService) V2Login(ctx context.Context, req *pb.V2LoginRequest) (*pb.V2LoginResponse, error) {
	s.log.Info("V2Login --> ", logger.Any("request: ", req))

	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.V2Login", req)
	defer dbSpan.Finish()

	var (
		user   = &pb.User{}
		err    error
		data   *pbObject.LoginDataRes
		before runtime.MemStats
	)

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

	switch req.Type {
	case config.Default:
		if len(req.Username) < 6 {
			err := errors.New("invalid username")
			s.log.Error("!!!V2Login--->InvalidUsername", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if len(req.Password) < 6 {
			err := errors.New("invalid password")
			s.log.Error("!!!V2Login--->InvalidPassword", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		user, err = s.strg.User().V2GetByUsername(ctx, req.GetUsername(), config.WithLogin)
		if err != nil {
			s.log.Error("!!!V2Login--->GetByUsername", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		hashType := user.GetHashType()
		if config.HashTypes[hashType] == 1 {
			match, err := security.ComparePassword(user.GetPassword(), req.Password)
			if err != nil {
				s.log.Error("!!!MultiCompanyLogin-->ComparePasswordArgon", logger.Error(err))
				return nil, err
			}
			if !match {
				err := errors.New("username or password is wrong")
				s.log.Error("!!!MultiCompanyOneLogin-->Wrong", logger.Error(err))
				return nil, err
			}

			go func() {
				hashedPassword, err := security.HashPasswordBcrypt(req.Password)
				if err != nil {
					s.log.Error("!!!MultiCompanyOneLogin--->HashPasswordBcryptGo", logger.Error(err))
					return
				}

				err = s.strg.User().UpdatePassword(context.Background(), user.Id, hashedPassword)
				if err != nil {
					s.log.Error("!!!MultiCompanyOneLogin--->HashPasswordBcryptGo", logger.Error(err))
					return
				}
			}()
		} else if config.HashTypes[hashType] == 2 {
			match, err := security.ComparePasswordBcrypt(user.GetPassword(), req.Password)
			if err != nil {
				s.log.Error("!!!MultiCompanyOneLogin-->ComparePasswordBcrypt", logger.Error(err))
				return nil, err
			}
			if !match {
				err := errors.New("username or password is wrong")
				s.log.Error("!!!MultiCompanyOneLogin--->", logger.Error(err))
				return nil, err
			}
		} else {
			err := errors.New("invalid hash type")
			s.log.Error("!!!MultiCompanyOneLogin--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case config.WithPhone:
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
	case config.WithEmail:
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
			s.log.Error("!!!V2Login Email--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case config.WithGoogle:
		var email string
		if req.GetGoogleToken() != "" {
			userInfo, err := helper.GetGoogleUserInfo(req.GetGoogleToken())
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
			email = cast.ToString(userInfo["email"])
		} else {
			err := errors.New("google token is required when login type is google auth")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		user, err = s.strg.User().GetByUsername(ctx, email)
		if err != nil {
			s.log.Error("!!!V2Login Email--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	userStatus, err := s.strg.User().GetUserStatus(ctx, user.Id, req.GetProjectId())
	if err != nil {
		s.log.Error("!!!V2Login--->GetUserStatus", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if userStatus == config.UserStatusBlocked {
		err := errors.New("user blocked")
		s.log.Error("!!!V2Login--->UserBlocked", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	reqLoginData := &pbObject.LoginDataReq{
		UserId:                user.GetId(),
		ClientType:            req.GetClientType(),
		ProjectId:             req.GetProjectId(),
		ResourceEnvironmentId: req.GetResourceEnvironmentId(),
	}

	services, err := s.serviceNode.GetByNodeType(req.ProjectId, req.NodeType)
	if err != nil {
		return nil, err
	}

	switch req.ResourceType {
	case 1:
		data, err = services.GetLoginServiceByType(req.NodeType).LoginData(ctx, reqLoginData)
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

		newData, err := services.GoLoginService().LoginData(ctx, &newReq)
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
		Role:             data.GetRole(),
		UserId:           data.GetUserId(),
		UserData:         data.GetUserData(),
		UserFound:        data.GetUserFound(),
		ClientType:       data.GetClientType(),
		UserIdAuth:       data.GetUserIdAuth(),
		Permissions:      data.GetPermissions(),
		ClientPlatform:   data.GetClientPlatform(),
		LoginTableSlug:   data.GetLoginTableSlug(),
		AppPermissions:   data.GetAppPermissions(),
		GlobalPermission: data.GetGlobalPermission(),
	})

	resp, err := s.SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
		LoginData:     res,
		Tables:        req.Tables,
		ProjectId:     req.GetProjectId(),
		EnvironmentId: req.GetEnvironmentId(),
		ClientId:      req.GetClientId(),
		ClientIp:      req.GetClientIp(),
		UserAgent:     req.GetUserAgent(),
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
	s.log.Info("V2LoginWithOption-->", logger.Any("request: ", req))

	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.V2LoginWithOption", req)
	defer dbSpan.Finish()

	var (
		before   runtime.MemStats
		userId   string
		verified bool
		user     *pb.User
		err      error
	)
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

		user, err = s.strg.User().GetByUsername(ctx, username)
		if err != nil {
			s.log.Error("!!!V2V2LoginWithOption--->", logger.Error(err))
			if err == sql.ErrNoRows {
				errNoRows := errors.New("no user found")
				return nil, status.Error(codes.Internal, errNoRows.Error())
			}
			return nil, status.Error(codes.Internal, err.Error())
		}

		userId = user.Id
	case "PHONE":
		phone, ok := req.GetData()["phone"]
		if !ok {
			err := errors.New("phone is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		user, err = s.strg.User().GetByUsername(ctx, phone)
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

		user, err = s.strg.User().GetByUsername(ctx, email)
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

		user, err = s.strg.User().GetByUsername(ctx, username)
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
			ctx, &pbCompany.GetProjectResourceListRequest{
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
				ctx, &sms_service.ConfirmOtpRequest{
					SmsId: sms_id, Otp: otp,
				},
			)
			if err != nil {
				s.log.Error("!!!V2LoginWithOption--->ConfirmOTP", logger.Error(err))
				return nil, err
			}
		}
		verified = true

		user, err = s.strg.User().GetByUsername(ctx, phone)
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

		emailOtpSettings, err := s.services.ResourceService().GetProjectResourceList(
			ctx, &pbCompany.GetProjectResourceListRequest{
				EnvironmentId: req.Data["environment_id"],
				ProjectId:     req.Data["project_id"],
				Type:          pbCompany.ResourceType_SMTP,
			},
		)
		if err != nil {
			s.log.Error("!!!V2LoginWithOption.EmailtpSettingsService().GetList--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		var defaultOtp string
		if len(emailOtpSettings.GetResources()) > 0 {
			if emailOtpSettings.GetResources()[0].GetSettings().GetSmtp().GetDefaultOtp() != "" {
				defaultOtp = emailOtpSettings.GetResources()[0].GetSettings().GetSmtp().GetDefaultOtp()
			}
		}

		if otp != defaultOtp {
			_, err := s.services.SmsService().ConfirmOtp(
				ctx, &sms_service.ConfirmOtpRequest{
					SmsId: sms_id,
					Otp:   otp,
				},
			)
			if err != nil {
				return nil, err
			}
		}
		verified = true

		user, err = s.strg.User().GetByUsername(ctx, email)
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

		user, err = s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{
			Id: userIdRes.GetId(),
		})
		if err != nil {
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

		userId = user.GetId()
	case "GOOGLE_AUTH":
		email, ok := req.GetData()["email"]
		if !ok {
			err := errors.New("email is empty")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if gooleToken, ok := req.GetData()["google_token"]; ok {
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

	userStatus, err := s.strg.User().GetUserStatus(ctx, user.Id, req.Data["project_id"])
	if err != nil {
		s.log.Error("!!!V2Login--->GetUserStatus", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	if userStatus == config.UserStatusBlocked {
		err := errors.New("user blocked")
		s.log.Error("!!!V2Login--->UserBlocked", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	req.Data["user_id"] = userId
	data, err := s.LoginMiddleware(ctx, models.LoginMiddlewareReq{
		Data:      req.Data,
		Tables:    req.Tables,
		ClientId:  req.ClientId,
		ClientIp:  req.GetClientIp(),
		UserAgent: req.GetUserAgent(),
	})
	if err != nil {
		var httpErrorStr = ""

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
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.LoginMiddleware", req)
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
		var data *pbObject.LoginDataRes

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
			NodeType:              serviceResource.GetNodeType(),
			ProjectId:             req.Data["project_id"],
			ClientType:            req.Data["client_type_id"],
			ResourceEnvironmentId: serviceResource.GetResourceEnvironmentId(),
			Password:              req.Data["password"],
		}

		services, err := s.serviceNode.GetByNodeType(req.Data["project_id"], req.NodeType)
		if err != nil {
			s.log.Error("!!!LoginMiddleware--->GetByNodeType", logger.Error(err))
			return nil, err
		}

		switch serviceResource.ResourceType {
		case 1:
			data, err = services.GetLoginServiceByType(req.NodeType).LoginData(ctx, reqLoginData)
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

			if err = helper.MarshalToStruct(goData, &data); err != nil {
				s.log.Error("!!!LoginMiddleware--->LoginDataMarshal", logger.Error(err))
				return nil, status.Error(codes.Internal, err.Error())
			}
		}

		if !data.UserFound {
			customError := errors.New("user not found")
			s.log.Error("!!!LoginMiddleware--->", logger.Error(customError))
			return nil, status.Error(codes.NotFound, customError.Error())
		}

		if !data.ComparePassword {
			err := errors.New("invalid password")
			s.log.Error("!!!V2LoginWithOption--->", logger.Error(err))
			return nil, err
		}

		res = helper.ConvertPbToAnotherPb(&pbObject.V2LoginResponse{
			Role:           data.GetRole(),
			UserId:         data.GetUserId(),
			UserData:       data.GetUserData(),
			UserFound:      data.GetUserFound(),
			UserIdAuth:     data.GetUserIdAuth(),
			ClientType:     data.GetClientType(),
			Permissions:    data.GetPermissions(),
			ClientPlatform: data.GetClientPlatform(),
			LoginTableSlug: data.GetLoginTableSlug(),
		})

	}
	if req.Tables == nil {
		req.Tables = []*pb.Object{}
	}

	resp, err := s.SessionAndTokenGenerator(ctx, &pb.SessionAndTokenRequest{
		Tables:        req.Tables,
		LoginData:     res,
		ProjectId:     req.Data["project_id"],
		EnvironmentId: req.Data["environment_id"],
		ClientId:      req.ClientId,
		ClientIp:      req.ClientIp,
		UserAgent:     req.UserAgent,
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

	return &pb.V2LoginWithOptionsResponse{
		User:            resp.GetUser(),
		Role:            resp.GetRole(),
		Token:           resp.GetToken(),
		Tables:          resp.GetTables(),
		UserId:          resp.GetUserId(),
		Sessions:        resp.GetSessions(),
		UserData:        res.GetUserData(),
		UserFound:       true,
		ResourceId:      resp.GetResourceId(),
		ClientType:      resp.GetClientType(),
		Permissions:     resp.GetPermissions(),
		EnvironmentId:   resp.GetEnvironmentId(),
		ClientPlatform:  resp.GetClientPlatform(),
		AppPermissions:  resp.GetAppPermissions(),
		LoginTableSlug:  resp.GetLoginTableSlug(),
		AddationalTable: resp.GetAddationalTable(),
	}, nil
}

func (s *sessionService) V2RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2LoginResponse, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.V2RefreshToken", req)
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
		s.log.Error("!!!RefreshToken--->ParseClaims", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	session, err := s.strg.Session().GetByPK(ctx, &pb.SessionPrimaryKey{Id: tokenInfo.ID})
	if err != nil {
		s.log.Error("!!!RefreshToken--->SessionGetByPK", logger.Error(err))
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

	_, err = s.strg.User().CHeckUserProject(ctx, session.GetUserIdAuth(), session.GetProjectId())
	if err != nil {
		s.log.Error("!!!V2Login--->CHeckUserProject", logger.Error(err))
		if err == sql.ErrNoRows {
			errNoRows := errors.New("no user found")
			return nil, status.Error(codes.Internal, errNoRows.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	_, err = s.strg.Session().Update(ctx, &pb.UpdateSessionRequest{
		Id:               session.Id,
		Ip:               session.Ip,
		Data:             session.Data,
		EnvId:            session.EnvId,
		UserId:           session.UserId,
		RoleId:           session.RoleId,
		ProjectId:        session.ProjectId,
		IsChanged:        session.IsChanged,
		ClientTypeId:     session.ClientTypeId,
		ClientPlatformId: session.ClientPlatformId,
	})
	if err != nil {
		s.log.Error("!!!V2RefreshToken.SessionUpdate--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var authTables []*pb.TableBody

	if tokenInfo.Tables != nil {
		for _, table := range tokenInfo.Tables {
			authTables = append(authTables, &pb.TableBody{
				TableSlug: table.TableSlug,
				ObjectId:  table.ObjectID,
			})
		}
	}

	// TODO - wrap in a function
	m := map[string]interface{}{
		"id":                 session.Id,
		"ip":                 session.Data,
		"data":               session.Data,
		"tables":             authTables,
		"user_id":            session.UserId,
		"role_id":            session.RoleId,
		"project_id":         session.ProjectId,
		"user_id_auth":       session.UserIdAuth,
		"client_type_id":     session.ClientTypeId,
		"login_table_slug":   tokenInfo.LoginTableSlug,
		"client_platform_id": session.ClientPlatformId,
	}

	accessToken, err := security.GenerateJWT(m, config.AccessTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->GenerateAccessJWT", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	refreshToken, err := security.GenerateJWT(m, config.RefreshTokenExpiresInTime, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!RefreshToken--->GenerateRefreshJWT", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &pb.V2LoginResponse{
		Token: &pb.Token{
			AccessToken:      accessToken,
			RefreshToken:     refreshToken,
			CreatedAt:        session.CreatedAt,
			UpdatedAt:        session.UpdatedAt,
			ExpiresAt:        time.Now().Add(24 * time.Hour).Format(config.DatabaseTimeLayout),
			RefreshInSeconds: int32(config.AccessTokenExpiresInTime.Seconds()),
		},
	}

	return res, nil
}

func (s *sessionService) SessionAndTokenGenerator(ctx context.Context, input *pb.SessionAndTokenRequest) (*pb.V2LoginResponse, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.SessionAndTokenGenerator", input)
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

	if _, err := uuid.Parse(input.GetLoginData().GetUserIdAuth()); err != nil {
		err := errors.New("INVALID USER_ID(UUID)" + err.Error())
		s.log.Error("!!!TokenGenerator->UserIdAuthExist-->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// TODO - Delete all old sessions & refresh token has this function too
	_, err := s.strg.Session().DeleteExpiredUserSessions(ctx, input.GetLoginData().GetUserIdAuth())
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	userSessionList, err := s.strg.Session().GetSessionListByUserID(ctx, input.GetLoginData().GetUserIdAuth())
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	input.LoginData.Sessions = userSessionList.GetSessions()

	_, err = uuid.Parse(input.GetProjectId())
	if err != nil {
		err = errors.New("project id is invalid")
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	sessionPKey, err := s.strg.Session().Create(ctx, &pb.CreateSessionRequest{
		Ip:               input.GetClientIp(),
		Data:             input.GetUserAgent(),
		EnvId:            input.GetEnvironmentId(),
		UserId:           input.GetLoginData().GetUserId(),
		RoleId:           input.GetLoginData().GetRole().GetId(),
		ClientId:         input.GetClientId(),
		ProjectId:        input.GetProjectId(),
		ExpiresAt:        time.Now().Add(config.RefreshTokenExpiresInTime).Format(config.DatabaseTimeLayout),
		UserIdAuth:       input.GetLoginData().GetUserIdAuth(),
		ClientTypeId:     input.GetLoginData().GetClientType().GetId(),
		ClientPlatformId: input.GetLoginData().GetClientPlatform().GetId(),
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
		Id:        input.GetLoginData().GetUserIdAuth(),
	})
	if err != nil {
		s.log.Error("!!!Login->GetByPK--->", logger.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// TODO - wrap in a function
	m := map[string]any{
		"id":                 session.GetId(),
		"ip":                 session.GetData(),
		"data":               session.GetData(),
		"tables":             input.GetTables(),
		"user_id":            session.GetUserId(),
		"role_id":            session.GetRoleId(),
		"client_id":          session.GetClientId(),
		"project_id":         session.GetProjectId(),
		"user_id_auth":       session.GetUserIdAuth(),
		"client_type_id":     session.GetClientTypeId(),
		"login_table_slug":   input.GetLoginData().GetLoginTableSlug(),
		"client_platform_id": session.GetClientPlatformId(),
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

	go func() {
		err = s.strg.ApiKeys().CreateClientToken(context.Background(), input.ClientId, m)
	}()

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
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.UpdateSessionsByRoleId", input)
	defer dbSpan.Finish()

	_, err := s.strg.Session().UpdateByRoleId(ctx, input)
	if err != nil {
		s.log.Error("!!!Login--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *sessionService) V2HasAccessUser(ctx context.Context, req *pb.V2HasAccessUserReq) (*pb.V2HasAccessUserRes, error) {
	s.log.Info("!!!V2HasAccessUser--->", logger.Any("req", req))

	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.V2HasAccessUser", req)
	defer dbSpan.Finish()

	var (
		before                 runtime.MemStats
		arr_path               = strings.Split(req.Path, "/")
		methodField            string
		exist, checkPermission bool
		authTables             []*pb.TableBody
	)
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

	tokenInfo, err := security.ParseClaims(req.AccessToken, s.cfg.SecretKey)
	if err != nil {
		s.log.Error("!!!V2HasAccessUser->ParseClaims--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if tokenInfo.ClientID != "" {
		stats, err := s.strg.ApiKeys().CheckClientIdStatus(ctx, tokenInfo.ClientID)
		if err != nil {
			s.log.Error("!!!V2HasAccessUser->CheckClientIdStatus--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if !stats {
			err = config.ErrInactiveClientId
			s.log.Error("!!!V2HasAccessUser->InactiveClientId--->", logger.Error(err))
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
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
	if ((strings.Contains(req.GetPath(), "object/get-list")) ||
		(strings.Contains(req.GetPath(), "object/get-list-group-by")) ||
		(strings.Contains(req.GetPath(), "object/get-group-by-field"))) && req.GetMethod() != "GET" {
		methodField = "read"
	}

	projects, err := s.services.UserService().GetProjectsByUserId(ctx, &pb.GetProjectsByUserIdReq{
		UserId: session.GetUserIdAuth(),
	})
	if err != nil {
		s.log.Error("---V2HasAccessUser->GetProjectsByUserId--->", logger.Error(err))
		return nil, err
	}
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

	for _, path := range arr_path {
		if exist := config.Path[path]; exist {
			checkPermission = exist
		}
	}

	if checkPermission {
		var tableSlug string
		if strings.Contains(arr_path[len(arr_path)-1], ":") {
			tableSlug = arr_path[len(arr_path)-2]
		} else {
			tableSlug = arr_path[len(arr_path)-1]
		}

		resource, err := s.services.ServiceResource().GetSingle(ctx,
			&pbCompany.GetSingleServiceResourceReq{
				ProjectId:     session.ProjectId,
				EnvironmentId: session.EnvId,
				ServiceType:   pbCompany.ServiceType_BUILDER_SERVICE,
			},
		)
		if err != nil {
			s.log.Error("!!!V2HasAccessUser->GetSingleServiceResource--->", logger.Error(err))
			return nil, err
		}

		switch resource.ResourceType {
		case pbCompany.ResourceType_MONGODB:
			services, err := s.serviceNode.GetByNodeType(resource.ProjectId, resource.NodeType)
			if err != nil {
				return nil, err
			}

			resp, err := services.GetBuilderPermissionServiceByType(resource.NodeType).GetTablePermission(ctx,
				&pbObject.GetTablePermissionRequest{
					TableSlug:             tableSlug,
					RoleId:                session.RoleId,
					ResourceEnvironmentId: resource.ResourceEnvironmentId,
					Method:                methodField,
				},
			)
			if err != nil {
				s.log.Error("!!!V2HasAccessUser->GetTablePermission--->", logger.Error(err))
				return nil, err
			}

			if !resp.IsHavePermission {
				err := status.Error(codes.PermissionDenied, "Permission denied")
				return nil, err
			}
		case pbCompany.ResourceType_POSTGRESQL:
		}
	}

	for _, table := range tokenInfo.Tables {
		authTables = append(authTables, &pb.TableBody{
			TableSlug: table.TableSlug,
			ObjectId:  table.ObjectID,
		})
	}

	return &pb.V2HasAccessUserRes{
		Id:               session.Id,
		Ip:               session.Ip,
		Data:             session.Data,
		EnvId:            session.EnvId,
		UserId:           session.UserId,
		RoleId:           session.RoleId,
		Tables:           authTables,
		ProjectId:        session.ProjectId,
		ExpiresAt:        session.ExpiresAt,
		CreatedAt:        session.CreatedAt,
		UpdatedAt:        session.UpdatedAt,
		UserIdAuth:       session.UserIdAuth,
		ClientTypeId:     session.ClientTypeId,
		ClientPlatformId: session.ClientPlatformId,
	}, nil
}

func (s *sessionService) V2MultiCompanyOneLogin(ctx context.Context, req *pb.V2MultiCompanyLoginReq) (*pb.V2MultiCompanyOneLoginRes, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.V2MultiCompanyOneLogin", req)
	defer dbSpan.Finish()

	var (
		before runtime.MemStats
		user   = &pb.User{}
		err    error
		resp   = pb.V2MultiCompanyOneLoginRes{Companies: []*pb.Company2{}}
	)

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

	switch req.Type {
	case config.Default:
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

		user, err = s.strg.User().V2GetByUsername(ctx, req.GetUsername(), config.WithLogin)
		if err != nil {
			s.log.Error("!!!MultiCompanyLogin--->UserGetByUsername", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}

		hashType := user.GetHashType()
		if config.HashTypes[hashType] == 1 {
			match, err := security.ComparePassword(user.GetPassword(), req.Password)
			if err != nil {
				s.log.Error("!!!MultiCompanyLogin-->ComparePasswordArgon", logger.Error(err))
				return nil, err
			}
			if !match {
				err := errors.New("username or password is wrong")
				s.log.Error("!!!MultiCompanyOneLogin-->Wrong", logger.Error(err))
				return nil, err
			}

			go func() {
				hashedPassword, err := security.HashPasswordBcrypt(req.Password)
				if err != nil {
					s.log.Error("!!!MultiCompanyOneLogin--->HashPasswordBcryptGo", logger.Error(err))
					return
				}
				err = s.strg.User().UpdatePassword(context.Background(), user.Id, hashedPassword)
				if err != nil {
					s.log.Error("!!!MultiCompanyOneLogin--->HashPasswordBcryptGo", logger.Error(err))
					return
				}
			}()
		} else if config.HashTypes[hashType] == 2 {
			match, err := security.ComparePasswordBcrypt(user.GetPassword(), req.Password)
			if err != nil {
				s.log.Error("!!!MultiCompanyOneLogin-->ComparePasswordBcrypt", logger.Error(err))
				return nil, err
			}
			if !match {
				err := errors.New("username or password is wrong")
				s.log.Error("!!!MultiCompanyOneLogin--->", logger.Error(err))
				return nil, err
			}
		} else {
			err := errors.New("invalid hash type")
			s.log.Error("!!!MultiCompanyOneLogin--->", logger.Error(err))
			return nil, status.Error(codes.Internal, err.Error())
		}
	case config.WithPhone:
		if config.DefaultOtp != req.Otp {
			_, err := s.services.SmsService().ConfirmOtp(ctx, &sms_service.ConfirmOtpRequest{
				SmsId: req.GetSmsId(),
				Otp:   req.GetOtp(),
			})
			if err != nil {
				return nil, err
			}
		}

		user, err = s.strg.User().GetByUsername(ctx, req.GetPhone())
		if err != nil {
			s.log.Error("!!!MultiCompanyLogin Phone--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case config.WithEmail:
		if config.DefaultOtp != req.Otp {
			_, err := s.services.SmsService().ConfirmOtp(ctx, &sms_service.ConfirmOtpRequest{
				SmsId: req.GetSmsId(),
				Otp:   req.GetOtp(),
			})
			if err != nil {
				return nil, err
			}
		}

		user, err = s.strg.User().GetByUsername(ctx, req.GetEmail())
		if err != nil {
			s.log.Error("!!!MultiCompanyLogin Email--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	case config.WithGoogle:
		var email string
		if req.GetGoogleToken() != "" {
			userInfo, err := helper.GetGoogleUserInfo(req.GoogleToken)
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
			email = cast.ToString(userInfo["email"])
		} else {
			err := errors.New("google token is required when login type is google auth")
			s.log.Error("!!!MultiCompanyOneLogin--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		user, err = s.strg.User().GetByUsername(ctx, email)
		if err != nil {
			s.log.Error("!!!MultiCompanyOneLogin Email--->", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if user.Id == "" {
			err = errors.New("user not found with this email")
			s.log.Error("!!!MultiCompanyOneLogin--->", logger.Error(err))
			return nil, status.Error(codes.NotFound, err.Error())
		}
	}

	userProjects, err := s.strg.User().GetUserProjects(ctx, user.GetId())
	if err != nil {
		errGetProjects := errors.New("cant get user projects")
		s.log.Error("!!!MultiCompanyLogin--->GetUserProjects", logger.Error(err))
		return nil, status.Error(codes.NotFound, errGetProjects.Error())
	}

	userEnvProject, err := s.strg.User().GetUserEnvProjects(ctx, user.GetId())
	if err != nil {
		errGetEnvProjects := errors.New("cant get user env projects")
		s.log.Error("!!!MultiCompanyLogin--->GetUserEnvProjects", logger.Error(err))
		return nil, status.Error(codes.NotFound, errGetEnvProjects.Error())
	}

	for _, item := range userProjects.Companies {
		projects := make([]*pb.Project2, 0, 20)
		company, err := s.services.CompanyServiceClient().GetById(ctx, &pbCompany.GetCompanyByIdRequest{
			Id: item.Id,
		})
		if err != nil {
			errGetProjects := errors.New("cant get user projects")
			s.log.Error("!!!MultiCompanyLogin--->CompanyGetById", logger.Error(err))
			return nil, status.Error(codes.NotFound, errGetProjects.Error())
		}

		for _, projectId := range item.ProjectIds {
			clientType, _ := s.strg.User().GetUserProjectClientTypes(ctx,
				&models.UserProjectClientTypeRequest{UserId: user.GetId(), ProjectId: projectId})

			projectInfo, err := s.services.ProjectServiceClient().GetById(ctx, &pbCompany.GetProjectByIdRequest{
				ProjectId: projectId, CompanyId: item.Id,
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

			environments, err := s.services.EnvironmentService().GetList(ctx,
				&pbCompany.GetEnvironmentListRequest{
					Ids:       userEnvProject.EnvProjects[projectId],
					Limit:     1000,
					ProjectId: projectId,
				},
			)
			if err != nil {
				errGetProjects := errors.New("cant get environments")
				s.log.Error("!!!MultiCompanyLogin--->EnvironmentsList", logger.Error(err))
				return nil, status.Error(codes.NotFound, errGetProjects.Error())
			}

			for _, en := range environments.Environments {
				resourceEnv, err := s.services.ServiceResource().GetList(ctx,
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
					AccessType:    en.AccessType,
					ResourceId:    resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceId,
					ServiceType:   int32(resourceEnv.ServiceResources[config.ObjectBuilderService].ServiceType.Number()),
					Description:   en.Description,
					IsConfigured:  true,
					ResourceType:  int32(resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceType.Number()),
					DisplayColor:  en.DisplayColor,
					EnvironmentId: en.Id,
				}

				if resourceEnv.ServiceResources[config.ObjectBuilderService] == nil || resourceEnv.ServiceResources[config.ObjectBuilderService].ResourceEnvironmentId == "" {
					continue
				}

				if clientType == nil || len(clientType.ClientTypeIds) == 0 {
					clientTypes, err := s.services.ClientService().V2GetClientTypeList(ctx,
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
					clientTypes, err := s.services.ClientService().V2GetClientTypeList(ctx,
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
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.V2ResetPassword", req)
	defer dbSpan.Finish()

	s.log.Info("V2ResetPassword -> ", logger.Any("req: ", req))

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

	if req.GetPassword() != "" {
		if len(req.GetPassword()) < 6 {
			err := fmt.Errorf("password must not be less than 6 characters")
			s.log.Error("!!!ResetPassword-->PasswordCheck", logger.Error(err))
			return nil, err
		}

		hashedPassword, err := security.HashPasswordBcrypt(req.GetPassword())
		if err != nil {
			s.log.Error("!!!ResetPassword-->HashPassword", logger.Error(err))
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		req.Password = hashedPassword
	}
	rowsAffected, err := s.strg.User().V2ResetPassword(ctx, req)
	if err != nil {
		s.log.Error("!!!ResetPassword-->V2ResetPassword", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if rowsAffected <= 0 {
		return nil, status.Error(codes.InvalidArgument, "no rows were affected")
	}

	return s.strg.User().GetByPK(ctx, &pb.UserPrimaryKey{Id: req.GetUserId()})
}

func (s *sessionService) V2RefreshTokenForEnv(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.V2LoginResponse, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.V2RefreshTokenForEnv", req)
	defer dbSpan.Finish()

	var (
		res        = &pb.V2LoginResponse{}
		before     runtime.MemStats
		data       *pbObject.LoginDataRes
		roleId     string
		authTables = []*pb.TableBody{}
	)
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

	resource, err := s.services.ServiceResource().GetSingle(ctx,
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

	services, err := s.serviceNode.GetByNodeType(resource.ProjectId, resource.NodeType)
	if err != nil {
		return nil, err
	}

	reqLoginData := &pbObject.LoginDataReq{
		UserId:                user.GetId(),
		ClientType:            session.GetClientTypeId(),
		ProjectId:             session.GetProjectId(),
		ResourceEnvironmentId: resource.GetResourceEnvironmentId(),
	}

	switch resource.ResourceType {
	case 1:
		data, err = services.GetLoginServiceByType(resource.NodeType).LoginData(ctx, reqLoginData)
		if err != nil {
			errGetUserProjectData := errors.New("invalid user project data")
			s.log.Error("!!!V2RefreshTokenForEnv--->", logger.Error(err))
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

	if tokenInfo.Tables != nil {
		for _, table := range tokenInfo.Tables {
			authTables = append(authTables, &pb.TableBody{TableSlug: table.TableSlug, ObjectId: table.ObjectID})
		}
	}

	// TODO - wrap in a function
	m := map[string]interface{}{
		"id":                 session.Id,
		"ip":                 session.Data,
		"data":               session.Data,
		"tables":             authTables,
		"user_id":            session.UserId,
		"role_id":            session.RoleId,
		"project_id":         session.ProjectId,
		"user_id_auth":       session.GetUserIdAuth(),
		"client_type_id":     session.ClientTypeId,
		"login_table_slug":   tokenInfo.LoginTableSlug,
		"client_platform_id": session.ClientPlatformId,
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

	res.Token = token

	return res, nil
}

func (s *sessionService) ExpireSessions(ctx context.Context, req *pb.ExpireSessionsRequest) (*emptypb.Empty, error) {
	dbSpan, ctx := span.StartSpanFromContext(ctx, "grpc_session_v2.ExpireSessions", req)
	defer dbSpan.Finish()

	s.log.Info("---ExpireSessions--->>>", logger.Any("req", req.SessionIds))

	err := s.strg.Session().ExpireSessions(ctx, req)
	if err != nil {
		s.log.Error("!!!ExpireSessiona--->", logger.Error(err))
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}
