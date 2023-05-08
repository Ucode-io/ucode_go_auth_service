package handlers

import (
	"context"
	"errors"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateAppleIdSettings godoc
// @ID createAppleIdSettings
// @Router /v2/apple-id-settings [POST]
// @Summary CreateAppleIdSettings
// @Description CreateAppleIdSettings
// @Tags AppleId
// @Accept json
// @Produce json
// @Param X-API-KEY header string false "X-API-KEY"
// @Param registerBody body pb.AppleIdSettings true "register_body"
// @Success 201 {object} http.Response{data=pb.AppleIdSettings} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateAppleIdSettings(c *gin.Context) {

	var body *pb.AppleIdSettings

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

	resp, err := h.services.AppleIdService().CreateAppleIdSettings(
		c.Request.Context(),
		&pb.AppleIdSettings{
			Id:        uuid.String(),
			ProjectId: body.ProjectId,
			ClientId:  body.ClientId,
			TeamId:    body.TeamId,
			KeyId:     body.KeyId,
			Secret:    body.Secret,
		},
	)
	if err != nil {
		h.log.Error("---> error in create apple id settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// UpdateAppleIdSettings godoc
// @ID updateAppleIdSettings
// @Router /v2/apple-id-settings [PUT]
// @Summary UpdateAppleIdSettings
// @Description UpdateAppleIdSettings
// @Tags AppleId
// @Accept json
// @Produce json
// @Param registerBody body pb.AppleIdSettings true "register_body"
// @Success 200 {object} http.Response{data=pb.AppleIdSettings} "Apple Config data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateAppleIdSettings(c *gin.Context) {

	var body *pb.AppleIdSettings

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.AppleIdService().UpdateAppleIdSettings(
		c.Request.Context(),
		&pb.AppleIdSettings{
			Id:        body.Id,
			ProjectId: body.ProjectId,
			ClientId:  body.ClientId,
			TeamId:    body.TeamId,
			KeyId:     body.KeyId,
			Secret:    body.Secret,
		},
	)
	if err != nil {
		h.log.Error("---> error in update apple Id settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// GetListAppleIdSettings godoc
// @ID getListAppleIdSettings
// @Router /v2/apple-id-settings [GET]
// @Summary GetListAppleIdSettings
// @Description GetListAppleIdSettings
// @Tags AppleId
// @Accept json
// @Produce json
// @Param project_id query string true "project_id"
// @Success 200 {object} http.Response{data=pb.GetListAppleIdSettingsResponse} "Apple Config data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetAppleIdSettings(c *gin.Context) {

	resp, err := h.services.AppleIdService().GetListAppleIdSettings(
		c.Request.Context(),
		&pb.GetListAppleIdSettingsRequest{
			ProjectId: c.Query("project_id"),
		},
	)
	if err != nil {
		h.log.Error("---> error in get list apple settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, resp)
}

// DeleteAppleIdSettings godoc
// @ID deleteAppleIdSettings
// @Router /v2/apple-id-settings/{id} [DELETE]
// @Summary DeleteAppleIdSettings
// @Description DeleteAppleIdSettings
// @Tags AppleId
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) DeleteAppleIdSettings(c *gin.Context) {

	id := c.Param("id")

	resp, err := h.services.AppleIdService().DeleteAppleIdSettings(
		c.Request.Context(),
		&pb.AppleIdSettingsPrimaryKey{
			Id: id,
		},
	)
	
	if err != nil {
		h.log.Error("---> error in delete apple settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}


	h.handleResponse(c, http.NoContent, resp)
}

func (h *Handler) GetAppleConfig(projectId string) (*models.AppleConfig , error) {
	
	resp, err := h.services.AppleIdService().GetListAppleIdSettings(
		context.Background(),
		&pb.GetListAppleIdSettingsRequest{
			ProjectId: projectId,
		},
	)
	if err != nil {
		h.log.Error("---> error in apple id config ", logger.Error(err))
		
		return nil, err
	}

	if len(resp.Items)<1{
		return nil, errors.New("project hasn't apple configs")
	}
	return &models.AppleConfig{
		TeamId: resp.Items[0].TeamId,
		ClientId: resp.Items[0].ClientId,
		KeyId: resp.Items[0].KeyId,
		SecretKey: resp.Items[0].Secret,
	},nil

}
