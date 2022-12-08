package handlers

import (
	"ucode/ucode_go_auth_service/api/http"

	"ucode/ucode_go_auth_service/genproto/auth_service"

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
// @Param project body auth_service.CreateProjectRequest true "CreateProjectRequestBody"
// @Success 201 {object} http.Response{data=auth_service.Project} "Project data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateProject(c *gin.Context) {
	var project auth_service.CreateProjectRequest

	err := c.ShouldBindJSON(&project)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.ProjectService().Create(
		c.Request.Context(),
		&project,
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
// @Success 200 {object} http.Response{data=auth_service.GetProjectListResponse} "GetProjectListResponseBody"
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

	resp, err := h.services.ProjectService().GetList(
		c.Request.Context(),
		&auth_service.GetProjectListRequest{
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

// GetProjectByID godoc
// @ID get_project_by_id
// @Router /project/{project-id} [GET]
// @Summary Get Project By ID
// @Description Get Project By ID
// @Tags Project
// @Accept json
// @Produce json
// @Param project-id path string true "project-id"
// @Success 200 {object} http.Response{data=auth_service.Project} "ProjectBody"
// @Response 400 {object} http.Response{data=string} "Invalid Argument"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetProjectByID(c *gin.Context) {
	projectID := c.Param("project-id")

	if !util.IsValidUUID(projectID) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}

	resp, err := h.services.ProjectService().GetByPK(
		c.Request.Context(),
		&auth_service.ProjectPrimaryKey{
			Id: projectID,
		},
	)

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
// @Param project body auth_service.UpdateProjectRequest true "UpdateProjectRequestBody"
// @Success 200 {object} http.Response{data=auth_service.Project} "Project data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateProject(c *gin.Context) {
	var project auth_service.UpdateProjectRequest

	err := c.ShouldBindJSON(&project)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.ProjectService().Update(
		c.Request.Context(),
		&project,
	)

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

	resp, err := h.services.ProjectService().Delete(
		c.Request.Context(),
		&auth_service.ProjectPrimaryKey{
			Id: projectID,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}
