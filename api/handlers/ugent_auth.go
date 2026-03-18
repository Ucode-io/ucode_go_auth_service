package handlers

import (
	"ucode/ucode_go_auth_service/api/http"
	pba "ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/gin-gonic/gin"
)

// UgenRegister godoc
// @ID create_company
// @Router /v3/ugen-register [POST]
// @Summary Register User
// @Description Register User
// @Tags Ugen
// @Accept json
// @Produce json
// @Param company body auth_service.RegisterCompanyRequest true "RegisterCompanyRequestBody"
// @Success 201 {object} http.Response{data=auth_service.CompanyPrimaryKey} "Company data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UgenRegister(c *gin.Context) {
	var company pba.RegisterCompanyRequest

	if err := c.ShouldBindJSON(&company); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	company.IsUgen = true

	resp, err := h.services.CompanyService().Register(
		c.Request.Context(), &company,
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

func (h *Handler) UgenLogin(c *gin.Context) {
	var login pba.UgenLoginReq

	if err := c.ShouldBindJSON(&login); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	login.ClientIp = c.ClientIP()
	login.UserAgent = c.Request.UserAgent()

	response, err := h.services.SessionService().UgenLogin(c.Request.Context(), &login)
	if err != nil {
		h.handleError(c, http.BadRequest, err)
		return
	}

	h.handleResponse(c, http.Created, response)
}
