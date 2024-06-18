package handlers

import (
	"context"
	"errors"
	"fmt"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/genproto/new_object_builder_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/saidamir98/udevs_pkg/util"
	"github.com/spf13/cast"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/gin-gonic/gin"
)

// GetGlobalPermissionByRoleId godoc
// @ID get_global_permission_by_role_id
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/role-golabal-permission/{project-id}/{role-id} [GET]
// @Summary Get Global Role By ID
// @Description Get Global Role By ID
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param role-id path string true "role-id"
// @Param project-id query string false "project-id"
// @Success 200 {object} http.Response{data=object_builder_service.GlobalPermission} "ClientTypeBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetGlobalPermission(c *gin.Context) {
	var (
		resp *object_builder_service.GlobalPermission
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

	// environmentId, ok := c.Get("environment_id")
	// if !ok || !util.IsValidUUID(environmentId.(string)) {
	// 	h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
	// 	return
	// }
	// fmt.Println(">>>>>>>>>>>>>>>   test #0.3")
	// resource, err := h.services.ServiceResource().GetSingle(
	// 	c.Request.Context(),
	// 	&pbCompany.GetSingleServiceResourceReq{
	// 		ProjectId:     projectId,
	// 		EnvironmentId: environmentId.(string),
	// 		ServiceType:   pbCompany.ServiceType_BUILDER_SERVICE,
	// 	},
	// )
	// fmt.Println(">>>>>>>>>>>>>>>   test #1")
	// if err != nil {
	// 	h.handleResponse(c, http.GRPCError, err.Error())
	// 	return
	// }
	// fmt.Println(">>>>>>>>>>>>>>>   test #2")
	// switch resource.ResourceType {
	// case pbCompany.ResourceType_MONGODB:
	// resp, err = h.services.BuilderPermissionService().GetGlobalPermissionByRoleId(
	// 	c.Request.Context(),
	// 	&object_builder_service.GetGlobalPermissionsByRoleIdRequest{
	// 		RoleId:    roleId,
	// 		ProjectId: "1",
	// 	},
	// )

	// if err != nil {
	// 	h.handleResponse(c, http.GRPCError, err.Error())
	// 	return
	// }
	// case pbCompany.ResourceType_POSTGRESQL:
	// resp, err = h.services.PostgresBuilderPermissionService().GetGlobalPermissionByRoleId(
	// 	c.Request.Context(),
	// 	&object_builder_service.GetListWithRoleAppTablePermissionsRequest{
	// 		RoleId:    c.Param("role-id"),
	// 		ProjectId: resource.ResourceEnvironmentId,
	// 	},
	// )

	// if err != nil {
	// 	h.handleResponse(c, http.GRPCError, err.Error())
	// 	return
	// }
	// }

	h.handleResponse(c, http.OK, resp)
}

// @Security ApiKeyAuth
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
		resp *auth_service.CommonMessage
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

	role.ProjectId = resource.ResourceEnvironmentId
	role.ResourceType = int32(resource.ResourceType)
	role.NodeType = resource.NodeType

	userId, _ := c.Get("user_id")
	var (
		logReq = &models.CreateVersionHistoryRequest{
			NodeType:     resource.NodeType,
			ProjectId:    resource.ResourceEnvironmentId,
			ActionSource: c.Request.URL.String(),
			ActionType:   "CREATE ROLE",
			UsedEnvironments: map[string]bool{
				cast.ToString(environmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   &role,
			TableSlug: "ROLE",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
			h.log.Info("!!!V2AddRole -> error")
		} else {
			logReq.Response = resp
			h.log.Info("V2AddRole -> success")
		}
		// go h.versionHistory(c, logReq)
	}()

	resp, err = h.services.PermissionService().V2AddRole(
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

	resp, err := h.services.PermissionService().V2GetRoleById(c.Request.Context(), &auth_service.V2RolePrimaryKey{
		Id:           roleId,
		ProjectId:    resource.ResourceEnvironmentId,
		ResourceType: int32(resource.ResourceType),
		NodeType:     resource.NodeType,
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

	var clientTypeId string
	if c.Query("clienty_type_id") != "" {
		clientTypeId = c.Query("clienty_type_id")
	} else {
		clientTypeId = c.Query("client-type-id")
	}
	resp, err := h.services.PermissionService().V2GetRolesList(
		c.Request.Context(),
		&auth_service.V2GetRolesListRequest{
			Offset:                uint32(offset),
			Limit:                 uint32(limit),
			ClientPlatformId:      c.Query("client-platform-id"),
			ClientTypeId:          clientTypeId,
			ProjectId:             resource.ProjectId,
			ResourceEnvironmentId: resource.ResourceEnvironmentId,
			ResourceType:          int32(resource.ResourceType),
			NodeType:              resource.NodeType,
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
	role.NodeType = resource.NodeType

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

// @Security ApiKeyAuth
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
		err  error
		resp *auth_service.CommonMessage
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

	userId, _ := c.Get("user_id")
	var (
		logReq = &models.CreateVersionHistoryRequest{
			NodeType:     resource.NodeType,
			ProjectId:    resource.ResourceEnvironmentId,
			ActionSource: c.Request.URL.String(),
			ActionType:   "DELETE ROLE",
			UsedEnvironments: map[string]bool{
				cast.ToString(environmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   roleID,
			TableSlug: "ROLE",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
			h.log.Info("!!!V2RemoveRole -> error")
		} else {
			logReq.Response = resp
			h.log.Info("V2RemoveRole -> success")
		}

	}()

	fmt.Printf("roleID: %+v\n", resource)

	resp, err = h.services.PermissionService().V2RemoveRole(
		c.Request.Context(),
		&auth_service.V2RolePrimaryKey{
			Id:                    roleID,
			ProjectId:             resource.ResourceEnvironmentId,
			ResourceType:          int32(resource.ResourceType),
			NodeType:              resource.NodeType,
			ResourceEnvironmentId: resource.ResourceEnvironmentId,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// @Security ApiKeyAuth
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
// @Success 200 {object} http.Response{data=object_builder_service.GetListWithRoleAppTablePermissionsResponse} "GetPermissionListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetListWithRoleAppTablePermissions(c *gin.Context) {
	var (
		resp *object_builder_service.GetListWithRoleAppTablePermissionsResponse
		err  error
	)

	projectId := c.Param("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.BadRequest, errors.New("not valid project id"))
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

	services, _ := h.GetProjectSrvc(
		c,
		resource.ProjectId,
		resource.NodeType,
	)

	switch resource.ResourceType {
	case pbCompany.ResourceType_MONGODB:

		resp, err = services.GetBuilderPermissionServiceByType(resource.NodeType).GetListWithRoleAppTablePermissions(
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
		resp.ProjectId = projectId

		h.handleResponse(c, http.OK, resp)
	case pbCompany.ResourceType_POSTGRESQL:
		resp, err := services.GoObjectBuilderPermissionService().GetListWithRoleAppTablePermissions(context.Background(),
			&new_object_builder_service.GetListWithRoleAppTablePermissionsRequest{
				ProjectId: resource.ResourceEnvironmentId,
				RoleId:    c.Param("role-id"),
			})

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}

		resp.ProjectId = projectId

		h.handleResponse(c, http.OK, resp)

	}
}

// @Security ApiKeyAuth
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
			ActionType:   "UPDATE PERMISSION",
			UsedEnvironments: map[string]bool{
				cast.ToString(environmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   &permission,
			TableSlug: "ROLE_PERMISSION",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
			h.log.Info("!!!UpdateRoleAppTablePermissions -> error")
		} else {
			logReq.Response = resp
			h.log.Info("UpdateRoleAppTablePermissions -> success")
		}
		go h.versionHistory(c, logReq)
	}()

	services, _ := h.GetProjectSrvc(
		c,
		resource.ProjectId,
		resource.NodeType,
	)

	permission.ProjectId = resource.ResourceEnvironmentId
	switch resource.ResourceType {
	case pbCompany.ResourceType_MONGODB:

		resp, err = services.GetBuilderPermissionServiceByType(resource.NodeType).UpdateRoleAppTablePermissions(
			context.Background(),
			&permission,
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}

		h.handleResponse(c, http.OK, resp)
	case pbCompany.ResourceType_POSTGRESQL:

		newPermission := &new_object_builder_service.UpdateRoleAppTablePermissionsRequest{}

		err := helper.MarshalToStruct(&permission, &newPermission)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}

		resp, err := services.GoObjectBuilderPermissionService().UpdateRoleAppTablePermissions(
			c.Request.Context(),
			newPermission,
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}

		h.handleResponse(c, http.OK, resp)
	}

}

// @Security ApiKeyAuth
// GetListMenuPermissions godoc
// @ID get_list_menu_permissions
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/menu-permission/detailed/{project-id}/{role-id}/{parent-id} [GET]
// @Summary Get Permission List
// @Description  Get Permission List
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param project-id path string true "project-id"
// @Param role-id path string true "role-id"
// @Param parent-id path string true "parent-id"
// @Success 200 {object} http.Response{data=object_builder_service.GetAllMenuPermissionsResponse} "GetMenuPermissionListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetListMenuPermissions(c *gin.Context) {
	var (
		resp *object_builder_service.GetAllMenuPermissionsResponse
		err  error
	)

	projectId := c.Param("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.BadRequest, errors.New("not valid project id"))
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

	services, _ := h.GetProjectSrvc(
		c,
		resource.ProjectId,
		resource.NodeType,
	)

	switch resource.ResourceType {
	case pbCompany.ResourceType_MONGODB:

		resp, err = services.GetBuilderPermissionServiceByType(resource.NodeType).GetAllMenuPermissions(
			c.Request.Context(),
			&object_builder_service.GetAllMenuPermissionsRequest{
				RoleId:    c.Param("role-id"),
				ProjectId: resource.ResourceEnvironmentId,
				ParentId:  c.Param("parent-id"),
			},
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	case pbCompany.ResourceType_POSTGRESQL:
		resp, err := services.GoObjectBuilderPermissionService().GetAllMenuPermissions(
			c.Request.Context(),
			&new_object_builder_service.GetAllMenuPermissionsRequest{
				RoleId:    c.Param("role-id"),
				ProjectId: resource.ResourceEnvironmentId,
				ParentId:  c.Param("parent-id"),
			},
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
		h.handleResponse(c, http.OK, resp)
		return
	}
	h.handleResponse(c, http.OK, resp)
}

// @Security ApiKeyAuth
// UpdateMenuPermissions godoc
// @ID update_menu_permissions
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/menu-permission/detailed [PUT]
// @Summary Update Permission
// @Description Update Permission
// @Tags V2_Permission
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Param permission body object_builder_service.UpdateMenuPermissionsRequest true "UpdateMenuPermissionRequestBody"
// @Success 200 {object} http.Response{data=auth_service.CommonMessage} "Role data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateMenuPermissions(c *gin.Context) {
	var (
		permission object_builder_service.UpdateMenuPermissionsRequest
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
			ActionType:   "UPDATE PERMISSION",
			UsedEnvironments: map[string]bool{
				cast.ToString(environmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   &permission,
			TableSlug: "MENU_PERMISSION",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
		} else {
			logReq.Response = resp
		}
		go h.versionHistory(c, logReq)
	}()

	services, _ := h.GetProjectSrvc(
		c,
		resource.ProjectId,
		resource.NodeType,
	)

	permission.ProjectId = resource.ResourceEnvironmentId
	switch resource.ResourceType {
	case pbCompany.ResourceType_MONGODB:

		resp, err = services.GetBuilderPermissionServiceByType(resource.NodeType).UpdateMenuPermissions(
			context.Background(),
			&permission,
		)

		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}

	case pbCompany.ResourceType_POSTGRESQL:
		newReq := new_object_builder_service.UpdateMenuPermissionsRequest{}
		err = helper.MarshalToStruct(&permission, &newReq)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}

		resp, err = services.GoObjectBuilderPermissionService().UpdateMenuPermissions(
			c.Request.Context(),
			&newReq,
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	h.handleResponse(c, http.OK, resp)
}
