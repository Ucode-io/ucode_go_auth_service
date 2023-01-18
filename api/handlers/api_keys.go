package handlers

import (
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/genproto/auth_service"

	"github.com/gin-gonic/gin"
)

// CreateApiKey godoc
// @ID create_api_keys
// @Router /v2/api-key/{project-id} [POST]
// @Summary Create ApiKey
// @Description Create ApiKey
// @Tags V2_ApiKey
// @Accept json
// @Produce json
// @Param project-id path string true "project-id"
// @Param api-key body auth_service.CreateReq true "ApiKeyReqBody"
// @Success 201 {object} http.Response{data=auth_service.CreateRes} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateApiKey(c *gin.Context) {
	var apiKey auth_service.CreateReq

	err := c.ShouldBindJSON(&apiKey)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	res, err := h.services.ApiKeysService().Create(
		c.Request.Context(),
		&apiKey,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, res)
}

// UpdateApiKey godoc
// @ID update_api_keys
// @Router /v2/api-key/{project-id}/{id} [PUT]
// @Summary Update ApiKey
// @Description Update ApiKey
// @Tags V2_ApiKey
// @Accept json
// @Produce json
// @Param project-id path string true "project-id"
// @Param id path string true "id"
// @Param api-key body auth_service.UpdateReq true "ApiKeyReqBody"
// @Success 201 {object} http.Response{data=auth_service.UpdateRes} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateApiKey(c *gin.Context) {
	var apiKey auth_service.UpdateReq

	err := c.ShouldBindJSON(&apiKey)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	apiKey.Id = c.Param("id")

	res, err := h.services.ApiKeysService().Update(
		c.Request.Context(),
		&apiKey,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}

// GetApiKey godoc
// @ID get_api_key
// @Router /v2/api-key/{project-id}/{id} [GET]
// @Summary Get ApiKey
// @Description Get ApiKey
// @Tags V2_ApiKey
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Param project-id path string true "project-id"
// @Success 201 {object} http.Response{data=auth_service.GetRes} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetApiKey(c *gin.Context) {

	res, err := h.services.ApiKeysService().Get(
		c.Request.Context(),
		&auth_service.GetReq{
			Id: c.Param("id"),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}

// GetListApiKeys godoc
// @ID get_list_api_key
// @Router /v2/api-key/{project-id} [GET]
// @Summary Get List ApiKey
// @Description Get List ApiKey
// @Tags V2_ApiKey
// @Accept json
// @Produce json
// @Param project-id path string true "project-id"
// @Param resource-environment-id query string false "resource-environment-id"
// @Param offset query integer false "offset"
// @Param limit query integer false "limit"
// @Param search query string false "search"
// @Success 201 {object} http.Response{data=auth_service.GetListRes} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetListApiKeys(c *gin.Context) {

	res, err := h.services.ApiKeysService().GetList(
		c.Request.Context(),
		&auth_service.GetListReq{
			ResourceEnvironmentId: c.DefaultQuery("resource-environment-id", ""),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}

// DeleteApiKeys godoc
// @ID delete_api_keys
// @Router /v2/api-key/{project-id}/{id} [DELETE]
// @Summary Delete ApiKeys
// @Description Delete ApiKeys
// @Tags V2_ApiKey
// @Accept json
// @Produce json
// @Param project-id path string true "project-id"
// @Param id path string true "id"
// @Success 201 {object} http.Response{data=auth_service.DeleteRes} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) DeleteApiKeys(c *gin.Context) {

	res, err := h.services.ApiKeysService().Delete(
		c.Request.Context(),
		&auth_service.DeleteReq{Id: c.Param("id")},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}

// GenerateApiKeyToken godoc
// @ID generate_api_key_token
// @Router /v2/api-key/generate-token [POST]
// @Summary Generate Api Key Token
// @Description Generate Api Key Token
// @Tags V2_ApiKey
// @Accept json
// @Produce json
// @Param api-key body auth_service.GenerateApiTokenReq true "ApiKeyReqBody"
// @Success 201 {object} http.Response{data=auth_service.GenerateApiTokenRes} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GenerateApiKeyToken(c *gin.Context) {
	var apiKey auth_service.GenerateApiTokenReq

	err := c.ShouldBindJSON(&apiKey)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	res, err := h.services.ApiKeysService().GenerateApiToken(
		c.Request.Context(),
		&apiKey,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}

// RefreshApiKeyToken godoc
// @ID refresh_api_key_token
// @Router /v2/api-key/refresh-token [POST]
// @Summary Refresh Api Key Token
// @Description Refresh Api Key Token
// @Tags V2_ApiKey
// @Accept json
// @Produce json
// @Param api-key body auth_service.RefreshApiTokenReq true "ApiKeyReqBody"
// @Success 201 {object} http.Response{data=auth_service.RefreshApiTokenReq} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) RefreshApiKeyToken(c *gin.Context) {
	var apiKey auth_service.RefreshApiTokenReq

	err := c.ShouldBindJSON(&apiKey)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	res, err := h.services.ApiKeysService().RefreshApiToken(
		c.Request.Context(),
		&apiKey,
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}
