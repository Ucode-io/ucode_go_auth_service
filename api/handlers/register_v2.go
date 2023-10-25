package handlers

import (
	"context"
	"errors"
	"fmt"
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	cfg "ucode/ucode_go_auth_service/config"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	obs "ucode/ucode_go_auth_service/genproto/company_service"
	pbCompany "ucode/ucode_go_auth_service/genproto/company_service"
	pbSms "ucode/ucode_go_auth_service/genproto/sms_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// V2SendCodeApp godoc
// @ID V2SendCodeApp
// @Router /v2/send-code-app [POST]
// @Summary SendCodeApp
// @Description SendCode type must be one of the following values ["EMAIL", "PHONE"]
// @Tags v2_register
// @Accept json
// @Produce json
// @Param login body models.V2SendCodeRequest true "SendCode"
// @Success 201 {object} http.Response{data=models.V2SendCodeResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2SendCodeApp(c *gin.Context) {

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

	resourceEnvironment, err := h.services.ResourceService().GetResourceEnvironment(
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
		fmt.Println("test 10")
		smsOtpSettings, err := h.services.SmsOtpSettingsService().GetList(context.Background(), &auth_service.GetListSmsOtpSettingsRequest{
			ProjectId:     resourceEnvironment.ProjectId,
			EnvironmentId: environmentId.(string),
		})
		if err != nil {
			h.handleResponse(c, http.GRPCError, err.Error())
			return
		}
		fmt.Println("test 11")
		if len(smsOtpSettings.GetItems()) > 0 {
			if smsOtpSettings.GetItems()[0].GetNumberOfOtp() != 0 {
				code, err := util.GenerateCode(int(smsOtpSettings.GetItems()[0].GetNumberOfOtp()))
				if err != nil {
					h.handleResponse(c, http.InvalidArgument, "invalid number of otp")
					return
				}
				body.Otp = code
			}
			body.DevEmail = smsOtpSettings.GetItems()[0].Login
			body.DevEmailPassword = smsOtpSettings.GetItems()[0].Password
			body.Originator = smsOtpSettings.GetItems()[0].Originator
		}
	case "EMAIL":

		valid = util.IsValidEmail(request.Recipient)
		if !valid {
			h.handleResponse(c, http.BadRequest, "Email is not valid")
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
// @Description client_type_id and role_id must be in body parameters
// @Description you must be give environment_id and project_id in body or
// @Description Environment-Id hearder and project-id in query parameters or
// @Description X-API-KEY in hearder
// @Tags v2_register
// @Accept json
// @Produce json
// @Param X-API-KEY header string false "X-API-KEY"
// @Param Resource-Id header string false "Resource-Id"
// @Param Environment-Id header string false "Environment-Id"
// @Param project-id query string false "project-id"
// @Param registerBody body models.RegisterOtp true "register_body"
// @Success 201 {object} http.Response{data=pb.V2LoginResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2Register(c *gin.Context) {
	var (
		body models.RegisterOtp
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
	if _, ok := body.Data["client_type_id"].(string); !ok {
		if !util.IsValidUUID(body.Data["client_type_id"].(string)) {
			h.handleResponse(c, http.BadRequest, "client_type_id is an invalid uuid")
			return
		}
		h.handleResponse(c, http.BadRequest, "client_type_id is required")
		return
	}
	if _, ok := body.Data["role_id"].(string); !ok {
		if !util.IsValidUUID(body.Data["role_id"].(string)) {
			h.handleResponse(c, http.BadRequest, "role_id is an invalid uuid")
			return
		}
		h.handleResponse(c, http.BadRequest, "role_id is required")
		return
	}
	projectId, ok := c.Get("project_id")
	if !ok || !util.IsValidUUID(projectId.(string)) {

		h.handleResponse(c, http.BadRequest, "cant get project_id")
		return
	}

	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}

	serviceResource, err := h.services.ServiceResource().GetSingle(
		c.Request.Context(),
		&obs.GetSingleServiceResourceReq{
			EnvironmentId: environmentId.(string),
			ProjectId:     projectId.(string),
			ServiceType:   pbCompany.ServiceType_BUILDER_SERVICE,
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	project, err := h.services.ProjectServiceClient().GetById(context.Background(), &company_service.GetProjectByIdRequest{
		ProjectId: serviceResource.GetProjectId(),
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

	body.Data["project_id"] = serviceResource.GetProjectId()
	body.Data["environment_id"] = serviceResource.GetEnvironmentId()
	body.Data["resource_environment_id"] = serviceResource.GetResourceEnvironmentId()
	body.Data["environment_id"] = serviceResource.GetEnvironmentId()
	body.Data["company_id"] = project.GetCompanyId()
	body.Data["resource_type"] = serviceResource.GetResourceType()

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
