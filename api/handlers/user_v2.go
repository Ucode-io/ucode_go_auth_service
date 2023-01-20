package handlers

import (
	"context"
	"fmt"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/pkg/errors"

	"ucode/ucode_go_auth_service/genproto/auth_service"
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
		resourceEnvironment *company_service.ResourceEnvironment
		user                auth_service.CreateUserRequest
	)

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if !util.IsValidUUID(user.GetProjectId()) {
		h.handleResponse(c, http.BadRequest, errors.New("not valid project id"))
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
				ProjectId:  user.GetProjectId(),
			},
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	user.ResourceEnvironmentId = resourceEnvironment.GetId()

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
// @Param project-id query string false "project-id"
// @Success 200 {object} http.Response{data=auth_service.GetUserListResponse} "GetUserListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetUserList(c *gin.Context) {
	var resourceEnvironment *company_service.ResourceEnvironment
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
				ProjectId:  projectId,
			},
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
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
			ResourceEnvironmentId: resourceEnvironment.GetId(),
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
// @Success 200 {object} http.Response{data=auth_service.User} "UserBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2GetUserByID(c *gin.Context) {
	var (
		resourceEnvironment *company_service.ResourceEnvironment
		err                 error
	)
	userID := c.Param("user-id")

	if !util.IsValidUUID(userID) {
		h.handleResponse(c, http.InvalidArgument, "user-id is an invalid uuid")
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

	resp, err := h.services.UserService().V2GetUserByID(
		c.Request.Context(),
		&auth_service.UserPrimaryKey{
			Id:        userID,
			ProjectId: resourceEnvironment.GetId(),
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
		resourceEnvironment *company_service.ResourceEnvironment
		user                auth_service.UpdateUserRequest
	)

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if !util.IsValidUUID(user.GetProjectId()) {
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
				ProjectId:  user.GetProjectId(),
			},
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

	fmt.Println(resourceEnvironment.GetId())

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
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2DeleteUser(c *gin.Context) {
	var (
		userDataToMap       = make(map[string]interface{})
		resourceEnvironment *company_service.ResourceEnvironment
		err                 error
	)
	userID := c.Param("user-id")

	if !util.IsValidUUID(userID) {
		h.handleResponse(c, http.InvalidArgument, "user id is an invalid uuid")
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
			Id:        userID,
			ProjectId: projectID,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	userDataToMap["id"] = userID
	structData, err := helper.ConvertMapToStruct(userDataToMap)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	_, err = h.services.ObjectBuilderService().Delete(
		context.Background(),
		&obs.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: resourceEnvironment.GetId(),
		},
	)

	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

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
		req                 = auth_service.AddUserToProjectReq{}
		resourceEnvironment *company_service.ResourceEnvironment
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
				ProjectId:  req.GetProjectId(),
			},
		)
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
	}

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
			Id:        req.UserId,
			ProjectId: req.ProjectId,
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

	structData, err := helper.ConvertMapToStruct(userDataToMap)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	_, err = h.services.ObjectBuilderService().Create(
		context.Background(),
		&obs.CommonMessage{
			TableSlug: "user",
			Data:      structData,
			ProjectId: resourceEnvironment.GetId(),
		},
	)

	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, res)
}
