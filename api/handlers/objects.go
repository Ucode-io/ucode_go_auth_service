package handlers

import (
	"errors"
	"fmt"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	cps "ucode/ucode_go_auth_service/genproto/company_service"
	obs "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
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
// @Param project_id query string false "project_id"
// @Param table_slug path string true "table_slug"
// @Param data body models.CommonMessage true "data"
// @Success 200 {object} http.Response{data=string} "GetObjectListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetListObjects(c *gin.Context) {
	var (
		resourceEnvironment *cps.ResourceEnvironment
		body                models.CommonMessage
	)

	projectId := c.DefaultQuery("project_id", "")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get project_id"))
		return
	}
	err := c.ShouldBindJSON(&body)

	resourceId, ok := c.Get("resource_id")
	if !ok {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	if util.IsValidUUID(resourceId.(string)) {
		resourceEnvironment, err = h.services.ResourceService().GetResourceEnvironment(
			c.Request.Context(),
			&cps.GetResourceEnvironmentReq{
				EnvironmentId: environmentId.(string),
				ResourceId:    resourceId.(string),
			},
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	} else {
		resourceEnvironment, err = h.services.ResourceService().GetDefaultResourceEnvironment(
			c.Request.Context(),
			&cps.GetDefaultResourceEnvironmentReq{
				ResourceId: resourceId.(string),
				ProjectId:  projectId,
			},
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				h.handleResponse(c, http.GRPCError, "У вас нет ресурса по умолчанию, установите один ресурс по умолчанию")
				return
			}
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}
	structData, err := helper.ConvertMapToStruct(body.Data)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	fmt.Println("test")
	// this is get list connection list from object builder
	resp, err := h.services.ObjectBuilderService().GetList(
		c.Request.Context(),
		&obs.CommonMessage{
			TableSlug: c.Param("table_slug"),
			ProjectId: resourceEnvironment.GetId(),
			Data:      structData,
		},
	)
	fmt.Println("ress:::", resp.Data.AsMap()["response"])

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}
