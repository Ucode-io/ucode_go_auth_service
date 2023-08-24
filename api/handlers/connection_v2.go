package handlers

import (
	"errors"
	"fmt"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"
	obs "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/gin-gonic/gin"
	"github.com/saidamir98/udevs_pkg/util"
)

// V2CreateConnection godoc
// @ID create_connection_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/connection [POST]
// @Summary Create Connection
// @Description Create Connection
// @Tags V2_Connection
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param connection body models.CreateConnectionRequest true "CreateConnectionRequestBody"
// @Success 201 {object} http.Response{data=auth_service.CommonMessage} "Connection data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2CreateConnection(c *gin.Context) {
	var (
		connection models.CreateConnectionRequest
		// resourceEnvironment *cps.ResourceEnvironment
		resp *obs.CommonMessage
	)

	err := c.ShouldBindJSON(&connection)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	projectId := c.Query("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	//resourceId, ok := c.Get("resource_id")
	//if !ok {
	//	h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
	//	return
	//}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	//if util.IsValidUUID(resourceId.(string)) {
	//	resourceEnvironment, err = h.services.ResourceService().GetResourceEnvironment(
	//		c.Request.Context(),
	//		&cps.GetResourceEnvironmentReq{
	//			EnvironmentId: environmentId.(string),
	//			ResourceId:    resourceId.(string),
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, http.GRPCError, err.Error())
	//		return
	//	}
	//} else {
	//	resourceEnvironment, err = h.services.ResourceService().GetDefaultResourceEnvironment(
	//		c.Request.Context(),
	//		&cps.GetDefaultResourceEnvironmentReq{
	//			ResourceId: resourceId.(string),
	//			ProjectId:  connection.ProjectId,
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, http.GRPCError, err.Error())
	//		return
	//	}
	//}

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

	structData, err := helper.ConvertRequestToSturct(connection)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}
	//this used for import auth_service proto
	var cm = auth_service.CommonMessage{}
	cm.Data = structData

	connection.ProjectId = resource.ResourceEnvironmentId

	// This is create connection by client type id
	switch resource.ResourceType {
	case 1:
		resp, err = h.services.ObjectBuilderService().Create(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				ProjectId: connection.ProjectId,
				Data:      structData,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	case 3:
		resp, err = h.services.PostgresObjectBuilderService().Create(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				ProjectId: connection.ProjectId,
				Data:      structData,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	h.handleResponse(c, http.Created, resp)
}

// V2UpdateConnection godoc
// @ID update_connection_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/connection [PUT]
// @Summary Update Connection
// @Description Update Connection
// @Tags V2_Connection
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param connection body models.Connection true "UpdateConnectionRequestBody"
// @Success 201 {object} http.Response{data=auth_service.CommonMessage} "Connection data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2UpdateConnection(c *gin.Context) {
	var (
		connection models.Connection
		// resourceEnvironment *cps.ResourceEnvironment
		resp *obs.CommonMessage
	)

	err := c.ShouldBindJSON(&connection)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	projectId := c.Query("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	//resourceId, ok := c.Get("resource_id")
	//if !ok {
	//	h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
	//	return
	//}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	//if util.IsValidUUID(resourceId.(string)) {
	//	resourceEnvironment, err = h.services.ResourceService().GetResourceEnvironment(
	//		c.Request.Context(),
	//		&cps.GetResourceEnvironmentReq{
	//			EnvironmentId: environmentId.(string),
	//			ResourceId:    resourceId.(string),
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, http.GRPCError, err.Error())
	//		return
	//	}
	//} else {
	//	resourceEnvironment, err = h.services.ResourceService().GetDefaultResourceEnvironment(
	//		c.Request.Context(),
	//		&cps.GetDefaultResourceEnvironmentReq{
	//			ResourceId: resourceId.(string),
	//			ProjectId:  connection.ProjectId,
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, http.GRPCError, err.Error())
	//		return
	//	}
	//}

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

	structData, err := helper.ConvertRequestToSturct(connection)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	connection.ProjectId = resource.ResourceEnvironmentId

	// This is create connection by client type id
	switch resource.ResourceType {
	case 1:
		resp, err = h.services.ObjectBuilderService().Update(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				ProjectId: connection.ProjectId,
				Data:      structData,
			},
		)
	case 3:
		resp, err = h.services.ObjectBuilderService().Update(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				ProjectId: connection.ProjectId,
				Data:      structData,
			},
		)
	}

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// V2GetConnectionList godoc
// @ID get_connection_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/connection [GET]
// @Summary Get Connection List
// @Description  Get Connection List
// @Tags V2_Connection
// @Accept json
// @Produce json
// @Param offset query integer false "offset"
// @Param limit query integer false "limit"
// @Param project-id query string false "project-id"
// @Param user-id query string true "user-id"
// @Param client_type_id query string true "client_type_id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "GetConnectionListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetConnectionList(c *gin.Context) {
	var (
	// resourceEnvironment *cps.ResourceEnvironment
	)
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

	structData, err := helper.ConvertMapToStruct(map[string]interface{}{
		"limit":          limit,
		"offset":         offset,
		"client_type_id": c.DefaultQuery("client_type_id", ""),
	})
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}
	// this is get list connection list from object builder
	var resp *obs.CommonMessage
	switch resource.ResourceType {
	case 1:
		resp, err = h.services.ObjectBuilderService().GetList(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				ProjectId: resource.ResourceEnvironmentId,
				Data:      structData,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	case 3:
		resp, err = h.services.PostgresObjectBuilderService().GetList(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				ProjectId: resource.ResourceEnvironmentId,
				Data:      structData,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}
	response, ok := resp.Data.AsMap()["response"].([]interface{})
	responseWithOptions := make([]interface{}, 0, len(response))
	if ok && c.Query("user-id") != "" {
		for _, v := range response {
			if res, ok := v.(map[string]interface{}); ok {
				if guid, ok := res["guid"].(string); ok {
					options, err := h.services.LoginService().GetConnetionOptions(
						c.Request.Context(),
						&obs.GetConnetionOptionsRequest{
							ConnectionId:          guid,
							ResourceEnvironmentId: resource.ResourceEnvironmentId,
							UserId:                c.Query("user-id"),
						},
					)
					if err != nil {
						continue
					}
					fmt.Println("options response::", options.Data.AsMap()["response"])
					res["options"] = options.Data.AsMap()["response"]
				}
				v = res
			}
			responseWithOptions = append(responseWithOptions, v)
		}
	}
	if len(responseWithOptions) < 0 {
		if res, ok := resp.Data.AsMap()["response"].([]interface{}); ok {
			responseWithOptions = res
		} else {
			responseWithOptions = []interface{}{}
		}
	}
	h.handleResponse(c, http.OK, map[string]interface{}{
		"data": map[string]interface{}{
			"response": responseWithOptions,
			"count":    resp.Data.AsMap()["count"],
		},
	},
	)
}

// V2GetConnectionByID godoc
// @ID get_connection_by_id_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/connection/{connection_id} [GET]
// @Summary Get Connection By ID
// @Description Get Connection By ID
// @Tags V2_Connection
// @Accept json
// @Produce json
// @Param connection_id path string true "connection_id"
// @Param project-id query string true "project-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "ConnectionBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetConnectionByID(c *gin.Context) {
	var (
		// resourceEnvironment *cps.ResourceEnvironment
		err error
	)
	connectionId := c.Param("connection_id")

	if !util.IsValidUUID(connectionId) {
		h.handleResponse(c, http.InvalidArgument, "connection id is an invalid uuid")
		return
	}

	projectId := c.Query("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	//resourceId, ok := c.Get("resource_id")
	//if !ok {
	//	err := errors.New("error getting resource id")
	//	h.handleResponse(c, http.BadRequest, err.Error())
	//	return
	//}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	//if util.IsValidUUID(resourceId.(string)) {
	//	resourceEnvironment, err = h.services.ResourceService().GetResourceEnvironment(
	//		c.Request.Context(),
	//		&cps.GetResourceEnvironmentReq{
	//			EnvironmentId: environmentId.(string),
	//			ResourceId:    resourceId.(string),
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, http.GRPCError, err.Error())
	//		return
	//	}
	//} else {
	//	resourceEnvironment, err = h.services.ResourceService().GetDefaultResourceEnvironment(
	//		c.Request.Context(),
	//		&cps.GetDefaultResourceEnvironmentReq{
	//			ResourceId: resourceId.(string),
	//			ProjectId:  projectId,
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, http.GRPCError, err.Error())
	//		return
	//	}
	//}

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

	structData, err := helper.ConvertMapToStruct(map[string]interface{}{
		"id": c.Param("connection_id"),
	})
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}
	var resp *obs.CommonMessage
	switch resource.ResourceType {
	case 1:
		resp, err = h.services.ObjectBuilderService().GetSingle(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				ProjectId: resource.ResourceEnvironmentId,
				Data:      structData,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	case 3:
		resp, err = h.services.PostgresObjectBuilderService().GetSingle(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				ProjectId: resource.ResourceEnvironmentId,
				Data:      structData,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	// this is get list connection list from object builder

	h.handleResponse(c, http.OK, resp)
}

// V2DeleteConnection godoc
// @ID delete_connection
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/connection/{connection_id} [DELETE]
// @Summary Delete Connection
// @Description Get Connection
// @Tags V2_Connection
// @Accept json
// @Produce json
// @Param connection_id path string true "connection_id"
// @Param project-id query string true "project-id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2DeleteConnection(c *gin.Context) {
	connectionId := c.Param("connection_id")
	var (
		// resourceEnvironment *cps.ResourceEnvironment
		err error
	)

	if !util.IsValidUUID(connectionId) {
		h.handleResponse(c, http.InvalidArgument, "connection id is an invalid uuid")
		return
	}

	projectId := c.Query("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	//resourceId, ok := c.Get("resource_id")
	//if !ok {
	//	err := errors.New("error getting resource id")
	//	h.handleResponse(c, http.BadRequest, err.Error())
	//	return
	//}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	//if util.IsValidUUID(resourceId.(string)) {
	//	resourceEnvironment, err = h.services.ResourceService().GetResourceEnvironment(
	//		c.Request.Context(),
	//		&cps.GetResourceEnvironmentReq{
	//			EnvironmentId: environmentId.(string),
	//			ResourceId:    resourceId.(string),
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, http.GRPCError, err.Error())
	//		return
	//	}
	//} else {
	//	resourceEnvironment, err = h.services.ResourceService().GetDefaultResourceEnvironment(
	//		c.Request.Context(),
	//		&cps.GetDefaultResourceEnvironmentReq{
	//			ResourceId: resourceId.(string),
	//			ProjectId:  projectId,
	//		},
	//	)
	//	if err != nil {
	//		h.handleResponse(c, http.GRPCError, err.Error())
	//		return
	//	}
	//}

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

	structData, err := helper.ConvertMapToStruct(map[string]interface{}{
		"id": connectionId,
	})
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}
	var resp *obs.CommonMessage
	switch resource.ResourceType {
	case 1:
		resp, err = h.services.ObjectBuilderService().Delete(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				Data:      structData,
				ProjectId: resource.ResourceEnvironmentId,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	case 3:
		resp, err = h.services.ObjectBuilderService().Delete(
			c.Request.Context(),
			&obs.CommonMessage{
				TableSlug: "connections",
				Data:      structData,
				ProjectId: resource.ResourceEnvironmentId,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	h.handleResponse(c, http.NoContent, resp)
}

// GetConnectionOptions godoc
// @ID get_connection_options
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/get-connection-options/{connection_id}/{user_id} [GET]
// @Summary Get Connection Options
// @Description Get Connection Options
// @Tags V2_Connection
// @Accept json
// @Produce json
// @Param connection_id path string true "connection_id"
// @Param user_id path string true "user_id"
// @Param project-id query string true "project-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "ConnectionOptionsBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetConnectionOptions(c *gin.Context) {
	var (
		// resourceEnvironment *cps.ResourceEnvironment
		err error
	)
	connectionId := c.Param("connection_id")

	if !util.IsValidUUID(connectionId) {
		h.handleResponse(c, http.InvalidArgument, "connection id is an invalid uuid")
		return
	}

	userId := c.Param("user_id")

	if !util.IsValidUUID(userId) {
		h.handleResponse(c, http.InvalidArgument, "user id is an invalid uuid")
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
	var resp *obs.GetConnectionOptionsResponse
	switch resource.ResourceType {
	case 1:
		resp, err = h.services.LoginService().GetConnetionOptions(
			c.Request.Context(),
			&obs.GetConnetionOptionsRequest{
				ConnectionId:          connectionId,
				ResourceEnvironmentId: resource.ResourceEnvironmentId,
				UserId:                userId,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	case 3:
		resp, err = h.services.PostgresLoginService().GetConnetionOptions(
			c.Request.Context(),
			&obs.GetConnetionOptionsRequest{
				ConnectionId:          connectionId,
				ResourceEnvironmentId: resource.ResourceEnvironmentId,
				UserId:                userId,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	// this is get list connection list from object builder

	h.handleResponse(c, http.OK, resp)
}
