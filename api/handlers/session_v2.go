package handlers

import (
	"context"
	"errors"
	"strings"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/pkg/helper"

	"ucode/ucode_go_auth_service/genproto/auth_service"
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
// @ID V2login
// @Router /v2/login/superadmin [POST]
// @Summary V2LoginSuperAdmin
// @Description V2LoginSuperAdmin
// @Tags V2_Session
// @Accept json
// @Produce json
// @Param login body auth_service.V2LoginSuperAdminRequest true "LoginSuperAdminRequestBody"
// @Success 201 {object} http.Response{data=string} "User data"
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

	userResp, err := h.services.ObjectBuilderService().GetSingle(
		context.Background(),
		&object_builder_service.CommonMessage{
			TableSlug: "user",
			Data:      userReq,
		})
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	userData := userResp.Data.AsMap()["data"].(map[string]interface{})
	userClientTypeID, ok := userData["client_type_id"].(string)
	if !ok {
		h.handleResponse(c, http.BadRequest, "Необходимо выбрать тип пользователя")
		return
	}

	login.ClientType = userClientTypeID

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
