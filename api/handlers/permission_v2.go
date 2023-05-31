package handlers

import (
	"errors"
	"fmt"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"

	"github.com/saidamir98/udevs_pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/gin-gonic/gin"
)

// V2AddRole godoc
// @ID create_role_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role [POST]
// @Summary Create Role
// @Description Create Role
// @Tags V2_Role
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param role body auth_service.V2AddRoleRequest true "AddRoleRequestBody"
// @Success 201 {object} http.Response{data=auth_service.CommonMessage} "Role data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2AddRole(c *gin.Context) {
	var (
		// resourceEnvironment *obs.ResourceEnvironment
		role auth_service.V2AddRoleRequest
	)

	err := c.ShouldBindJSON(&role)
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
	//		&obs.GetResourceEnvironmentReq{
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
	//		&obs.GetDefaultResourceEnvironmentReq{
	//			ResourceId: resourceId.(string),
	//			ProjectId:  role.GetProjectId(),
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

	role.ProjectId = resource.ResourceEnvironmentId
	role.ResourceType = int32(resource.ResourceType)

	resp, err := h.services.PermissionService().V2AddRole(
		c.Request.Context(),
		&role,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// V2GetRoleByID godoc
// @ID get_role_by_id_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role/{role-id} [GET]
// @Summary Get Role By ID
// @Description Get Role By ID
// @Tags V2_Role
// @Accept json
// @Produce json
// @Param role-id path string true "role-id"
// @Param project-id query string false "project-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "ClientTypeBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetRoleByID(c *gin.Context) {
	var (
		// resourceEnvironment *obs.ResourceEnvironment
		err error
	)
	roleId := c.Param("role-id")
	if !util.IsValidUUID(roleId) {
		h.handleResponse(c, http.InvalidArgument, "role id is an invalid uuid")
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
	//		&obs.GetResourceEnvironmentReq{
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
	//		&obs.GetDefaultResourceEnvironmentReq{
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

	resp, err := h.services.PermissionService().V2GetRoleById(c.Request.Context(), &auth_service.V2RolePrimaryKey{
		Id:           roleId,
		ProjectId:    resource.ResourceEnvironmentId,
		ResourceType: int32(resource.ResourceType),
	})

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2GetRolesList godoc
// @ID get_roles_list_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role [GET]
// @Summary Get Roles List
// @Description  Get Roles List
// @Tags V2_Role
// @Accept json
// @Produce json
// @Param offset query integer false "offset"
// @Param limit query integer false "limit"
// @Param client-platform-id query string false "client-platform-id"
// @Param client-type-id query string false "client-type-id"
// @Param project-id query string false "project-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "GetRolesListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetRolesList(c *gin.Context) {
	var (
		// resourceEnvironment *obs.ResourceEnvironment
		err error
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
	//		&obs.GetResourceEnvironmentReq{
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
	//		&obs.GetDefaultResourceEnvironmentReq{
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

	resp, err := h.services.PermissionService().V2GetRolesList(
		c.Request.Context(),
		&auth_service.V2GetRolesListRequest{
			Offset:           uint32(offset),
			Limit:            uint32(limit),
			ClientPlatformId: c.Query("client-platform-id"),
			ClientTypeId:     c.Query("client-type-id"),
			ProjectId:        resource.ResourceEnvironmentId,
			ResourceType:     int32(resource.ResourceType),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2UpdateRole godoc
// @ID update_role_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role [PUT]
// @Summary Update Role
// @Description Update Role
// @Tags V2_Role
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param role body auth_service.V2UpdateRoleRequest true "UpdateRoleRequestBody"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "Role data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2UpdateRole(c *gin.Context) {
	var (
		// resourceEnvironment *obs.ResourceEnvironment
		role auth_service.V2UpdateRoleRequest
	)

	err := c.ShouldBindJSON(&role)
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
	//		&obs.GetResourceEnvironmentReq{
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
	//		&obs.GetDefaultResourceEnvironmentReq{
	//			ResourceId: resourceId.(string),
	//			ProjectId:  role.GetProjectId(),
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

	role.ProjectId = resource.ResourceEnvironmentId
	role.ResourceType = int32(resource.ResourceType)

	resp, err := h.services.PermissionService().V2UpdateRole(
		c.Request.Context(),
		&role,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2RemoveRole godoc
// @ID delete_role_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role/{role-id} [DELETE]
// @Summary Delete Role
// @Description Get Role
// @Tags V2_Role
// @Accept json
// @Produce json
// @Param role-id path string true "role-id"
// @Param project-id query string false "project-id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2RemoveRole(c *gin.Context) {
	var (
		// resourceEnvironment *obs.ResourceEnvironment
		err error
	)
	roleID := c.Param("role-id")

	if !util.IsValidUUID(roleID) {
		h.handleResponse(c, http.InvalidArgument, "role id is an invalid uuid")
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
	//		&obs.GetResourceEnvironmentReq{
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
	//		&obs.GetDefaultResourceEnvironmentReq{
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

	resp, err := h.services.PermissionService().V2RemoveRole(
		c.Request.Context(),
		&auth_service.V2RolePrimaryKey{
			Id:           roleID,
			ProjectId:    resource.ResourceEnvironmentId,
			ResourceType: int32(resource.ResourceType),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// V2CreatePermission godoc
// @ID create_permission_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/permission [POST]
// @Summary Create Permission
// @Description Create Permission
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param permission body auth_service.CreatePermissionRequest true "CreatePermissionRequestBody"
// @Success 201 {object} http.Response{data=auth_service.CommonMessage} "Permission data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2CreatePermission(c *gin.Context) {
	var permission auth_service.CreatePermissionRequest

	err := c.ShouldBindJSON(&permission)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.PermissionService().V2CreatePermission(
		c.Request.Context(),
		&permission,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// V2GetPermissionList godoc
// @ID get_permission_list_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/permission [GET]
// @Summary Get Permission List
// @Description  Get Permission List
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param offset query integer false "offset"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Param project-id query string true "project-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "GetPermissionListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetPermissionList(c *gin.Context) {
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

	resp, err := h.services.PermissionService().V2GetPermissionList(
		c.Request.Context(),
		&auth_service.GetPermissionListRequest{
			Limit:  int32(limit),
			Offset: int32(offset),
			Search: c.Query("search"),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2GetPermissionByID godoc
// @ID get_permission_by_id_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/permission/{permission-id} [GET]
// @Summary Get Permission By ID
// @Description Get Permission By ID
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param permission-id path string true "permission-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "PermissionBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetPermissionByID(c *gin.Context) {
	permissionID := c.Param("permission-id")

	if !util.IsValidUUID(permissionID) {
		h.handleResponse(c, http.InvalidArgument, "permission id is an invalid uuid")
		return
	}

	resp, err := h.services.PermissionService().V2GetPermissionByID(
		c.Request.Context(),
		&auth_service.PermissionPrimaryKey{
			Id: permissionID,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2UpdatePermission godoc
// @ID update_permission_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/permission [PUT]
// @Summary Update Permission
// @Description Update Permission
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param permission body auth_service.UpdatePermissionRequest true "UpdatePermissionRequestBody"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "Permission data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2UpdatePermission(c *gin.Context) {
	var permission auth_service.UpdatePermissionRequest

	err := c.ShouldBindJSON(&permission)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.PermissionService().V2UpdatePermission(
		c.Request.Context(),
		&permission,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2DeletePermission godoc
// @ID delete_permission_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/permission/{permission-id} [DELETE]
// @Summary Delete Permission
// @Description Get Permission
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param permission-id path string true "permission-id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2DeletePermission(c *gin.Context) {
	permissionID := c.Param("permission-id")

	if !util.IsValidUUID(permissionID) {
		h.handleResponse(c, http.InvalidArgument, "permission id is an invalid uuid")
		return
	}

	resp, err := h.services.PermissionService().V2DeletePermission(
		c.Request.Context(),
		&auth_service.PermissionPrimaryKey{
			Id: permissionID,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// V2GetScopesList godoc
// @ID get_scopes_list_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/scope [GET]
// @Summary Get Scopes List
// @Description  Get Scopes List
// @Tags V2_Scope
// @Accept json
// @Produce json
// @Param offset query integer false "offset"
// @Param limit query integer false "limit"
// @Param client-platform-id query string true "client-platform-id"
// @Param search query string false "search"
// @Param order_by query string false "order_by"
// @Param order_type query string false "order_type"
// @Param project-id query string true "project-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "GetScopesListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetScopesList(c *gin.Context) {
	clientPlatformID := c.Query("client-platform-id")
	if !util.IsValidUUID(clientPlatformID) {
		h.handleResponse(c, http.InvalidArgument, "Client platform id is an invalid uuid")
		return
	}

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

	resp, err := h.services.PermissionService().V2GetScopeList(
		c.Request.Context(),
		&auth_service.GetScopeListRequest{
			Offset:           uint32(offset),
			Limit:            uint32(limit),
			Search:           c.Query("search"),
			OrderBy:          c.Query("order_by"),
			OrderType:        c.Query("order_type"),
			ClientPlatformId: clientPlatformID,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2AddPermissionScope godoc
// @ID add_permission_scope_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/permission-scope [POST]
// @Summary Create PermissionScope
// @Description Create PermissionScope
// @Tags V2_PermissionScope
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param permission-scope body auth_service.AddPermissionScopeRequest true "AddPermissionScopeRequestBody"
// @Success 201 {object} http.Response{data=auth_service.CommonMessage} "PermissionScope data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2AddPermissionScope(c *gin.Context) {
	var permissionScope auth_service.AddPermissionScopeRequest

	err := c.ShouldBindJSON(&permissionScope)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.PermissionService().V2AddPermissionScope(
		c.Request.Context(),
		&permissionScope,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// V2RemovePermissionScope godoc
// @ID delete_permission_scope_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/permission-scope [DELETE]
// @Summary Delete PermissionScope
// @Description Get PermissionScope
// @Tags V2_PermissionScope
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param permission-scope body auth_service.PermissionScopePrimaryKey true "PermissionScopePrimaryKeyBody"
// @Success 204
// @Response 400 {object} http.Response{data=auth_service.CommonMessage} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2RemovePermissionScope(c *gin.Context) {
	var permissionScope auth_service.PermissionScopePrimaryKey

	err := c.ShouldBindJSON(&permissionScope)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.PermissionService().V2RemovePermissionScope(
		c.Request.Context(),
		&permissionScope,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// V2AddRolePermission godoc
// @ID add_role_permission_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role-permission [POST]
// @Summary Create RolePermission
// @Description Create RolePermission
// @Tags V2_RolePermission
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param role-permission body auth_service.AddRolePermissionRequest true "AddRolePermissionRequestBody"
// @Success 201 {object} http.Response{data=auth_service.CommonMessage} "RolePermission data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2AddRolePermission(c *gin.Context) {
	var rolePermission auth_service.AddRolePermissionRequest

	err := c.ShouldBindJSON(&rolePermission)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.PermissionService().V2AddRolePermission(
		c.Request.Context(),
		&rolePermission,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// V2RemoveRolePermission godoc
// @ID delete_role_permission_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role-permission [DELETE]
// @Summary Delete RolePermission
// @Description Get RolePermission
// @Tags V2_RolePermission
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param role-permission body auth_service.RolePermissionPrimaryKey true "RolePermissionPrimaryKeyBody"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2RemoveRolePermission(c *gin.Context) {
	var rolePermission auth_service.RolePermissionPrimaryKey

	err := c.ShouldBindJSON(&rolePermission)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.PermissionService().V2RemoveRolePermission(
		c.Request.Context(),
		&rolePermission,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// GetListWithRoleAppTablePermissions godoc
// @ID get_list_with_role_app_table_permissions
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role-permission/detailed/{project-id}/{role-id} [GET]
// @Summary Get Permission List
// @Description  Get Permission List
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param project-id path string false "project-id"
// @Param role-id path string false "role-id"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "GetPermissionListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetListWithRoleAppTablePermissions(c *gin.Context) {
	var (
		resp *object_builder_service.GetListWithRoleAppTablePermissionsResponse
	)
	// offset, err := h.getOffsetParam(c)
	// if err != nil {
	// 	h.handleResponse(c, http.InvalidArgument, err.Error())
	// 	return
	// }

	// limit, err := h.getLimitParam(c)
	// if err != nil {
	// 	h.handleResponse(c, http.InvalidArgument, err.Error())
	// 	return
	// }
	var (
		// resourceEnvironment *obs.ResourceEnvironment
		err error
	)

	projectId := c.Param("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.BadRequest, errors.New("not valid project id"))
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
	//		&obs.GetResourceEnvironmentReq{
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
	//		&obs.GetDefaultResourceEnvironmentReq{
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

	fmt.Println("\n Resource type ", resource.ResourceType)
	switch resource.ResourceType {
	case pbCompany.ResourceType_MONGODB:
		fmt.Println("\n Mongo db")
		resp, err = h.services.BuilderPermissionService().GetListWithRoleAppTablePermissions(
			c.Request.Context(),
			&object_builder_service.GetListWithRoleAppTablePermissionsRequest{
				RoleId:    c.Param("role-id"),
				ProjectId: resource.ResourceEnvironmentId,
			},
		)
		fmt.Println("\nResponse perm 123", resp, "\n\n")
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	case pbCompany.ResourceType_POSTGRESQL:
		resp, err = h.services.PostgresBuilderPermissionService().GetListWithRoleAppTablePermissions(
			c.Request.Context(),
			&object_builder_service.GetListWithRoleAppTablePermissionsRequest{
				RoleId:    c.Param("role-id"),
				ProjectId: resource.ResourceEnvironmentId,
			},
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}

	}

	resp.ProjectId = projectId

	h.handleResponse(c, http.OK, resp)
}

// UpdateRoleAppTablePermissions godoc
// @ID update_role_app_table_permissions
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role-permission/detailed [PUT]
// @Summary Update Permission
// @Description Update Permission
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param permission body object_builder_service.UpdateRoleAppTablePermissionsRequest true "UpdateRoleRequestBody"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "Role data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateRoleAppTablePermissions(c *gin.Context) {
	var (
		permission object_builder_service.UpdateRoleAppTablePermissionsRequest
		resp       *emptypb.Empty
		// resourceEnvironment *obs.ResourceEnvironment
	)

	err := c.ShouldBindJSON(&permission)
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
	//		&obs.GetResourceEnvironmentReq{
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
	//		&obs.GetDefaultResourceEnvironmentReq{
	//			ResourceId: resourceId.(string),
	//			ProjectId:  permission.GetProjectId(),
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

	permission.ProjectId = resource.ResourceEnvironmentId
	switch resource.ResourceType {
	case pbCompany.ResourceType_MONGODB:
		fmt.Println("test permission before update builder")
		resp, err = h.services.BuilderPermissionService().UpdateRoleAppTablePermissions(
			c.Request.Context(),
			&permission,
		)
		fmt.Println("test permission before error update builder")
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
		fmt.Println("test permission after update builder")
	case pbCompany.ResourceType_POSTGRESQL:
		resp, err = h.services.PostgresBuilderPermissionService().UpdateRoleAppTablePermissions(
			c.Request.Context(),
			&permission,
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	h.handleResponse(c, http.OK, resp)
}
