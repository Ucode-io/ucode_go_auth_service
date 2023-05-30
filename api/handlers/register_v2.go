package handlers

import (
	"context"
	"errors"
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	cfg "ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	obs "ucode/ucode_go_auth_service/genproto/company_service"
	pbSms "ucode/ucode_go_auth_service/genproto/sms_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// V2SendCode godoc
// @ID V2SendCode
// @Router /v2/send-code [POST]
// @Summary SendCode
// @Description SendCode type must be one of the following values ["EMAIL", "PHONE"]
// @Tags v2_register
// @Accept json
// @Produce json
// @Param X-API-KEY header string false "X-API-KEY"
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string false "Environment-Id"
// @Param login body models.V2SendCodeRequest true "SendCode"
// @Success 201 {object} http.Response{data=models.V2SendCodeResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2SendCode(c *gin.Context) {

	var (
		request models.V2SendCodeRequest
	)

	err := c.ShouldBindJSON(&request)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	id, err := uuid.NewRandom()
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	_, valid := util.ValidRecipients[request.Type]
	if !valid {
		h.handleResponse(c, http.BadRequest, "Invalid recipient type")
		return
	}
	resourceId, ok := c.Get("resource_id")
	if !ok {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id").Error())
		return
	}
	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id").Error())
		return
	}
	expire := time.Now().Add(time.Minute * 5) // todo dont write expire time here

	code, err := util.GenerateCode(4)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}
	body := &pbSms.Sms{
		Id:        id.String(),
		Text:      request.Text,
		Otp:       code,
		Recipient: request.Recipient,
		ExpiresAt: expire.String()[:19],
		Type:      request.Type,
	}

	switch request.Type {
	case "PHONE":
		valid = util.IsValidPhone(request.Recipient)
		if !valid {
			h.handleResponse(c, http.BadRequest, "Неверный номер телефона, он должен содержать двенадцать цифр и +")
			return
		}
	case "EMAIL":
		valid = util.IsValidEmail(request.Recipient)
		if !valid {
			h.handleResponse(c, http.BadRequest, "Email is not valid")
			return
		}
		if !util.IsValidUUID(resourceId.(string)) {
			h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id").Error())
			return
		}
		resourceEnvironment, err := h.services.ResourceService().GetResourceEnvironment(
			c.Request.Context(),
			&obs.GetResourceEnvironmentReq{
				EnvironmentId: environmentId.(string),
				ResourceId:    resourceId.(string),
			},
		)
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
		if len(emailSettings.Items) < 1 {
			h.handleResponse(c, http.InvalidArgument, errors.New("email settings not found"))
			return
		}
		body.DevEmail = emailSettings.Items[0].Email
		body.DevEmailPassword = emailSettings.Items[0].Password
	}
	_, err = h.services.UserService().V2GetUserByLoginTypes(c.Request.Context(), &auth_service.GetUserByLoginTypesRequest{
		Email: request.Recipient,
		Phone: request.Recipient,
	})

	resp, err := h.services.SmsService().Send(
		c.Request.Context(),
		body,
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}
	res := models.V2SendCodeResponse{
		SmsId: resp.SmsId,
	}

	h.handleResponse(c, http.Created, res)
}

// V2Register godoc
// @ID V2register
// @Router /v2/register [POST]
// @Summary V2Register
// @Description V2Register
// @Description in data must be have type, type must be one of the following values
// @Description ["google", "apple", "email", "phone"]
// @Tags v2_register
// @Accept json
// @Produce json
// @Param registerBody body models.RegisterOtp true "register_body"
// @Param X-API-KEY header string false "X-API-KEY"
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string false "Environment-Id"
// @Success 201 {object} http.Response{data=pb.V2LoginResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2Register(c *gin.Context) {
	var (
		body                models.RegisterOtp
		resourceEnvironment *obs.ResourceEnvironment
	)

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	if _, ok := body.Data["type"]; !ok {
		h.handleResponse(c, http.BadRequest, "register type is required")
		return
	}
	if _, ok := cfg.RegisterTypes[body.Data["type"].(string)]; !ok {
		h.handleResponse(c, http.BadRequest, "invalid register type")
		return
	}

	resourceId, ok := c.Get("resource_id")
	if !ok || !util.IsValidUUID(resourceId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
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

	project, err := h.services.ProjectServiceClient().GetById(context.Background(), &company_service.GetProjectByIdRequest{
		ProjectId: resourceEnvironment.GetProjectId(),
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	switch body.Data["type"] {
	case cfg.WithGoogle:
		{
			h.handleResponse(c, http.BadRequest, "register with goole not implemented")
			return

		}
	case cfg.WithApple:
		{
			h.handleResponse(c, http.BadRequest, "registre with apple not implemented")
			return
		}
	case cfg.WithEmail:
		{
			if v, ok := body.Data["email"]; ok {
				if !util.IsValidEmail(v.(string)) {
					h.handleResponse(c, http.BadRequest, "Неверный формат email")
					return
				}
			} else {
				h.handleResponse(c, http.BadRequest, "Поле email не заполнено")
				return
			}

			if _, ok := body.Data["login"]; !ok {
				h.handleResponse(c, http.BadRequest, "Поле login не заполнено")
				return
			}

			if _, ok := body.Data["name"]; !ok {
				h.handleResponse(c, http.BadRequest, "Поле name не заполнено")
				return
			}

			if _, ok := body.Data["phone"]; !ok {
				h.handleResponse(c, http.BadRequest, "Поле phone не заполнено")
				return
			}
		}
	case cfg.WithPhone:
		{
			if _, ok := body.Data["phone"]; !ok {
				h.handleResponse(c, http.BadRequest, "Поле phone не заполнено")
				return

			}
		}
	}
	if body.Data["addational_table"] != nil {
		if body.Data["addational_table"].(map[string]interface{})["table_slug"] == nil {
			h.log.Error("Addational user create >>>> ")
			h.handleResponse(c, http.BadRequest, "If addional table have, table slug is required")
			return
		}
	}
	body.Data["project_id"] = resourceEnvironment.GetProjectId()
	body.Data["resource_environment_id"] = resourceEnvironment.GetId()
	body.Data["environment_id"] = resourceEnvironment.GetEnvironmentId()
	body.Data["company_id"] = project.GetCompanyId()
	body.Data["resource_type"] = resourceEnvironment.GetResourceType()

	structData, err := helper.ConvertMapToStruct(body.Data)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	response, err := h.services.RegisterService().RegisterUser(c.Request.Context(), &auth_service.RegisterUserRequest{
		Data: structData,
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, response)
}
