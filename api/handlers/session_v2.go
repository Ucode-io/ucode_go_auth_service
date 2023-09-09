package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	obs "ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// V2Login godoc
// @ID V2login
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/login [POST]
// @Summary V2Login
// @Description V2Login
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body auth_service.V2LoginRequest true "LoginRequestBody"
// @Success 201 {object} http.Response{data=string} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2Login(c *gin.Context) {
	var (
		login auth_service.V2LoginRequest
	)
	err := c.ShouldBindJSON(&login)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if login.ClientType == "" {
		h.handleResponse(c, http.BadRequest, "Необходимо выбрать тип пользователя")
		return
	}
	if login.ProjectId == "" {
		h.handleResponse(c, http.BadRequest, "Необходимо выбрать проекта")
		return
	}

	resourceId, ok := c.Get("resource_id")
	if !ok {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	resourceEnvironment, err := h.services.ResourceService().GetResourceEnvironment(
		c.Request.Context(),
		&obs.GetResourceEnvironmentReq{
			EnvironmentId: environmentId.(string),
			ResourceId:    resourceId.(string),
			ProjectId:     login.GetProjectId(),
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	login.ResourceEnvironmentId = resourceEnvironment.GetId()
	login.ResourceType = resourceEnvironment.GetResourceType()
	login.EnvironmentId = resourceEnvironment.GetEnvironmentId()

	resp, err := h.services.SessionService().V2Login(
		c.Request.Context(),
		&login,
	)
	httpErrorStr := ""
	if err != nil {
		httpErrorStr = strings.Split(err.Error(), "=")[len(strings.Split(err.Error(), "="))-1][1:]
		httpErrorStr = strings.ToLower(httpErrorStr)
	}
	if httpErrorStr == "user not found" {
		err := errors.New("пользователь не найдено")
		h.handleResponse(c, http.NotFound, err.Error())
		return
	} else if httpErrorStr == "user has been expired" {
		err := errors.New("срок действия пользователя истек")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if httpErrorStr == "invalid username" {
		err := errors.New("неверное имя пользователя")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if httpErrorStr == "invalid password" {
		err := errors.New("неверное пароль")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	resp.EnvironmentId = resourceEnvironment.GetEnvironmentId()
	resp.ResourceId = resourceEnvironment.GetResourceId()

	h.handleResponse(c, http.Created, resp)
}

// V2RefreshToken godoc
// @ID v2refresh
// @Router /v2/refresh [PUT]
// @Summary V2Refresh Token
// @Description V2Refresh Token
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param user body auth_service.RefreshTokenRequest true "RefreshTokenRequestBody"
// @Success 200 {object} http.Response{data=auth_service.V2RefreshTokenResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2RefreshToken(c *gin.Context) {
	var user auth_service.RefreshTokenRequest

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().V2RefreshToken(
		c.Request.Context(),
		&user,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2RefreshTokenSuperAdmin godoc
// @ID v2refresh_superadmin
// @Router /v2/refresh-superadmin [PUT]
// @Summary V2Refresh Token
// @Description V2Refresh Token
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param user body auth_service.RefreshTokenRequest true "RefreshTokenRequestBody"
// @Success 200 {object} http.Response{data=auth_service.V2RefreshTokenSuperAdminResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2RefreshTokenSuperAdmin(c *gin.Context) {
	var user auth_service.RefreshTokenRequest

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().V2RefreshTokenSuperAdmin(
		c.Request.Context(),
		&user,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2LoginSuperAdmin godoc
// @ID V2login_superadmin
// @Router /v2/login/superadmin [POST]
// @Summary V2LoginSuperAdmin
// @Description V2LoginSuperAdmin
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body auth_service.V2LoginSuperAdminReq true "V2LoginRequest"
// @Success 201 {object} http.Response{data=auth_service.V2LoginSuperAdminRes} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2LoginSuperAdmin(c *gin.Context) {
	var login auth_service.V2LoginRequest
	err := c.ShouldBindJSON(&login)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	//userReq, err := helper.ConvertMapToStruct(map[string]interface{}{
	//	"login":    login.Username,
	//	"password": login.Password,
	//})
	//if err != nil {
	//	h.handleResponse(c, http.BadRequest, err.Error())
	//	return
	//}
	//
	//userResp, err := h.services.ObjectBuilderService().GetList(
	//	context.Background(),
	//	&object_builder_service.CommonMessage{
	//		TableSlug: "user",
	//		Data:      userReq,
	//		ProjectId: config.UcodeDefaultProjectID,
	//	})
	//if err != nil {
	//	h.handleResponse(c, http.BadRequest, err.Error())
	//	return
	//}

	resp, err := h.services.SessionService().V2LoginSuperAdmin(
		c.Request.Context(),
		&auth_service.V2LoginSuperAdminReq{
			Username: login.GetUsername(),
			Password: login.GetPassword(),
		})

	httpErrorStr := ""
	if err != nil {
		httpErrorStr = strings.Split(err.Error(), "=")[len(strings.Split(err.Error(), "="))-1][1:]
		httpErrorStr = strings.ToLower(httpErrorStr)
	}
	if httpErrorStr == "user not found" {
		err := errors.New("Пользователь не найдено")
		h.handleResponse(c, http.NotFound, err.Error())
		return
	} else if httpErrorStr == "user has been expired" {
		err := errors.New("Срок действия пользователя истек")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if httpErrorStr == "invalid username" {
		err := errors.New("Неверное имя пользователя")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if httpErrorStr == "invalid password" {
		err := errors.New("Неверное пароль")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	companies, err := h.services.CompanyServiceClient().GetList(context.Background(), &obs.GetCompanyListRequest{
		Offset:  0,
		Limit:   128,
		OwnerId: resp.GetUserId(),
	})

	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	companiesResp := []*auth_service.Company{}

	if len(companies.Companies) < 1 {
		companiesById := make([]*obs.Company, 0)

		user, err := h.services.UserService().GetUserByID(c.Request.Context(), &auth_service.UserPrimaryKey{
			Id: resp.GetUserId(),
		})
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}

		company, err := h.services.CompanyServiceClient().GetById(c.Request.Context(), &obs.GetCompanyByIdRequest{
			Id: user.GetCompanyId(),
		})
		if err != nil {
			h.handleResponse(c, http.BadRequest, err.Error())
			return
		}

		companiesById = append(companiesById, company.Company)
		companies.Companies = companiesById
		companies.Count = 1

	}
	bytes, err := json.Marshal(companies.GetCompanies())
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	err = json.Unmarshal(bytes, &companiesResp)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	res := &auth_service.V2LoginSuperAdminRes{
		UserFound: resp.GetUserFound(),
		Token:     resp.GetToken(),
		Companies: companiesResp,
		UserId:    resp.GetUserId(),
		Sessions:  resp.GetSessions(),
	}

	h.handleResponse(c, http.Created, res)
}

// V2LoginWithOption godoc
// @ID V2login_withoption
// @Router /v2/login/with-option [POST]
// @Summary V2LoginWithOption
// @Description V2LoginWithOption
// @Description inside the data you must be passed client_type_id field
// @Description you must be give environment_id and project_id in body or
// @Description Environment-Id hearder and project-id in query parameters or
// @Description X-API-KEY in hearder
// @Description login strategy must be one of the following values
// @Description ["EMAIL", "PHONE", "EMAIL_OTP", "PHONE_OTP", "LOGIN", "LOGIN_PWD", "GOOGLE_AUTH", "APPLE_AUTH", "PHONE_PWD", "EMAIL_PWD"]
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param Environment-Id header string false "Environment-Id"
// @Param X-API-KEY header string false "X-API-KEY"
// @Param project-id query string false "project-id"
// @Param login body auth_service.V2LoginWithOptionRequest true "V2LoginRequest"
// @Success 201 {object} http.Response{data=auth_service.V2LoginSuperAdminRes} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2LoginWithOption(c *gin.Context) {
	var login auth_service.V2LoginWithOptionRequest
	err := c.ShouldBindJSON(&login)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	clientType := login.Data["client_type_id"]
	if clientType == "" {
		h.handleResponse(c, http.InvalidArgument, "inside data client_type_id is required")
		return
	}
	if ok := util.IsValidUUID(clientType); !ok {
		h.handleResponse(c, http.InvalidArgument, "lient_type_id is an invalid uuid")
		return
	}
	projectId, ok := c.Get("project_id")
	if !ok || !util.IsValidUUID(projectId.(string)) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		err = errors.New("error getting environment id | not valid")
		h.handleResponse(c, http.BadRequest, err)
		return
	}
	login.Data["environment_id"] = environmentId.(string)
	login.Data["project_id"] = projectId.(string)

	resp, err := h.services.SessionService().V2LoginWithOption(
		c.Request.Context(),
		&auth_service.V2LoginWithOptionRequest{
			Data:          login.GetData(),
			LoginStrategy: login.GetLoginStrategy(),
			Tables:        login.GetTables(),
		})

	httpErrorStr := ""
	if err != nil {
		httpErrorStr = strings.Split(err.Error(), "=")[len(strings.Split(err.Error(), "="))-1][1:]
		httpErrorStr = strings.ToLower(httpErrorStr)
	}
	if httpErrorStr == "user not found" {
		err := errors.New("Пользователь не найдено")
		h.handleResponse(c, http.NotFound, err.Error())
		return
	} else if httpErrorStr == "user verified but not found" {
		err := errors.New("Пользователь проверен, но не найден")
		h.handleResponse(c, http.OK, err.Error())
		return
	} else if httpErrorStr == "user has been expired" {
		err := errors.New("Срок действия пользователя истек")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if httpErrorStr == "invalid username" {
		err := errors.New("Неверное имя пользователя")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if httpErrorStr == "invalid password" {
		err := errors.New("Неверное пароль")
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	} else if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}
	res := &auth_service.V2LoginSuperAdminRes{
		UserFound: resp.GetUserFound(),
		Token:     resp.GetToken(),
		Companies: resp.GetCompanies(),
		UserId:    resp.GetUserId(),
		Sessions:  resp.GetSessions(),
		UserData:  resp.GetUserData(),
	}

	h.handleResponse(c, http.Created, res)
}

// MultiCompanyLogin godoc
// @ID multi_company_login
// @Router /v2/multi-company/login [POST]
// @Summary MultiCompanyLogin
// @Description MultiCompanyLogin
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body auth_service.MultiCompanyLoginRequest true "LoginRequestBody"
// @Success 201 {object} http.Response{data=auth_service.MultiCompanyLoginResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) MultiCompanyLogin(c *gin.Context) {
	var login auth_service.MultiCompanyLoginRequest

	err := c.ShouldBindJSON(&login)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().MultiCompanyLogin(
		c.Request.Context(),
		&login,
	)

	if err != nil {
		httpErrorStr := strings.Split(err.Error(), "=")[len(strings.Split(err.Error(), "="))-1][1:]
		httpErrorStr = strings.ToLower(httpErrorStr)

		if httpErrorStr == "user not found" {
			err := errors.New("Пользователь не найдено")
			h.handleResponse(c, http.NotFound, err.Error())
			return
		} else if httpErrorStr == "user has been expired" {
			err := errors.New("Срок действия пользователя истек")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if httpErrorStr == "invalid username" {
			err := errors.New("Неверное имя пользователя")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if httpErrorStr == "invalid password" {
			err := errors.New("Неверное пароль")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	h.handleResponse(c, http.Created, resp)
}

// V2MultiCompanyLogin godoc
// @ID multi_company_login_v2
// @Router /v2/v2multi-company/login [POST]
// @Summary MultiCompanyLogin
// @Description MultiCompanyLogin
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body auth_service.V2MultiCompanyLoginReq true "LoginRequestBody"
// @Success 201 {object} http.Response{data=auth_service.V2MultiCompanyLoginRes} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2MultiCompanyLogin(c *gin.Context) {
	var login auth_service.V2MultiCompanyLoginReq
	err := c.ShouldBindJSON(&login)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().V2MultiCompanyLogin(
		c.Request.Context(),
		&login,
	)

	if err != nil {
		httpErrorStr := strings.Split(err.Error(), "=")[len(strings.Split(err.Error(), "="))-1][1:]
		httpErrorStr = strings.ToLower(httpErrorStr)

		if httpErrorStr == "user not found" {
			err := errors.New("Пользователь не найдено")
			h.handleResponse(c, http.NotFound, err.Error())
			return
		} else if httpErrorStr == "user has been expired" {
			err := errors.New("Срок действия пользователя истек")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if httpErrorStr == "invalid username" {
			err := errors.New("Неверное имя пользователя")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if httpErrorStr == "invalid password" {
			err := errors.New("Неверное пароль")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	h.handleResponse(c, http.Created, resp)
}

// V2MultiCompanyOneLogin godoc
// @ID multi_company_one_login_v2
// @Router /v2/multi-company/one-login [POST]
// @Summary V2MultiCompanyOneLogin
// @Description V2MultiCompanyOneLogin
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body auth_service.V2MultiCompanyLoginReq true "LoginRequestBody"
// @Success 201 {object} http.Response{data=auth_service.V2MultiCompanyOneLoginRes} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2MultiCompanyOneLogin(c *gin.Context) {
	var login auth_service.V2MultiCompanyLoginReq

	err := c.ShouldBindJSON(&login)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().V2MultiCompanyOneLogin(
		c.Request.Context(),
		&login,
	)

	if err != nil {
		httpErrorStr := strings.Split(err.Error(), "=")[len(strings.Split(err.Error(), "="))-1][1:]
		httpErrorStr = strings.ToLower(httpErrorStr)

		if httpErrorStr == "user not found" {
			err := errors.New("Пользователь не найдено")
			h.handleResponse(c, http.NotFound, err.Error())
			return
		} else if httpErrorStr == "user has been expired" {
			err := errors.New("Срок действия пользователя истек")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if httpErrorStr == "invalid username" {
			err := errors.New("Неверное имя пользователя")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if httpErrorStr == "invalid password" {
			err := errors.New("Неверное пароль")
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		} else if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	h.handleResponse(c, http.Created, resp)
}

// V2ForgotPassword godoc
// @ID forgot_password
// @Router /v2/forgot-password [POST]
// @Summary ForgotPassword
// @Description Forgot Password
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body auth_service.ForgotPasswordRequest true "ForgotPasswordRequest"
// @Success 201 {object} http.Response{data=models.ForgotPasswordResponse} "Response"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) ForgotPassword(c *gin.Context) {
	var (
		request auth_service.ForgotPasswordRequest
	)

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*60)
	defer cancel()

	err := c.ShouldBindJSON(&request)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	user, err := h.services.UserService().GetUserByUsername(ctx, &auth_service.GetUserByUsernameRequest{
		Username: request.Login,
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}
	if user.GetEmail() == "" {
		h.handleResponse(c, http.OK, models.ForgotPasswordResponse{
			EmailFound: false,
			UserId:     user.GetId(),
			Email:      user.GetEmail(),
		})
		return
	}
	code, err := util.GenerateCode(6)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	expire := time.Now().Add(time.Hour * 5).Add(time.Minute * 5)

	id, err := uuid.NewRandom()
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	resp, err := h.services.EmailService().Create(
		c.Request.Context(),
		&auth_service.Email{
			Id:        id.String(),
			Email:     user.GetEmail(),
			Otp:       code,
			ExpiresAt: expire.String()[:19],
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	err = helper.SendCodeToEmail("Код для подтверждения", user.GetEmail(), code, h.cfg.Email, h.cfg.Password)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	h.handleResponse(c, http.OK, models.ForgotPasswordResponse{
		EmailFound: true,
		SmsId:      resp.GetId(),
		UserId:     user.GetId(),
		Email:      user.GetEmail(),
	})
}

// V2ForgotPassword godoc
// @ID set_email
// @Router /v2/set-email/send-code [PUT]
// @Summary SetEmail
// @Description Set Email
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body models.SetEmail true "SetEmailRequest"
// @Success 201 {object} http.Response{data=models.ForgotPasswordResponse} "Response"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) EmailEnter(c *gin.Context) {
	var (
		request models.SetEmail
	)

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*60)
	defer cancel()

	err := c.ShouldBindJSON(&request)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	res, err := h.services.SessionService().V2ResetPassword(ctx, &auth_service.V2ResetPasswordRequest{
		UserId: request.UserId,
		Email:  request.Email,
	})
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	code, err := util.GenerateCode(6)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	expire := time.Now().Add(time.Hour * 5).Add(time.Minute * 5)

	id, err := uuid.NewRandom()
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	resp, err := h.services.EmailService().Create(
		c.Request.Context(),
		&auth_service.Email{
			Id:        id.String(),
			Email:     res.GetEmail(),
			Otp:       code,
			ExpiresAt: expire.String()[:19],
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	err = helper.SendCodeToEmail("Код для подтверждения", res.GetEmail(), code, h.cfg.Email, h.cfg.Password)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	h.handleResponse(c, http.OK, models.ForgotPasswordResponse{
		EmailFound: true,
		SmsId:      resp.GetId(),
		UserId:     res.GetId(),
		Email:      res.GetEmail(),
	})
}

// V2ResetPassword godoc
// @ID v2_reset_password
// @Router /v2/reset-password [PUT]
// @Summary ResetPassword
// @Description Reset Password
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param body body models.ResetPassword true "ResetPasswordRequest"
// @Success 201 {object} http.Response{data=string} "Response"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2ResetPassword(c *gin.Context) {
	var (
		request models.ResetPassword
	)

	ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second*60)
	defer cancel()

	err := c.ShouldBindJSON(&request)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	res, err := h.services.SessionService().V2ResetPassword(ctx, &auth_service.V2ResetPasswordRequest{
		Password: request.Password,
		UserId:   request.UserId,
	})
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	h.handleResponse(c, http.OK, res)
}
