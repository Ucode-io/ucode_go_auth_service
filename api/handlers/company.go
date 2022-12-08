package handlers

import (
	"ucode/ucode_go_auth_service/api/http"

	"ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/saidamir98/udevs_pkg/util"

	"github.com/gin-gonic/gin"
)

// RegisterCompany godoc
// @ID create_company
// @Router /company [POST]
// @Summary Register Company
// @Description Register Company
// @Tags Company
// @Accept json
// @Produce json
// @Param company body auth_service.RegisterCompanyRequest true "RegisterCompanyRequestBody"
// @Success 201 {object} http.Response{data=string} "Company data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) RegisterCompany(c *gin.Context) {
	var company auth_service.RegisterCompanyRequest

	err := c.ShouldBindJSON(&company)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.CompanyService().Register(
		c.Request.Context(),
		&company,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// UpdateCompany godoc
// @ID update_company
// @Router /company [PUT]
// @Summary Update Company
// @Description Update Company
// @Tags Company
// @Accept json
// @Produce json
// @Param company body auth_service.UpdateCompanyRequest true "UpdateCompanyRequestBody"
// @Success 200 {object} http.Response{data=string} "Company data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateCompany(c *gin.Context) {
	var company auth_service.UpdateCompanyRequest

	err := c.ShouldBindJSON(&company)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.CompanyService().Update(
		c.Request.Context(),
		&company,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// RemoveCompany godoc
// @ID remove_company
// @Router /company/{company-id} [DELETE]
// @Summary Remove Company
// @Description Get Company
// @Tags Company
// @Accept json
// @Produce json
// @Param company-id path string true "company-id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) RemoveCompany(c *gin.Context) {
	companyID := c.Param("company-id")

	if !util.IsValidUUID(companyID) {
		h.handleResponse(c, http.InvalidArgument, "company id is an invalid uuid")
		return
	}

	resp, err := h.services.CompanyService().Remove(
		c.Request.Context(),
		&auth_service.CompanyPrimaryKey{
			Id: companyID,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}
