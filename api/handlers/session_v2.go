package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/pkg/helper"

	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"

	"github.com/gin-gonic/gin"
)

// V2Login godoc
// @ID V2login
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
	var login auth_service.V2LoginRequest
	err := c.ShouldBindJSON(&login)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if login.ClientType == "" {
		h.handleResponse(c, http.BadRequest, "Необходимо выбрать тип пользователя")
		return
	}
	// if login.ProjectId == "" {
	// 	h.handleResponse(c, http.BadRequest, "Необходимо выбрать проекта")
	// 	return
	// }

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

// V2LoginSuperAdmin godoc
// @ID V2login_superadmin
// @Router /v2/login/superadmin [POST]
// @Summary V2LoginSuperAdmin
// @Description V2LoginSuperAdmin
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body auth_service.V2LoginRequest true "V2LoginRequest"
// @Success 201 {object} http.Response{data=auth_service.V2LoginResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2LoginSuperAdmin(c *gin.Context) {
	var login auth_service.V2LoginRequest
	err := c.ShouldBindJSON(&login)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	userReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"login":    login.Username,
		"password": login.Password,
	})
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	userResp, err := h.services.ObjectBuilderService().GetList(
		context.Background(),
		&object_builder_service.CommonMessage{
			TableSlug: "user",
			Data:      userReq,
			ProjectId: config.UcodeDefaultProjectID,
		})
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	userDatas, ok := userResp.Data.AsMap()["response"].([]interface{})
	if !ok {
		h.handleResponse(c, http.BadRequest, "Произошло ошибка")
		return
	}

	if len(userDatas) < 1 {
		h.handleResponse(c, http.BadRequest, "Пользователь не найдено")
		return
	} else if len(userDatas) > 1 {
		h.handleResponse(c, http.BadRequest, "Много пользователы найдено")
		return
	}

	userData, ok := userDatas[0].(map[string]interface{})
	if !ok {
		h.handleResponse(c, http.BadRequest, "Произошло ошибка")
		return
	}

	userClientTypeID, ok := userData["client_type_id"].(string)
	if !ok {
		h.handleResponse(c, http.BadRequest, "Необходимо выбрать тип пользователя")
		return
	}

	clientTypeReq, err := helper.ConvertMapToStruct(map[string]interface{}{
		"id": userClientTypeID,
	})
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	clientTypeResp, err := h.services.ObjectBuilderService().GetSingle(
		context.Background(),
		&object_builder_service.CommonMessage{
			TableSlug: "client_type",
			Data:      clientTypeReq,
			ProjectId: config.UcodeDefaultProjectID,
		})
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	clientTypeData, ok := clientTypeResp.Data.AsMap()["response"].(map[string]interface{})
	if !ok {
		h.handleResponse(c, http.BadRequest, "Произошло ошибка")
		return
	}

	clientTypeName, ok := clientTypeData["name"].(string)
	if !ok {
		h.handleResponse(c, http.BadRequest, "Произошло ошибка")
		return
	}

	login.ClientType = clientTypeName

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

	companies, err := h.services.CompanyServiceClient().GetList(context.Background(), &company_service.GetCompanyListRequest{
		Offset:  0,
		Limit:   128,
		OwnerId: resp.UserId,
	})
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	companiesResp := []*auth_service.Company{}

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

	res := &auth_service.V2SuperAdminLoginResponse{
		UserFound:      resp.UserFound,
		ClientPlatform: resp.ClientPlatform,
		ClientType:     resp.ClientType,
		Role:           resp.Role,
		Token:          resp.Token,
		Permissions:    resp.Permissions,
		Sessions:       resp.Sessions,
		LoginTableSlug: resp.LoginTableSlug,
		AppPermissions: resp.AppPermissions,
		Companies:      companiesResp,
	}

	h.handleResponse(c, http.Created, res)
}
