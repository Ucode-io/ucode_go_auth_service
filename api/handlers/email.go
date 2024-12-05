package handlers

import (
	"errors"
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	cfg "ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	obs "ucode/ucode_go_auth_service/genproto/company_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	pbSms "ucode/ucode_go_auth_service/genproto/sms_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SendMessageToEmail godoc
// @ID send_message_to_email
// @Router /v2/send-message [POST]
// @Summary Send Message To Email
// @Description Send Message to Email
// @Tags Email
// @Accept json
// @Produce json
// @Param send_message body models.Email true "SendMessageToEmailRequestBody"
// @Param X-API-KEY header string false "X-API-KEY"
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Success 201 {object} http.Response{data=models.SendCodeResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) SendMessageToEmail(c *gin.Context) {
	var (
		resourceEnvironment *obs.ResourceEnvironment
		request             models.Email
		respObject          *pbObject.V2LoginResponse
		phone               string
		text                string
	)

	if err := c.ShouldBindJSON(&request); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if request.RegisterType == "" {
		h.handleResponse(c, http.BadRequest, "Must be register type(default, google, phone)")
		return
	}

	id, err := uuid.NewRandom()
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	if request.Email != "" {
		valid := util.IsValidEmail(request.Email)
		if !valid {
			h.handleResponse(c, http.BadRequest, "Неверная почта")
			return
		}
	}

	if request.Phone != "" {
		valid := util.IsValidPhone(request.Phone)
		if !valid {
			h.handleResponse(c, http.BadRequest, "Неверный номер телефона, он должен содержать двенадцать цифр и +")
			return
		}

		phone = helper.ConverPhoneNumberToMongoPhoneFormat(request.Phone)
	}

	expire := time.Now().Add(time.Hour * 5).Add(time.Minute * 5) // hard code time zone

	code, err := util.GenerateCode(6)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	resourceId, ok := c.Get("resource_id")
	if !ok || !util.IsValidUUID(resourceId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id").Error())
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id").Error())
		return
	}

	resourceEnvironment, err = h.services.ResourceService().GetResourceEnvironment(
		c.Request.Context(),
		&obs.GetResourceEnvironmentReq{
			EnvironmentId: environmentId.(string),
			ResourceId:    resourceId.(string),
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	services, err := h.GetProjectSrvc(c, resourceEnvironment.ProjectId, resourceEnvironment.NodeType)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	switch request.RegisterType {
	case cfg.Default:
		{
			respObject, err = services.GetLoginServiceByType(resourceEnvironment.NodeType).LoginWithEmailOtp(
				c.Request.Context(),
				&pbObject.EmailOtpRequest{
					ClientType: "WEB_USER",
					TableSlug:  "user",
					Email:      request.Email,
					ProjectId:  resourceEnvironment.GetId(), //@TODO:: temp added hardcoded project id
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}

			resp, err := h.services.EmailService().Create(
				c.Request.Context(),
				&pb.Email{
					Id:        id.String(),
					Email:     request.Email,
					Otp:       code,
					ExpiresAt: expire.String()[:19],
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}

			emailSettings, err := h.services.EmailService().GetListEmailSettings(
				c.Request.Context(),
				&pb.GetListEmailSettingsRequest{
					ProjectId: resourceEnvironment.GetProjectId(),
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}

			if len(emailSettings.Items) == 0 {
				h.handleResponse(c, http.GRPCError, "No email settings for send otp message in project")
				return
			}

			err = helper.SendCodeToEmail("Your verification code", request.Email, code, emailSettings.Items[0].Email, emailSettings.Items[0].Password)
			if err != nil {
				h.handleResponse(c, http.InvalidArgument, err.Error())
				return
			}

			if respObject == nil || !respObject.UserFound {
				res := models.SendCodeResponse{
					SmsId: resp.Id,
					Data: &pbObject.V2LoginResponse{
						UserFound: false,
					},
					GoogleAcces: false,
				}

				h.handleResponse(c, http.Created, res)
				return
			}

			res := models.SendCodeResponse{
				SmsId:       resp.Id,
				Data:        respObject,
				GoogleAcces: false,
			}

			h.handleResponse(c, http.Created, res)
			return
		}
	case cfg.WithPhone:
		{
			if request.Phone == "" {
				h.handleResponse(c, http.GRPCError, "Phone required when register type is phone")
				return
			}
			respObject, err = services.GetLoginServiceByType(resourceEnvironment.NodeType).LoginWithOtp(
				c.Request.Context(),
				&pbObject.PhoneOtpRequst{
					PhoneNumber: phone,
					ClientType:  request.ClientType,
					ProjectId:   resourceEnvironment.GetId(),
				})
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}
			if request.Text == "" {
				text = "Your one time password, don't get it to anyone: "
			} else {
				text = request.Text
			}
			resp, err := services.SmsService().Send(
				c.Request.Context(),
				&pbSms.Sms{
					Id:        id.String(),
					Text:      text,
					Otp:       code,
					Recipient: request.Phone,
					ExpiresAt: expire.String()[:19],
					Type:      request.RegisterType,
					// PhoneNumber: request.Phone,
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}
			if respObject == nil || !respObject.UserFound {
				res := models.SendCodeResponse{
					SmsId: resp.SmsId,
					Data: &pbObject.V2LoginResponse{
						UserFound: false,
					},
					GoogleAcces: false,
				}

				h.handleResponse(c, http.Created, res)
				return
			}

			res := models.SendCodeResponse{
				SmsId:       resp.SmsId,
				Data:        respObject,
				GoogleAcces: false,
			}

			h.handleResponse(c, http.Created, res)
			return
		}
	case cfg.WithGoogle:
		{
			if request.GoogleToken == "" {
				h.handleResponse(c, http.BadRequest, "google token is required when register type is google")
				return
			}

			userInfo, err := helper.GetGoogleUserInfo(request.GoogleToken)
			if err != nil {
				h.handleResponse(c, http.BadRequest, "Invalid arguments google auth")
				return
			}
			if userInfo["error"] != nil || !(userInfo["email_verified"].(bool)) {
				h.handleResponse(c, http.BadRequest, "Invalid google access token")
				return
			}

			request.Email = userInfo["email"].(string)

			respObject, err = services.GetLoginServiceByType(resourceEnvironment.NodeType).LoginWithEmailOtp(
				c.Request.Context(),
				&pbObject.EmailOtpRequest{
					ClientType: "WEB_USER",
					TableSlug:  "user",
					Email:      request.Email,
					ProjectId:  resourceEnvironment.GetId(), //@TODO:: temp added hardcoded project id
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}

			if respObject == nil || !respObject.UserFound {
				res := models.SendCodeResponse{
					SmsId: "",
					Data: &pbObject.V2LoginResponse{
						UserFound: false,
					},
					GoogleAcces: true,
				}

				h.handleResponse(c, http.Created, res)
				return
			}

			res := models.SendCodeResponse{
				SmsId:       "",
				GoogleAcces: true,
				Data:        respObject,
			}

			h.handleResponse(c, http.Created, res)
			return
		}
	}

	h.handleResponse(c, http.GRPCError, "Register type must be default or phone type")
}

// CreateEmailSettings godoc
// @ID createEmailSettings
// @Router /v2/email-settings [POST]
// @Summary CreateEmailSettings
// @Description CreateEmailSettings
// @Tags Email
// @Accept json
// @Produce json
// @Param X-API-KEY header string false "X-API-KEY"
// @Param registerBody body auth_service.EmailSettings true "register_body"
// @Success 201 {object} http.Response{data=auth_service.EmailSettings} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateEmailSettings(c *gin.Context) {
	var body *pb.EmailSettings

	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if !util.IsValidEmail(body.Email) {
		h.handleResponse(c, http.BadRequest, "Неверная почта")
		return
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	resp, err := h.services.EmailService().CreateEmailSettings(
		c.Request.Context(),
		&pb.EmailSettings{
			Id:        uuid.String(),
			ProjectId: body.ProjectId,
			Email:     body.Email,
			Password:  body.Password,
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// UpdateEmailSettings godoc
// @ID updateEmailSettings
// @Router /v2/email-settings [PUT]
// @Summary UpdateEmailSettings
// @Description UpdateEmailSettings
// @Tags Email
// @Accept json
// @Produce json
// @Param registerBody body auth_service.UpdateEmailSettingsRequest true "register_body"
// @Success 201 {object} http.Response{data=auth_service.EmailSettings} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateEmailSettings(c *gin.Context) {
	var body *pb.UpdateEmailSettingsRequest

	if err := c.ShouldBindJSON(&body); err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.EmailService().UpdateEmailSettings(
		c.Request.Context(), &pb.UpdateEmailSettingsRequest{
			Id:       body.Id,
			Email:    body.Email,
			Password: body.Password,
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// GetListEmailSettings godoc
// @ID getListEmailSettings
// @Router /v2/email-settings [GET]
// @Summary GetListEmailSettings
// @Description GetListEmailSettings
// @Tags Email
// @Accept json
// @Produce json
// @Param project_id query string true "project_id"
// @Success 201 {object} http.Response{data=auth_service.UpdateEmailSettingsResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetEmailSettings(c *gin.Context) {
	resp, err := h.services.EmailService().GetListEmailSettings(
		c.Request.Context(), &pb.GetListEmailSettingsRequest{
			ProjectId: c.Query("project_id"),
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}

// DeleteEmailSettings godoc
// @ID deleteEmailSettings
// @Router /v2/email-settings/{id} [DELETE]
// @Summary DeleteEmailSettings
// @Description DeleteEmailSettings
// @Tags Email
// @Accept json
// @Produce json
// @Param id path string true "id"
// @Success 204
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) DeleteEmailSettings(c *gin.Context) {
	resp, err := h.services.EmailService().DeleteEmailSettings(
		c.Request.Context(), &pb.EmailSettingsPrimaryKey{
			Id: c.Param("id"),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, resp)
}
