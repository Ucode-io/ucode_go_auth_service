package handlers

import (
	"ucode/ucode_go_auth_service/api/http"

	"ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/gin-gonic/gin"
)

// Login godoc
// @ID login
// @Router /login [POST]
// @Summary Login
// @Description Login
// @Tags Session
// @Accept json
// @Produce json
// @Param login body auth_service.LoginRequest true "LoginRequestBody"
// @Success 201 {object} http.Response{data=models.LoginResponse} "Login data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) Login(c *gin.Context) {
	var login auth_service.LoginRequest

	if err := c.ShouldBindJSON(&login); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().Login(
		c.Request.Context(), &login,
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// Logout godoc
// @ID logout
// @Router /logout [DELETE]
// @Summary Logout User
// @Description Logout User
// @Tags Session
// @Accept json
// @Produce json
// @Param data body auth_service.LogoutRequest true "LogoutRequest"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) Logout(c *gin.Context) {
	var logout auth_service.LogoutRequest

	err := c.ShouldBindJSON(&logout)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().Logout(
		c.Request.Context(),
		&logout,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// RefreshToken godoc
// @ID refresh
// @Router /refresh [PUT]
// @Summary Refresh Token
// @Description Refresh Token
// @Tags Session
// @Accept json
// @Produce json
// @Param user body auth_service.RefreshTokenRequest true "RefreshTokenRequestBody"
// @Success 200 {object} http.Response{data=auth_service.RefreshTokenResponse} "Refresh token data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) RefreshToken(c *gin.Context) {
	var user auth_service.RefreshTokenRequest

	if err := c.ShouldBindJSON(&user); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().RefreshToken(
		c.Request.Context(), &user,
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// HasAccess godoc
// @ID has_access
// @Router /has-access [POST]
// @Summary Has Access
// @Description Has Access
// @Tags Session
// @Accept json
// @Produce json
// @Param has-access body auth_service.HasAccessRequest true "HasAccessRequestBody"
// @Success 201 {object} http.Response{data=auth_service.HasAccessResponse} "User access data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) HasAccess(c *gin.Context) {
	var login auth_service.HasAccessRequest

	if err := c.ShouldBindJSON(&login); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().HasAccess(
		c.Request.Context(), &login,
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// HasAccessSuperAdmin godoc
// @ID has_access_superadmin
// @Router /has-access-super-admin [POST]
// @Summary Has Access
// @Description Has Access
// @Tags Session
// @Accept json
// @Produce json
// @Param has-access body auth_service.HasAccessSuperAdminReq true "HasAccessRequestBody"
// @Success 201 {object} http.Response{data=auth_service.HasAccessSuperAdminRes} "Admin access data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) HasAccessSuperAdmin(c *gin.Context) {
	var login auth_service.HasAccessSuperAdminReq

	if err := c.ShouldBindJSON(&login); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.SessionService().HasAccessSuperAdmin(
		c.Request.Context(), &login,
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

func (h *Handler) GetSessionList(c *gin.Context) {
	offset, err := h.getOffsetParam(c)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	limit, err := h.getLimitParam(c)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	resp, err := h.services.SessionService().GetList(
		c.Request.Context(),
		&auth_service.GetSessionListRequest{
			Limit:        int32(limit),
			Offset:       int32(offset),
			Search:       c.Query("search"),
			UserId:       c.Query("user_id"),
			ClientTypeId: c.Query("client_type_id"),
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

func (h *Handler) DeleteSession(c *gin.Context) {
	resp, err := h.services.SessionService().Delete(
		c.Request.Context(),
		&auth_service.SessionPrimaryKey{
			Id: c.Param("id"),
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}
