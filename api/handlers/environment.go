package handlers

import (
	"github.com/gin-gonic/gin"
	"ucode/ucode_go_auth_service/api/http"
	obs "ucode/ucode_go_auth_service/genproto/company_service"
)

// GetAllResourceEnvironments godoc
// @ID get_environments
// @Router /v2/resource-environment [GET]
// @Summary Get Environments
// @Description Get Environments
// @Tags V2_Environment
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Success 201 {object} http.Response{data=obs.GetListConfiguredResourceEnvironmentRes} "Environment data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetAllResourceEnvironments(c *gin.Context) {
	var (
		res *obs.GetListConfiguredResourceEnvironmentRes
	)
	resp, err := h.services.ResourceService().GetListConfiguredResourceEnvironment(
		c.Request.Context(),
		&obs.GetListConfiguredResourceEnvironmentReq{
			ProjectId: c.DefaultQuery("project-id", ""),
		},
	)

	for _, item := range resp.GetData() {
		if item.GetServiceType() == int32(obs.ServiceType_BUILDER_SERVICE.Number()) {
			res.Data = append(res.Data, item)
		}
	}

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}
