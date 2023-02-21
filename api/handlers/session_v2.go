package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	obs "ucode/ucode_go_auth_service/genproto/company_service"

	"github.com/gin-gonic/gin"
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
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	login.ResourceEnvironmentId = resourceEnvironment.GetId()

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
	fmt.Println("COMPANY::::1")
	companies, err := h.services.CompanyServiceClient().GetList(context.Background(), &obs.GetCompanyListRequest{
		Offset:  0,
		Limit:   128,
		OwnerId: resp.UserId,
	})
	fmt.Println("COMPANY::::2")
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	fmt.Println("COMPANY::::3")

	companiesResp := []*auth_service.Company{}

	if len(companies.Companies) < 1 {
		companiesById := make([]*obs.Company, 0)
		fmt.Println("COMPANY::::4")
		user, err := h.services.UserService().GetUserByID(context.Background(), &auth_service.UserPrimaryKey{
			Id: resp.UserId,
		})
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
		fmt.Println("COMPANY::::5")
		company, err := h.services.CompanyServiceClient().GetById(context.Background(), &obs.GetCompanyByIdRequest{
			Id: user.GetCompanyId(),
		})
		if err != nil {
			h.handleResponse(c, http.BadRequest, err.Error())
			return
		}
		fmt.Println("COMPANY::::6")
		companiesById = append(companiesById, company.Company)
		companies.Companies = companiesById
		companies.Count = 1

	}
	bytes, err := json.Marshal(companies.Companies)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	err = json.Unmarshal(bytes, &companiesResp)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	fmt.Println("COMPANY::::7")
	res := &auth_service.V2LoginSuperAdminRes{
		UserFound: resp.UserFound,
		Token:     resp.Token,
		Companies: companiesResp,
		UserId:    resp.UserId,
		Sessions:  resp.Sessions,
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
