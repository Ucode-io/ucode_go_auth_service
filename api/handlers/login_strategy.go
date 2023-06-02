package handlers

import (
	"context"
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/helper"

	"github.com/gin-gonic/gin"
	"github.com/saidamir98/udevs_pkg/util"
)

// LoginStrategyList godoc
// @ID login_strategy_list
// @Router /v2/login-strategy [GET]
// @Summary List Login Strategy
// @Description Get List Login Strategy
// @Tags V2_LoginStrategy
// @Accept json
// @Produce json
// @Param project-id query string true "project-id"
// @Success 201 {object} http.Response{data=auth_service.GetListResponse} "LoginStrategy data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetLoginStrategy(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	projectId := c.Query("project-id")
	if !util.IsValidUUID(projectId) {
		h.handleResponse(c, http.InvalidArgument, "project id is an invalid uuid")
		return
	}
	res, err := h.services.LoginStrategyService().GetList(ctx, &auth_service.GetListRequest{
		ProjectId: projectId,
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, res)
}

// LoginStrategyById godoc
// @ID login_strategy_by_id
// @Router /v2/login-strategy/{login-strategy-id} [GET]
// @Summary Get Login Strategy
// @Description Get By Id Login Strategy
// @Tags V2_LoginStrategy
// @Accept json
// @Produce json
// @Param login-strategy-id path string true "login-strategy-id"
// @Success 201 {object} http.Response{data=auth_service.GetListResponse} "LoginStrategy data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetLoginStrategyById(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	loginStrategyId := c.Param("login-strategy-id")
	if !util.IsValidUUID(loginStrategyId) {
		h.handleResponse(c, http.InvalidArgument, "login-strategy-id is an invalid uuid")
		return
	}

	res, err := h.services.LoginStrategyService().GetByID(ctx, &auth_service.LoginStrategyPrimaryKey{
		Id: loginStrategyId,
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}
	h.handleResponse(c, http.Created, res)
}

// UpsertLoginStrategy godoc
// @ID login_strategy_upsert
// @Router /v2/upsert-login-strategy [POST]
// @Summary Upsert Login Strategy
// @Description Upsert Login Strategy
// @Tags V2_LoginStrategy
// @Accept json
// @Produce json
// @Param login-strategy body auth_service.UpdateRequest true "upsert login strategy request body"
// @Success 201 {object} http.Response{data=auth_service.UpdateResponse} "LoginStrategy data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpsertLoginStrategy(c *gin.Context) {

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	var (
		loginStrategy auth_service.UpdateRequest
	)
	err := c.ShouldBindJSON(&loginStrategy)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	for _, value := range loginStrategy.LoginStrategies {
		err := helper.ParsePsqlTypeToEnum(value.Type)
		if err != nil {
			h.handleResponse(c, http.InvalidArgument, err.Error())
			return
		}
	}
	res, err := h.services.LoginStrategyService().Upsert(ctx, &auth_service.UpdateRequest{
		LoginStrategies: loginStrategy.GetLoginStrategies(),
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, res)
}
