package handlers

import (
	"errors"
	"fmt"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"

	"ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/saidamir98/udevs_pkg/util"
	"github.com/spf13/cast"

	"github.com/gin-gonic/gin"
)

// @Security ApiKeyAuth
// V2CreateClientType godoc
// @ID create_client_type_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/client-type [POST]
// @Summary Create ClientType
// @Description Create ClientType
// @Tags V2_ClientType
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param client-type body auth_service.V2CreateClientTypeRequest true "CreateClientTypeRequestBody"
// @Success 201 {object} http.Response{data=auth_service.CommonMessage} "ClientType data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2CreateClientType(c *gin.Context) {
	var (
		clientType auth_service.V2CreateClientTypeRequest
		resp       *auth_service.CommonMessage
		// resourceEnvironment *obs.ResourceEnvironment
	)

	err := c.ShouldBindJSON(&clientType)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
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
		fmt.Println("HERE AGAIN")
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	clientType.ProjectId = resource.ProjectId
	clientType.ResourceEnvrironmentId = resource.ResourceEnvironmentId
	clientType.ResourceType = int32(resource.ResourceType)
	clientType.NodeType = resource.NodeType

	userId, _ := c.Get("user_id")
	var (
		logReq = &models.CreateVersionHistoryRequest{
			NodeType:     resource.NodeType,
			ProjectId:    resource.ResourceEnvironmentId,
			ActionSource: c.Request.URL.String(),
			ActionType:   "CREATE CLIENT TYPE",
			UsedEnvironments: map[string]bool{
				cast.ToString(environmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   &clientType,
			TableSlug: "CLIENT_TYPE",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
			h.log.Info("!!!V2CreateClientType -> error")
		} else {
			logReq.Response = resp
			h.log.Info("V2CreateClientType -> success")
		}
		go h.versionHistory(c, logReq)
	}()

	resp, err = h.services.ClientService().V2CreateClientType(
		c.Request.Context(),
		&clientType,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// V2GetClientTypeList godoc
// @ID get_client_type_list_v2
// @Router /v2/client-type [GET]
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Summary Get ClientType List
// @Description  Get ClientType List
// @Tags V2_ClientType
// @Accept json
// @Produce json
// @Param offset query integer false "offset"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Param project-id query string false "project-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "GetClientTypeListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetClientTypeList(c *gin.Context) {
	var (
	// resourceEnvironment *obs.ResourceEnvironment
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

	resp, err := h.services.ClientService().V2GetClientTypeList(
		c.Request.Context(),
		&auth_service.V2GetClientTypeListRequest{
			Limit:                  int32(limit),
			Offset:                 int32(offset),
			Search:                 c.Query("search"),
			ProjectId:              resource.ProjectId,
			ResourceEnvrironmentId: resource.ResourceEnvironmentId,
			ResourceType:           int32(resource.ResourceType),
			NodeType:               resource.NodeType,
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2GetClientTypeByID godoc
// @ID get_client_type_by_id_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/client-type/{client-type-id} [GET]
// @Summary Get ClientType By ID
// @Description Get ClientType By ID
// @Tags V2_ClientType
// @Accept json
// @Produce json
// @Param client-type-id path string true "client-type-id"
// @Param project-id query string false "project-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "ClientTypeBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetClientTypeByID(c *gin.Context) {
	var (
		// resourceEnvironment *obs.ResourceEnvironment
		err error
	)
	clientTypeid := c.Param("client-type-id")

	if !util.IsValidUUID(clientTypeid) {
		h.handleResponse(c, http.InvalidArgument, "client_type id is an invalid uuid")
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

	resp, err := h.services.ClientService().V2GetClientTypeByID(
		c.Request.Context(),
		&auth_service.V2ClientTypePrimaryKey{
			Id:                     clientTypeid,
			ProjectId:              resource.ProjectId,
			ResourceEnvrironmentId: resource.ResourceEnvironmentId,
			ResourceType:           int32(resource.ResourceType),
			NodeType:               resource.NodeType,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// @Security ApiKeyAuth
// V2UpdateClientType godoc
// @ID update_client_type_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/client-type [PUT]
// @Summary Update ClientType
// @Description Update ClientType
// @Tags V2_ClientType
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param client-type body auth_service.V2UpdateClientTypeRequest true "UpdateClientTypeRequestBody"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "ClientType data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2UpdateClientType(c *gin.Context) {
	var (
		clientType auth_service.V2UpdateClientTypeRequest
		// resourceEnvironment *obs.ResourceEnvironment
		resp *auth_service.CommonMessage
	)

	err := c.ShouldBindJSON(&clientType)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
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

	clientType.ProjectId = resource.ProjectId
	clientType.ResourceEnvrironmentId = resource.ResourceEnvironmentId
	clientType.ResourceType = int32(resource.ResourceType)
	clientType.NodeType = resource.NodeType

	userId, _ := c.Get("user_id")
	var (
		logReq = &models.CreateVersionHistoryRequest{
			NodeType:     resource.NodeType,
			ProjectId:    resource.ResourceEnvironmentId,
			ActionSource: c.Request.URL.String(),
			ActionType:   "UPDATE CLIENT TYPE",
			UsedEnvironments: map[string]bool{
				cast.ToString(environmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   &clientType,
			TableSlug: "CLIENT_TYPE",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
			h.log.Info("!!!V2UpdateClientType -> error")
		} else {
			logReq.Response = resp
			h.log.Info("V2UpdateClientType -> success")
		}
		go h.versionHistory(c, logReq)
	}()

	resp, err = h.services.ClientService().V2UpdateClientType(
		c.Request.Context(),
		&clientType,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// @Security ApiKeyAuth
// V2DeleteClientType godoc
// @ID delete_client_type_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/client-type/{client-type-id} [DELETE]
// @Summary Delete ClientType
// @Description Get ClientType
// @Tags V2_ClientType
// @Accept json
// @Produce json
// @Param client-type-id path string true "client-type-id"
// @Param project-id query string false "project-id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2DeleteClientType(c *gin.Context) {
	clientTypeid := c.Param("client-type-id")
	var (
		// resourceEnvironment *obs.ResourceEnvironment
		err error
	)

	if !util.IsValidUUID(clientTypeid) {
		h.handleResponse(c, http.InvalidArgument, "client_type id is an invalid uuid")
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

	userId, _ := c.Get("user_id")
	var (
		logReq = &models.CreateVersionHistoryRequest{
			NodeType:     resource.NodeType,
			ProjectId:    resource.ResourceEnvironmentId,
			ActionSource: c.Request.URL.String(),
			ActionType:   "DELETE CLIENT TYPE",
			UsedEnvironments: map[string]bool{
				cast.ToString(environmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   clientTypeid,
			TableSlug: "CLIENT_TYPE",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
			h.log.Info("!!!V2DeleteClientType -> error")
		} else {
			logReq.Response = "success"
			h.log.Info("V2DeleteClientType -> success")
		}
		go h.versionHistory(c, logReq)
	}()

	resp, err := h.services.ClientService().V2DeleteClientType(
		c.Request.Context(),
		&auth_service.V2ClientTypePrimaryKey{
			Id:                     clientTypeid,
			ProjectId:              resource.ProjectId,
			ResourceEnvrironmentId: resource.ResourceEnvironmentId,
			ResourceType:           int32(resource.ResourceType),
			NodeType:               resource.NodeType,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}
