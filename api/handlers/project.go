package handlers

import (
	"ucode/ucode_go_auth_service/api/http"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"

	"github.com/saidamir98/udevs_pkg/util"

	"github.com/gin-gonic/gin"
)

// CreateProject godoc
// @ID create_project
// @Router /project [POST]
// @Summary Create Project
// @Description Create Project
// @Tags Project
// @Accept json
// @Produce json
// @Param user_id query string false "user_id"
// @Param project body company_service.CreateProjectRequest true "CreateProjectRequestBody"
// @Success 201 {object} http.Response{data=company_service.Project} "Project data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateProject(c *gin.Context) {
	//var project auth_service.CreateProjectRequest
	var (
		project company_service.CreateProjectRequest
		resp    *company_service.CreateProjectResponse
	)

	err := c.ShouldBindJSON(&project)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err = h.services.ProjectServiceClient().Create(
		c.Request.Context(),
		&project,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	_, err = h.services.UserService().AddUserToProject(
		c.Request.Context(),
		&pb.AddUserToProjectReq{
			UserId:    c.Query("user_id"),
			CompanyId: project.GetCompanyId(),
			ProjectId: resp.GetProjectId(),
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// GetProjectList godoc
// @ID get_project_list
// @Router /project [GET]
// @Summary Get Project List
// @Description  Get Project List
// @Tags Project
// @Accept json
// @Produce json
// @Param offset query integer false "offset"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Success 200 {object} http.Response{data=company_service.GetProjectListResponse} "GetProjectListResponseBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetProjectList(c *gin.Context) {
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

	resp, err := h.services.ProjectServiceClient().GetList(
		c.Request.Context(),
		&company_service.GetProjectListRequest{
			Limit:  int32(limit),
			Offset: int32(offset),
			Search: c.Query("search"),
		})

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// GetProjectByID godoc
// @ID get_project_by_id
// @Router /project/{project-id} [GET]
// @Summary Get Project By ID
// @Description Get Project By ID
// @Tags Project
// @Accept json
// @Produce json
// @Param project-id path string true "project-id"
// @Success 200 {object} http.Response{data=company_service.Project} "ProjectBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetProjectByID(c *gin.Context) {
	projectID := c.Param("project-id")

	if !util.IsValidUUID(projectID) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	resp, err := h.services.ProjectServiceClient().GetById(
		c.Request.Context(),
		&company_service.GetProjectByIdRequest{
			ProjectId: projectID,
		})

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// UpdateProject godoc
// @ID update_project
// @Router /project [PUT]
// @Summary Update Project
// @Description Update Project
// @Tags Project
// @Accept json
// @Produce json
// @Param project body company_service.Project true "UpdateProjectRequestBody"
// @Success 200 {object} http.Response{data=company_service.Project} "Project data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateProject(c *gin.Context) {
	var project company_service.Project

	err := c.ShouldBindJSON(&project)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.ProjectServiceClient().Update(
		c.Request.Context(),
		&project)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// DeleteProject godoc
// @ID delete_project
// @Router /project/{project-id} [DELETE]
// @Summary Delete Project
// @Description Get Project
// @Tags Project
// @Accept json
// @Produce json
// @Param project-id path string true "project-id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) DeleteProject(c *gin.Context) {
	projectID := c.Param("project-id")

	if !util.IsValidUUID(projectID) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	resp, err := h.services.ProjectServiceClient().Delete(
		c.Request.Context(),
		&company_service.DeleteProjectRequest{
			ProjectId: projectID,
		})

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

// UpdateProjectUserData godoc
// @ID update_user_in_project
// @Router /project/{project-id}/user-update [PUT]
// @Summary Update Project
// @Description Update Project
// @Tags Project
// @Accept json
// @Produce json
// @Param project body company_service.UpdateProjectUserDataReq true "UpdateProjectUserDataReqBody"
// @Success 200 {object} http.Response{data=company_service.UpdateProjectUserDataRes} "Project data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateProjectUserData(c *gin.Context) {
	var UpdateProjectUserDataReq company_service.UpdateProjectUserDataReq

	// err := c.ShouldBindJSON(&UpdateProjectUserDataReq)
	// if err != nil {
	// 	h.handleResponse(c, http.BadRequest, err.Error())
	// 	return
	// }

	// req, err := helper.ConvertMapToStruct(map[string]interface{}{
	// 	"client_type_id":     UpdateProjectUserDataReq.GetClientTypeId(),
	// 	"role_id":            UpdateProjectUserDataReq.GetRoleId(),
	// 	"client_platform_id": UpdateProjectUserDataReq.GetClientPlatformId(),
	// 	"guid":               UpdateProjectUserDataReq.GetUserId(),
	// })
	// if err != nil {
	// 	errConvertMapToStruct := errors.New("cant parse to struct")
	// 	h.handleResponse(c, http.InvalidArgument, errConvertMapToStruct.Error())
	// 	return
	// }

	// ctx, finish := context.WithTimeout(context.Background(), 20*time.Second)
	// defer finish()
	// resp, err := h.services.GetObjectBuilderServiceByType("").Update(
	// 	ctx,
	// 	&object_builder_service.CommonMessage{
	// 		TableSlug: "user",
	// 		Data:      req,
	// 		ProjectId: UpdateProjectUserDataReq.GetProjectId(),
	// 	},
	// )
	// if err != nil {
	// 	errUpdateUserData := errors.New("cant update user project data")
	// 	h.handleResponse(c, http.InvalidArgument, errUpdateUserData.Error())
	// 	return
	// }

	h.handleResponse(c, http.OK, UpdateProjectUserDataReq)
}
