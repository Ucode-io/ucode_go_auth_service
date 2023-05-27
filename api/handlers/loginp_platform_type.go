package handlers

import (
	"ucode/ucode_go_auth_service/api/http"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateLoginPlatformType godoc
// @ID createLoginPlatformType
// @Router /v2/login-platform-type [POST]
// @Summary CreateLoginPlatformType
// @Description CreateLoginPlatformType
// @Tags LoginId
// @Accept json
// @Produce json
// @Param X-API-KEY header string false "X-API-KEY"
// @Param registerBody body pb.LoginPlatform true "register_body"
// @Success 201 {object} http.Response{data=pb.LoginPlatform} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateLoginPlatformType(c *gin.Context) {

	var body *pb.LoginPlatform

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	resp, err := h.services.LoginPlatformType().CreateLoginPlatformType(
		c.Request.Context(),
		&pb.LoginPlatform{
			Id:            uuid.String(),
			ProjectId:     body.ProjectId,
			EnvironmentId: body.EnvironmentId,
			Data:          body.Data,
			Type:          body.Type,
		},
	)

	if err != nil {
		h.log.Error("---> error in create login id settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// UpdateLoginPlatformType godoc
// @ID updateLoginPlatformType
// @Router /v2/login-platform-type [PUT]
// @Summary UpdateLoginPlatformType
// @Description UpdateLoginPlatformType
// @Tags LoginId
// @Accept json
// @Produce json
// @Param registerBody body pb.LoginPlatform true "register_body"
// @Success 200 {object} http.Response{data=pb.LoginPlatformType} "Login Config data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateLoginPlatformType(c *gin.Context) {

	var body *pb.LoginPlatform

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.LoginPlatformType().UpdateLoginPlatformType(
		c.Request.Context(),
		&pb.LoginPlatform{
			Id:        body.Id,
			ProjectId: body.ProjectId,
			Data:      body.Data,
		},
	)
	if err != nil {
		h.log.Error("---> error in update login Id settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// GetListLoginPlatformType godoc
// @ID getListLoginPlatformType
// @Router /v2/login-platform-type [GET]
// @Summary GetListLoginPlatformType
// @Description GetListLoginPlatformType
// @Tags LoginId
// @Accept json
// @Produce json
// @Param project_id query string true "project_id"
// @Success 200 {object} http.Response{data=pb.GetListLoginPlatformTypeResponse} "Login Config data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetLoginPlatformType(c *gin.Context) {

	resp, err := h.services.LoginPlatformType().GetListLoginPlatformType(
		c.Request.Context(),
		&pb.GetListLoginPlatformTypeRequest{
			ProjectId: c.Query("project_id"),
		},
	)
	if err != nil {
		h.log.Error("---> error in get list login settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// DeleteLoginPlatformType godoc
// @ID deleteLoginPlatformType
// @Router /v2/login-platform-type/{id} [DELETE]
// @Summary DeleteLoginPlatformType
// @Description DeleteLoginPlatformType
// @Tags LoginId
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) DeleteLoginPlatformType(c *gin.Context) {

	id := c.Param("id")

	resp, err := h.services.LoginPlatformType().DeleteLoginPlatformType(
		c.Request.Context(),
		&pb.LoginPlatformTypePrimaryKey{
			Id: id,
		},
	)

	if err != nil {
		h.log.Error("---> error in delete login settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, resp)
}

/*
func (h *Handler) GetLoginConfig(projectId string) (*models.LoginConfig, error) {

	resp, err := h.services.LoginIdService().GetListLoginPlatformType(
		context.Background(),
		&pb.GetListLoginPlatformTypeRequest{
			ProjectId: projectId,
		},
	)
	if err != nil {
		h.log.Error("---> error in login id config ", logger.Error(err))

		return nil, err
	}

	if len(resp.Items) < 1 {
		return nil, errors.New("project hasn't login configs")
	}
	return &models.LoginConfig{
		TeamId:    resp.Items[0].Data.TeamId,
		ClientId:  resp.Items[0].Data.ClientId,
		KeyId:     resp.Items[0].Data.KeyId,
		SecretKey: resp.Items[0].Data.Secret,
		Email:     resp.Items[0].Data.Email,
		Password:  resp.Items[0].Data.Password,
	}, nil
}
*/
