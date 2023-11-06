package handlers

import (
	"context"
	"log"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"

	"ucode/ucode_go_auth_service/genproto/auth_service"
	pb "ucode/ucode_go_auth_service/genproto/company_service"
	obs "ucode/ucode_go_auth_service/genproto/object_builder_service"

	"github.com/saidamir98/udevs_pkg/util"

	"github.com/gin-gonic/gin"
)

// V2CreateUser godoc
// @ID create_user_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/user [POST]
// @Summary Create User
// @Description Create User
// @Tags V2_User
// @Accept json
// @Produce json
// @Param user body auth_service.CreateUserRequest true "CreateUserRequestBody"
// @Success 201 {object} http.Response{data=auth_service.User} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2CreateUser(c *gin.Context) {
	var (
		// resourceEnvironment *company_service.ResourceEnvironment
		user auth_service.CreateUserRequest
	)

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}
	if !util.IsValidUUID(user.ProjectId) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get project_id"))
		return
	}
	resource, err := h.services.ServiceResource().GetSingle(context.Background(), &company_service.GetSingleServiceResourceReq{
		EnvironmentId: environmentId.(string),
		ProjectId:     user.ProjectId,
		ServiceType:   company_service.ServiceType_BUILDER_SERVICE,
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	// if util.IsValidUUID(resourceId.(string)) {
	// 	resourceEnvironment, err = h.services.ResourceService().GetResourceEnvironment(
	// 		c.Request.Context(),
	// 		&company_service.GetResourceEnvironmentReq{
	// 			EnvironmentId: environmentId.(string),
	// 			ResourceId:    resourceId.(string),
	// 		},
	// 	)
	// 	if err != nil {
	// 		h.handleResponse(c, http.GRPCError, err.Error())
	// 		return
	// 	}
	// 	if user.CompanyId == "" {
	// 		project, err := h.services.ProjectServiceClient().GetById(context.Background(), &company_service.GetProjectByIdRequest{
	// 			ProjectId: resourceEnvironment.GetProjectId(),
	// 		})
	// 		if err != nil {
	// 			h.handleResponse(c, http.GRPCError, err.Error())
	// 			return
	// 		}
	// 		user.CompanyId = project.GetCompanyId()
	// 		user.ProjectId = resourceEnvironment.GetProjectId()
	// 	}
	// } else {
	// 	if !util.IsValidUUID(user.GetProjectId()) {
	// 		h.handleResponse(c, http.BadRequest, errors.New("not valid project id"))
	// 		return
	// 	}
	// 	resourceEnvironment, err = h.services.ResourceService().GetDefaultResourceEnvironment(
	// 		c.Request.Context(),
	// 		&company_service.GetDefaultResourceEnvironmentReq{
	// 			ResourceId: resourceId.(string),
	// 			ProjectId:  user.GetProjectId(),
	// 		},
	// 	)
	// 	if err != nil {
	// 		h.handleResponse(c, http.GRPCError, err.Error())
	// 		return
	// 	}
	// }
	user.ResourceEnvironmentId = resource.ResourceEnvironmentId
	user.ResourceType = int32(resource.GetResourceType())
	user.EnvironmentId = resource.EnvironmentId
	resp, err := h.services.UserService().V2CreateUser(
		c.Request.Context(),
		&user,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// V2GetUserList godoc
// @ID get_user_list_v2
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/user [GET]
// @Summary Get User List
// @Description  Get User List
// @Tags V2_User
// @Accept json
// @Produce json
// @Param offset query integer false "offset"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Param client-platform-id query string false "client-platform-id"
// @Param client-type-id query string false "client-type-id"
// @Param project-id query string true "project-id"
// @Success 200 {object} http.Response{data=auth_service.GetUserListResponse} "GetUserListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetUserList(c *gin.Context) {
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

	projectId := c.DefaultQuery("project-id", "")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.BadEnvironment, "project-id is required")
		return
	}
	log.Println("project_id:::::", projectId)

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	resource, err := h.services.ServiceResource().GetSingle(c.Request.Context(), &company_service.GetSingleServiceResourceReq{
		ProjectId:     projectId,
		EnvironmentId: environmentId.(string),
		ServiceType:   pb.ServiceType_BUILDER_SERVICE,
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	resp, err := h.services.UserService().V2GetUserList(

		c.Request.Context(),
		&auth_service.GetUserListRequest{
			Limit:                 int32(limit),
			Offset:                int32(offset),
			Search:                c.Query("search"),
			ClientPlatformId:      c.Query("client-platform-id"),
			ClientTypeId:          c.Query("client-type-id"),
			ProjectId:             projectId,
			ResourceEnvironmentId: resource.GetResourceEnvironmentId(),
			ResourceType:          int32(resource.GetResourceType()),
			NodeType:              resource.NodeType,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2GetUserByID godoc
// @ID get_user_by_id_v2
// @Router /v2/user/{user-id} [GET]
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Summary Get User By ID
// @Description Get User By ID
// @Tags V2_User
// @Accept json
// @Produce json
// @Param user-id path string true "user-id"
// @Param project-id query string true "project-id"
// @Param client-type-id query string false "client-type-id"
// @Success 200 {object} http.Response{data=auth_service.User} "UserBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetUserByID(c *gin.Context) {
	var (
		// resourceEnvironment *company_service.ResourceEnvironment
		err error
	)
	userID := c.Param("user-id")

	if !util.IsValidUUID(userID) {
		h.handleResponse(c, http.InvalidArgument, "user-id is an invalid uuid")
		return
	}
	clientTypeID := c.Query("client-type-id")
	if clientTypeID == "" {
		h.handleResponse(c, http.InvalidArgument, "client type id is required")
		return
	}

	if !util.IsValidUUID(clientTypeID) {
		h.handleResponse(c, http.InvalidArgument, "client type id is an invalid uuid")
		return
	}

	projectID := c.Query("project-id")

	if !util.IsValidUUID(projectID) {
		h.handleResponse(c, http.InvalidArgument, "project-id is an invalid uuid")
		return
	}

	// resourceId, ok := c.Get("resource_id")
	// if !ok {
	// 	h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
	// 	return
	// }

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}
	resource, err := h.services.ServiceResource().GetSingle(c.Request.Context(), &company_service.GetSingleServiceResourceReq{
		ProjectId:     projectID,
		EnvironmentId: environmentId.(string),
		ServiceType:   pb.ServiceType_BUILDER_SERVICE,
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	resp, err := h.services.UserService().V2GetUserByID(
		c.Request.Context(),
		&auth_service.UserPrimaryKey{
			Id:                    userID,
			ResourceEnvironmentId: resource.ResourceEnvironmentId,
			ResourceType:          int32(resource.GetResourceType()),
			ClientTypeId:          clientTypeID,
			ProjectId:             resource.GetProjectId(),
			NodeType:              resource.NodeType,
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2UpdateUser godoc
// @ID update_user_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/user [PUT]
// @Summary Update User
// @Description Update User
// @Tags V2_User
// @Accept json
// @Produce json
// @Param user body auth_service.UpdateUserRequest true "UpdateUserRequestBody"
// @Success 200 {object} http.Response{data=auth_service.User} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2UpdateUser(c *gin.Context) {
	var (
		user auth_service.UpdateUserRequest
	)

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	// if !util.IsValidUUID(user.GetProjectId()) {
	// 	h.handleResponse(c, http.InvalidArgument, "project-id is an invalid uuid")
	// 	return
	// }
	projectId, ok := c.Get("project_id")
	if !ok || !util.IsValidUUID(projectId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get project-id in query param"))
		return
	}

	// resourceId, ok := c.Get("resource_id")
	// if !ok {
	// 	h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
	// 	return
	// }

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}
	resource, err := h.services.ServiceResource().GetSingle(context.Background(), &pb.GetSingleServiceResourceReq{
		EnvironmentId: environmentId.(string),
		ProjectId:     projectId.(string),
		ServiceType:   pb.ServiceType_BUILDER_SERVICE,
	})

	user.ResourceType = int32(resource.GetResourceType())
	user.ResourceEnvironmentId = resource.ResourceEnvironmentId
	user.EnvironmentId = environmentId.(string)
	user.NodeType = resource.NodeType

	resp, err := h.services.UserService().V2UpdateUser(
		c.Request.Context(),
		&user,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2DeleteUser godoc
// @ID delete_user_v2
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/user/{user-id} [DELETE]
// @Summary Delete User
// @Description Get User
// @Tags V2_User
// @Accept json
// @Produce json
// @Param user-id path string true "user-id"
// @Param project-id query string true "project-id"
// @Param client-type-id query string true "client-type-id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2DeleteUser(c *gin.Context) {
	var (
		// userDataToMap       = make(map[string]interface{})
		resourceEnvironment *company_service.ResourceEnvironment
		err                 error
	)
	userID := c.Param("user-id")

	if !util.IsValidUUID(userID) {
		h.handleResponse(c, http.InvalidArgument, "user id is an invalid uuid")
		return
	}
	clientTypeID := c.Query("client-type-id")
	if clientTypeID == "" {
		h.handleResponse(c, http.InvalidArgument, "client type id is required")
		return
	}

	if !util.IsValidUUID(clientTypeID) {
		h.handleResponse(c, http.InvalidArgument, "client type id is an invalid uuid")
		return
	}

	projectID := c.Query("project-id")

	if !util.IsValidUUID(projectID) {
		h.handleResponse(c, http.InvalidArgument, "project-id is an invalid uuid")
		return
	}

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
	project, err := h.services.ProjectServiceClient().GetById(context.Background(), &pb.GetProjectByIdRequest{
		ProjectId: projectID,
	})

	if util.IsValidUUID(resourceId.(string)) {
		resourceEnvironment, err = h.services.ResourceService().GetResourceEnvironment(
			c.Request.Context(),
			&company_service.GetResourceEnvironmentReq{
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
			&company_service.GetDefaultResourceEnvironmentReq{
				ResourceId: resourceId.(string),
				ProjectId:  projectID,
			},
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	resp, err := h.services.UserService().V2DeleteUser(
		c.Request.Context(),
		&auth_service.UserPrimaryKey{
			Id:                    userID,
			ProjectId:             projectID,
			ResourceType:          resourceEnvironment.ResourceType,
			ResourceEnvironmentId: resourceEnvironment.Id,
			ClientTypeId:          clientTypeID,
			CompanyId:             project.CompanyId,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	// userDataToMap["id"] = userID
	// structData, err := helper.ConvertMapToStruct(userDataToMap)
	// if err != nil {
	// 	h.handleResponse(c, http.InvalidArgument, err.Error())
	// 	return
	// }

	// _, err = h.services.ObjectBuilderService().Delete(
	// 	context.Background(),
	// 	&obs.CommonMessage{
	// 		TableSlug: "user",
	// 		Data:      structData,
	// 		ProjectId: resourceEnvironment.GetId(),
	// 	},
	// )

	// if err != nil {
	// 	h.handleResponse(c, http.InternalServerError, err.Error())
	// 	return
	// }

	h.handleResponse(c, http.NoContent, resp)
}

// AddUserToProject godoc
// @ID add user to project
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Router /v2/add-user-to-project [POST]
// @Summary Create User
// @Description Create User
// @Tags V2_User
// @Accept json
// @Produce json
// @Param user body auth_service.AddUserToProjectReq true "AddUserToProjectReq"
// @Success 201 {object} http.Response{data=auth_service.AddUserToProjectRes} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) AddUserToProject(c *gin.Context) {
	var (
		req = auth_service.AddUserToProjectReq{}
	)

	err := c.ShouldBindJSON(&req)
	if err != nil {
		errCantParseReq := errors.New("cant parse json")
		h.log.Error("!!!AddUserToProject -> cant parse json")
		h.handleResponse(c, http.BadRequest, errCantParseReq.Error())
		return
	}

	if !util.IsValidUUID(req.GetProjectId()) {
		h.handleResponse(c, http.InvalidArgument, "project-id is an invalid uuid")
		return
	}
	if req.ClientTypeId == "" {
		h.handleResponse(c, http.InvalidArgument, "client_type_id is required")
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}
	projectId, ok := c.Get("project_id")
	if !ok || !util.IsValidUUID(projectId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get project-id in query param"))
		return
	}
	resource, err := h.services.ServiceResource().GetSingle(context.Background(), &company_service.GetSingleServiceResourceReq{
		EnvironmentId: environmentId.(string),
		ProjectId:     projectId.(string),
	})
	req.EnvId = environmentId.(string)

	res, err := h.services.UserService().AddUserToProject(
		c.Request.Context(),
		&req,
	)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	user, err := h.services.UserService().V2GetUserByID(
		c.Request.Context(),
		&auth_service.UserPrimaryKey{
			Id:                    req.UserId,
			ResourceEnvironmentId: resource.ResourceEnvironmentId,
			ProjectId:             resource.GetProjectId(),
			ClientTypeId:          req.ClientTypeId,
			ResourceType:          int32(resource.ResourceType.Number()),
			NodeType:              resource.NodeType,
		},
	)
	if err != nil {
		if errors.Is(err, config.ErrUserAlradyMember) {
			h.handleResponse(c, http.BadEnvironment, "already member!")
			return
		}

		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	var userDataToMap = make(map[string]interface{})
	userDataToMap["guid"] = req.UserId
	userDataToMap["active"] = req.Active
	userDataToMap["project_id"] = req.ProjectId
	userDataToMap["role_id"] = req.RoleId
	userDataToMap["client_type_id"] = user.ClientTypeId
	userDataToMap["client_platform_id"] = user.ClientPlatformId
	userDataToMap["from_auth_service"] = true

	structData, err := helper.ConvertMapToStruct(userDataToMap)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}
	var tableSlug = "user"
	switch int32(resource.ResourceType.Number()) {
	case 1:
		clientType, err := h.services.GetObjectBuilderServiceByType(req.NodeType).GetSingle(
			context.Background(),
			&obs.CommonMessage{
				TableSlug: "client_type",
				Data: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"id": structpb.NewStringValue(req.ClientTypeId),
					},
				},
				ProjectId: resource.ResourceEnvironmentId,
			},
		)
		if clientTypeTableSlug, ok := clientType.Data.AsMap()["table_slug"].(string); ok {
			tableSlug = clientTypeTableSlug
		}
		_, err = h.services.GetObjectBuilderServiceByType(req.NodeType).Create(
			context.Background(),
			&obs.CommonMessage{
				TableSlug: tableSlug,
				Data:      structData,
				ProjectId: resource.ResourceEnvironmentId,
			},
		)
		if err != nil {
			h.handleResponse(c, http.InternalServerError, err.Error())
			return
		}
	case 3:
		clientType, err := h.services.PostgresObjectBuilderService().GetSingle(
			context.Background(),
			&obs.CommonMessage{
				TableSlug: "client_type",
				Data: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"id": structpb.NewStringValue(req.ClientTypeId),
					},
				},
				ProjectId: resource.ResourceEnvironmentId,
			},
		)
		if clientTypeTableSlug, ok := clientType.Data.AsMap()["table_slug"].(string); ok {
			tableSlug = clientTypeTableSlug
		}
		_, err = h.services.PostgresObjectBuilderService().Create(
			context.Background(),
			&obs.CommonMessage{
				TableSlug: tableSlug,
				Data:      structData,
				ProjectId: resource.ResourceEnvironmentId,
			},
		)
		if err != nil {
			h.handleResponse(c, http.InternalServerError, err.Error())
			return
		}
	}

	h.handleResponse(c, http.Created, res)
}

// V2GetUserByLoginType godoc
// @ID get_user_by_login_type
// @Router /v2/user/check [POST]
// @Summary Get User By Login type
// @Description Get User By Login type
// @Tags V2_User
// @Accept json
// @Produce json
// @Param user-check body auth_service.GetUserByLoginTypesRequest true "user-check"
// @Success 200 {object} http.Response{data=auth_service.GetUserByLoginTypesResponse} "UserBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetUserByLoginType(c *gin.Context) {

	var (
		request auth_service.GetUserByLoginTypesRequest
	)

	err := c.ShouldBindJSON(&request)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if request.Email != "" {
		var isValid = util.IsValidEmail(request.Email)
		if !isValid {
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		}
	}

	resp, err := h.services.UserService().V2GetUserByLoginTypes(
		c.Request.Context(),
		&request,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// V2UserResetPassword godoc
// @ID v2_user_reset_password
// @Router /v2/user/reset-password [PUT]
// @Summary Reset User password
// @Description Reset User Password
// @Tags V2_User
// @Accept json
// @Produce json
// @Param reset_password body auth_service.V2UserResetPasswordRequest true "ResetPasswordRequestBody"
// @Success 200 {object} http.Response{data=string} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2UserResetPassword(c *gin.Context) {

	var userPassword = &auth_service.V2UserResetPasswordRequest{}
	err := c.ShouldBindJSON(&userPassword)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err)
		return
	}

	user, err := h.services.UserService().V2ResetPassword(context.Background(), userPassword)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err)
		return
	}
	h.handleResponse(c, http.OK, user)
}
