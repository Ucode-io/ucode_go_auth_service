package handlers

import (
	"errors"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/pkg/helper"

	"ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/saidamir98/udevs_pkg/util"
	"github.com/spf13/cast"

	"ucode/ucode_go_auth_service/api/models"

	"github.com/gin-gonic/gin"
)

// CreateUser godoc
// @ID create_user
// @Router /user [POST]
// @Summary Create User
// @Description Create User
// @Tags User
// @Accept json
// @Produce json
// @Param user body auth_service.CreateUserRequest true "CreateUserRequestBody"
// @Success 201 {object} http.Response{data=auth_service.User} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateUser(c *gin.Context) {
	var (
		user auth_service.CreateUserRequest
		resp *auth_service.User
	)

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	userId, _ := c.Get("user_id")

	var (
		logReq = &models.CreateVersionHistoryRequest{
			NodeType:     user.NodeType,
			ProjectId:    user.ResourceEnvironmentId,
			ActionSource: c.Request.URL.String(),
			ActionType:   "CREATE",
			UsedEnvironments: map[string]bool{
				cast.ToString(user.EnvironmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   &user,
			TableSlug: "USER",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
			h.log.Info("!!!CreateUser -> error")
		} else {
			logReq.Response = resp
			h.log.Info("CreateUser -> success")
		}
		go h.versionHistory(c, logReq)
	}()

	resp, err = h.services.UserService().CreateUser(
		c.Request.Context(),
		&user,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// GetUserList godoc
// @ID get_user_list
// @Router /user [GET]
// @Summary Get User List
// @Description  Get User List
// @Tags User
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
func (h *Handler) GetUserList(c *gin.Context) {
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

	resp, err := h.services.UserService().GetUserList(
		c.Request.Context(),
		&auth_service.GetUserListRequest{
			Limit:            int32(limit),
			Offset:           int32(offset),
			Search:           c.Query("search"),
			ClientPlatformId: c.Query("client-platform-id"),
			ClientTypeId:     c.Query("client-type-id"),
			ProjectId:        c.Query("project-id"),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// GetUserByID godoc
// @ID get_user_by_id
// @Router /user/{user-id} [GET]
// @Summary Get User By ID
// @Description Get User By ID
// @Tags User
// @Accept json
// @Produce json
// @Param user-id path string true "user-id"
// @Success 200 {object} http.Response{data=auth_service.User} "UserBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetUserByID(c *gin.Context) {
	userID := c.Param("user-id")

	if !util.IsValidUUID(userID) {
		h.handleResponse(c, http.InvalidArgument, "user id is an invalid uuid")
		return
	}

	resp, err := h.services.UserService().GetUserByID(
		c.Request.Context(),
		&auth_service.UserPrimaryKey{
			Id: userID,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// UpdateUser godoc
// @ID update_user
// @Router /user [PUT]
// @Summary Update User
// @Description Update User
// @Tags User
// @Accept json
// @Produce json
// @Param user body auth_service.UpdateUserRequest true "UpdateUserRequestBody"
// @Success 200 {object} http.Response{data=auth_service.User} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateUser(c *gin.Context) {
	var (
		user auth_service.UpdateUserRequest
		resp *auth_service.User
	)

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	userId, _ := c.Get("user_id")

	var (
		logReq = &models.CreateVersionHistoryRequest{
			NodeType:     user.NodeType,
			ProjectId:    user.ResourceEnvironmentId,
			ActionSource: c.Request.URL.String(),
			ActionType:   "UPDATE",
			UsedEnvironments: map[string]bool{
				cast.ToString(user.EnvironmentId): true,
			},
			UserInfo:  cast.ToString(userId),
			Request:   &user,
			TableSlug: "USER",
		}
	)

	defer func() {
		if err != nil {
			logReq.Response = err.Error()
			h.log.Info("!!!UpdateUser -> error")
		} else {
			logReq.Response = resp
			h.log.Info("UpdateUser -> success")
		}
		go h.versionHistory(c, logReq)
	}()

	resp, err = h.services.UserService().UpdateUser(
		c.Request.Context(),
		&user,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// DeleteUser godoc
// @ID delete_user
// @Router /user/{user-id} [DELETE]
// @Summary Delete User
// @Description Get User
// @Tags User
// @Accept json
// @Produce json
// @Param user-id path string true "user-id"
// @Param project-id path string true "project-id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) DeleteUser(c *gin.Context) {
	// var userDataToMap = make(map[string]interface{})
	// userID := c.Param("user-id")
	// projectID := c.Param("project-id")

	// if !util.IsValidUUID(userID) {
	// 	h.handleResponse(c, http.InvalidArgument, "user id is an invalid uuid")
	// 	return
	// }

	// _, err := h.services.UserService().DeleteUser(
	// 	c.Request.Context(),
	// 	&auth_service.UserPrimaryKey{
	// 		Id: userID,
	// 	},
	// )

	// if err != nil {
	// 	h.handleResponse(c, http.GRPCError, err.Error())
	// 	return
	// }
	// userDataToMap["id"] = userID
	// structData, err := helper.ConvertMapToStruct(userDataToMap)
	// if err != nil {
	// 	h.handleResponse(c, http.InvalidArgument, err.Error())
	// 	return
	// }

	// _, err = h.services.SyncUserService().DeleteUser(
	// 	context.Background(),
	// 	&auth_service.DeleteSyncUserRequest{
	// 		ProjectId: projectID,
	// 		UserId:    userID,
	// 	},
	// )

	// h.handleResponse(c, http.NoContent, userDataToMap)
}

// AddUserRelation godoc
// @ID add_user_relation
// @Router /user-relation [POST]
// @Summary Create UserRelation
// @Description Create UserRelation
// @Tags UserRelation
// @Accept json
// @Produce json
// @Param user-relation body auth_service.AddUserRelationRequest true "AddUserRelationRequestBody"
// @Success 201 {object} http.Response{data=auth_service.UserRelation} "UserRelation data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) AddUserRelation(c *gin.Context) {
	var user_relation auth_service.AddUserRelationRequest

	err := c.ShouldBindJSON(&user_relation)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.UserService().AddUserRelation(
		c.Request.Context(),
		&user_relation,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// RemoveUserRelation godoc
// @ID delete_user_relation
// @Router /user-relation [DELETE]
// @Summary Delete UserRelation
// @Description Get UserRelation
// @Tags UserRelation
// @Accept json
// @Produce json
// @Param user-relation body auth_service.UserRelationPrimaryKey true "UserRelationPrimaryKeyBody"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) RemoveUserRelation(c *gin.Context) {
	var user_relation auth_service.UserRelationPrimaryKey

	err := c.ShouldBindJSON(&user_relation)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.UserService().RemoveUserRelation(
		c.Request.Context(),
		&user_relation,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// UpsertUserInfo godoc
// @ID upsert_user_info
// @Router /upsert-user-info/{user-id} [POST]
// @Summary Upsert UserInfo
// @Description Upsert UserInfo
// @Tags UpsertUserInfo
// @Accept json
// @Produce json
// @Param data body models.StructBody true "UpsertUserInfoRequestBody"
// @Param user-id path string true "user-id"
// @Success 201 {object} http.Response{data=auth_service.Role} "Role data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpsertUserInfo(c *gin.Context) {
	var data models.StructBody
	userID := c.Param("user-id")

	if !util.IsValidUUID(userID) {
		h.handleResponse(c, http.InvalidArgument, "user id is an invalid uuid")
		return
	}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	structPb, err := helper.ConvertMapToStruct(data.Body)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	resp, err := h.services.UserService().UpsertUserInfo(
		c.Request.Context(),
		&auth_service.UpsertUserInfoRequest{
			UserId: userID,
			Data:   structPb,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// UpdateUser godoc
// @ID reset_password
// @Router /user/reset-password [PUT]
// @Summary Update User
// @Description Reset Password
// @Tags User
// @Accept json
// @Produce json
// @Param reset_password body auth_service.ResetPasswordRequest true "ResetPasswordRequestBody"
// @Success 200 {object} http.Response{data=string} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) ResetPassword(c *gin.Context) {
	var user auth_service.ResetPasswordRequest

	err := c.ShouldBindJSON(&user)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.UserService().ResetPassword(
		c.Request.Context(),
		&user,
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	login, err := h.services.SessionService().Login(
		c.Request.Context(),
		&auth_service.LoginRequest{
			Username: resp.GetLogin(),
			Password: user.GetPassword(),
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, login)
}

// UpdateUser godoc
// @ID send_message_to_user_email
// @Router /user/send-message [POST]
// @Summary Send Message To User
// @Description Send Message to User Email
// @Tags User
// @Accept json
// @Produce json
// @Param send_message body auth_service.SendMessageToEmailRequest true "SendMessageToEmailRequestBody"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) SendMessageToUserEmail(c *gin.Context) {
	var customerMessage auth_service.SendMessageToEmailRequest

	err := c.ShouldBindJSON(&customerMessage)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.UserService().SendMessageToEmail(
		c.Request.Context(),
		&customerMessage,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// AddUserProject godoc
// @ID add user project
// @Router /add-user-project [POST]
// @Summary Add User Project
// @Description Add User Project
// @Tags User
// @Accept json
// @Produce json
// @Param user body auth_service.AddUserToProjectReq true "AddUserToProjectReq"
// @Success 201 {object} http.Response{data=auth_service.AddUserToProjectRes} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) AddUserProject(c *gin.Context) {
	var req = auth_service.AddUserToProjectReq{}
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

	if !util.IsValidUUID(req.GetEnvId()) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	if req.ClientTypeId == "" {
		h.handleResponse(c, http.InvalidArgument, "client_type_id is required")
		return
	}

	resp, err := h.services.UserService().AddUserToProject(
		c.Request.Context(),
		&req,
	)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// DeleteManyUserProject godoc
// @ID delete many user project
// @Router /delete-many-user-project [POST]
// @Summary Delete Many User Project
// @Description Delete Many User Project
// @Tags User
// @Accept json
// @Produce json
// @Param user body auth_service.DeleteManyUserRequest true "DeleteManyUserRequest"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) DeleteManyUserProject(c *gin.Context) {
	var (
		req = auth_service.DeleteManyUserRequest{}
	)

	err := c.ShouldBindJSON(&req)
	if err != nil {
		errCantParseReq := errors.New("cant parse json")
		h.log.Error("!!!DeleteManyUserToProject -> cant parse json")
		h.handleResponse(c, http.BadRequest, errCantParseReq.Error())
		return
	}

	if !util.IsValidUUID(req.GetProjectId()) {
		h.handleResponse(c, http.InvalidArgument, "project-id is an invalid uuid")
		return
	}

	if !util.IsValidUUID(req.GetCompanyId()) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get company_id"))
		return
	}

	if !util.IsValidUUID(req.GetEnvironmentId()) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	resp, err := h.services.SyncUserService().DeleteManyUser(
		c.Request.Context(),
		&req,
	)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}
