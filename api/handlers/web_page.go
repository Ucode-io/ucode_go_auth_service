package handlers

import (
	"context"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/genproto/web_page_service"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
)

// GetListWebPageApp godoc
// @ID get_list_web_page_app
// @Router /v2/webpage-app [GET]
// @Summary Get List webpage app
// @Description Get List webpage app
// @Tags WebPage
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param environment-id query string true "environment-id"
// @Success 200 {object} http.Response{data=web_page_service.GetListAppRes} "AppBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetListWebPageApp(c *gin.Context) {
	var (
	//resourceEnvironment *obs.ResourceEnvironment
	)

	//namespace := c.GetString("namespace")
	//services, err := h.GetService(namespace)
	//if err != nil {
	//	h.handleResponse(c, status_http.Forbidden, err)
	//	return
	//}

	//authInfo, err := h.GetAuthInfo(c)
	//if err != nil {
	//	h.handleResponse(c, status_http.Forbidden, err.Error())
	//	return
	//}

	projectId := c.Query("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	//resourceId, ok := c.Get("resource_id")
	//if !ok {
	//	err = errors.New("error getting resource id")
	//	h.handleResponse(c, status_http.BadRequest, err.Error())
	//	return
	//}
	//
	envId := c.Query("environment-id")
	if !util.IsValidUUID(envId) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	resource, err := h.services.MicroServiceResourceService().GetSingle(
		c.Request.Context(),
		&company_service.GetSingleServiceResourceReq{
			ProjectId:     projectId,
			EnvironmentId: envId,
			ServiceType:   company_service.ServiceType_WEB_PAGE_SERVICE,
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}
	//
	//if util.IsValidUUID(environmentId.(string)) {
	//	resourceEnvironment, err = services.CompanyService().Resource().GetResourceEnvironment(
	//		c.Request.Context(),
	//		&obs.GetResourceEnvironmentReq{
	//			EnvironmentId: environmentId.(string),
	//			ResourceId:    resourceId.(string),
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, status_http.GRPCError, err.Error())
	//		return
	//	}
	//} else {
	//	resourceEnvironment, err = services.CompanyService().Resource().GetDefaultResourceEnvironment(
	//		c.Request.Context(),
	//		&obs.GetDefaultResourceEnvironmentReq{
	//			ResourceId: resourceId.(string),
	//			ProjectId:  projectId,
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, status_http.GRPCError, err.Error())
	//		return
	//	}
	//}

	res, err := h.services.WebPageAppService().GetListApp(
		context.Background(),
		&web_page_service.GetListAppReq{
			ProjectId:     resource.GetProjectId(),
			EnvironmentId: envId,
			ResourceId:    resource.GetResourceEnvironmentId(),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}
