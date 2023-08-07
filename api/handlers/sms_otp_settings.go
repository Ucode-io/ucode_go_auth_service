package handlers

import (
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
)

// CreateSmsOtpSettings godoc
// @ID create_api_keys
// @Router /v2/sms-otp-settings [POST]
// @Summary Create sms otp settings
// @Description Create sms otp settings
// @Tags V2_SMS_OTP_SETTINGS
// @Accept json
// @Produce json
// @Param project-id query string false "project-id"
// @Param Environment-Id header string false "environment-id"
// @Param X-API-KEY header string false "X-API-KEY"
// @Param sms-otp-settings body models.CreateSmsOtpSettingsRequest true "SmsOtpRequest ReqBody"
// @Success 201 {object} http.Response{data=auth_service.SmsOtpSettings} "SmsOtpRequest  data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateSmsOtpSettings(c *gin.Context) {
	var smsOtpSettings models.CreateSmsOtpSettingsRequest

	err := c.ShouldBindJSON(&smsOtpSettings)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	envId, ok := c.Get("environment_id")
	if !ok {
		h.handleResponse(c, http.InvalidArgument, "Environment-Id is required")
		return
	}
	projectId, ok := c.Get("project_id")
	if !ok {
		h.handleResponse(c, http.InvalidArgument, "project-id is required")
		return
	}
	ok = util.IsValidUUID(projectId.(string))
	if ok {
		h.handleResponse(c, http.InvalidArgument, "project-id is an invalid UUID")
		return
	}
	ok = util.IsValidUUID(envId.(string))
	if ok {
		h.handleResponse(c, http.InvalidArgument, "environment-id is an invalid UUID")
		return
	}

	res, err := h.services.SmsOtpSettingsService().Create(
		c.Request.Context(),
		&auth_service.CreateSmsOtpSettingsRequest{
			Login:         smsOtpSettings.Login,
			Password:      smsOtpSettings.Password,
			DefaultOtp:    smsOtpSettings.DefaultOtp,
			NumberOfOtp:   smsOtpSettings.NumberOfOtp,
			EnvironmentId: envId.(string),
			ProjectId:     projectId.(string),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, res)
}

// UpdateSmsOtpSettings godoc
// @ID update_api_keys
// @Router /v2/sms-otp-settings [PUT]
// @Summary Update sms otp settings
// @Description Update sms otp settings
// @Tags V2_SMS_OTP_SETTINGS
// @Accept json
// @Produce json
// @Param project-id query string false "project-id"
// @Param Environment-Id header string false "environment-id"
// @Param X-API-KEY header string false "X-API-KEY"
// @Param body body models.UpdateSmsOtpSettingsRequest true "SmsOtpSettingsReqBody"
// @Success 200 {object} http.Response{data=auth_service.SmsOtpSettings} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateSmsOtpSettings(c *gin.Context) {
	var smsOtpSettings models.UpdateSmsOtpSettingsRequest

	envId, ok := c.Get("environment_id")
	if !ok {
		h.handleResponse(c, http.InvalidArgument, "Environment-Id is required")
		return
	}
	projectId, ok := c.Get("project_id")
	if !ok {
		h.handleResponse(c, http.InvalidArgument, "project-id is required")
		return
	}
	ok = util.IsValidUUID(projectId.(string))
	if ok {
		h.handleResponse(c, http.InvalidArgument, "project-id is an invalid UUID")
		return
	}
	ok = util.IsValidUUID(envId.(string))
	if ok {
		h.handleResponse(c, http.InvalidArgument, "environment-id is an invalid UUID")
		return
	}
	err := c.ShouldBindJSON(&smsOtpSettings)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	res, err := h.services.SmsOtpSettingsService().Update(
		c.Request.Context(),
		&auth_service.SmsOtpSettings{
			Id:            smsOtpSettings.Id,
			Login:         smsOtpSettings.Login,
			Password:      smsOtpSettings.Password,
			DefaultOtp:    smsOtpSettings.DefaultOtp,
			NumberOfOtp:   smsOtpSettings.NumberOfOtp,
			EnvironmentId: envId.(string),
			ProjectId:     projectId.(string),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}

// GetListSmsOtpSettings godoc
// @ID get_list_sms_otp_settings
// @Router /v2/sms-otp-settings [GET]
// @Summary Get Sms Otp Settings
// @Description Get Sms Otp Settings
// @Tags V2_SMS_OTP_SETTINGS
// @Accept json
// @Produce json
// @Param Environment-Id header string false "environment-id"
// @Param project-id query string false "project-id"
// @Param X-API-KEY header string false "X-API-KEY"
// @Success 200 {object} http.Response{data=auth_service.SmsOtpSettingsResponse} "ApiKey data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetListSmsOtpSettings(c *gin.Context) {

	envId, ok := c.Get("environment_id")
	if !ok {
		h.handleResponse(c, http.InvalidArgument, "Environment-Id is required")
		return
	}
	projectId, ok := c.Get("project_id")
	if !ok {
		h.handleResponse(c, http.InvalidArgument, "project-id is required")
		return
	}
	ok = util.IsValidUUID(projectId.(string))
	if ok {
		h.handleResponse(c, http.InvalidArgument, "project-id is an invalid UUID")
		return
	}
	ok = util.IsValidUUID(envId.(string))
	if ok {
		h.handleResponse(c, http.InvalidArgument, "environment-id is an invalid UUID")
		return
	}

	res, err := h.services.SmsOtpSettingsService().GetList(
		c.Request.Context(),
		&auth_service.GetListSmsOtpSettingsRequest{
			ProjectId:     projectId.(string),
			EnvironmentId: envId.(string),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}

// GetByIdSmsOtpSettings godoc
// @ID get_sms_otp_settings_by_id
// @Router /v2/sms-otp-settings/{id} [GET]
// @Summary Get sms otp settings by id
// @Description Get sms otp settings by id
// @Tags V2_SMS_OTP_SETTINGS
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 200 {object} http.Response{data=auth_service.SmsOtpSettings} "SmsOtpSettings data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetByIdSmsOtpSettings(c *gin.Context) {

	id := c.Param("id")
	if ok := util.IsValidUUID(id); !ok {
		h.handleResponse(c, http.InvalidArgument, "sms-otp-settings is an invalid UUID")
		return
	}

	res, err := h.services.SmsOtpSettingsService().GetById(
		c.Request.Context(),
		&auth_service.SmsOtpSettingsPrimaryKey{
			Id: id,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.OK, res)
}

// DeleteSmsOtpSettings godoc
// @ID delete_sms_otp_settings
// @Router /v2/sms-otp-settings/{id} [DELETE]
// @Summary Delete sms otp settings
// @Description Delete sms otp settings
// @Tags V2_SMS_OTP_SETTINGS
// @Accept json
// @Produce json
// @Param project-id query string false "project-id"
// @Param X-API-KEY header string false "X-API-KEY"
// @Param Environment-Id header string false "Environment-Id"
// @Param id path string true "id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) DeleteSmsOtpSettings(c *gin.Context) {

	res, err := h.services.ApiKeysService().Delete(
		c.Request.Context(),
		&auth_service.DeleteReq{Id: c.Param("id")},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.NoContent, res)
}
