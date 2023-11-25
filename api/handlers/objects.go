package handlers

import (
	"errors"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"
	obs "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/gin-gonic/gin"
	"github.com/saidamir98/udevs_pkg/util"
)

// V2GetListObjects godoc
// @ID get_list_object
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/object/get-list/{table_slug} [POST]
// @Summary Get object list
// @Description  Get object list
// @Tags V2_Object
// @Accept json
// @Produce json
// @Param project-id query string false "project-id"
// @Param table_slug path string true "table_slug"
// @Param data body models.CommonMessage true "data"
// @Success 200 {object} http.Response{data=string} "GetObjectListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetListObjects(c *gin.Context) {
	var (
		//resourceEnvironment *cps.ResourceEnvironment
		body models.CommonMessage
		resp *obs.CommonMessage
	)

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	projectId := c.Query("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	resource, err := h.services.ServiceResource().GetSingle(
		c.Request.Context(),
		&pbCompany.GetSingleServiceResourceReq{
			ProjectId:     projectId,
			EnvironmentId: environmentId.(string),
			ServiceType:   pbCompany.ServiceType_BUILDER_SERVICE,
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	structData, err := helper.ConvertMapToStruct(body.Data)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	services, err := h.GetProjectSrvc(
		c,
		resource.ProjectId,
		resource.NodeType,
	)

	// this is get list objects list from object builder
	switch resource.ResourceType {
	case 1:
		resp, err = services.GetObjectBuilderServiceByType(resource.NodeType).GetList(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: c.Param("table_slug"),
				ProjectId: resource.ResourceEnvironmentId,
				Data:      structData,
			},
		)
	case 3:
		resp, err = services.PostgresObjectBuilderService().GetList(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: c.Param("table_slug"),
				ProjectId: resource.ResourceEnvironmentId,
				Data:      structData,
			},
		)
	}

	h.handleResponse(c, http.OK, resp)
}
