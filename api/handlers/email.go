package handlers

import (
	"context"
	"encoding/json"
	_ "encoding/json"
	"errors"
	"fmt"
	"log"
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	cfg "ucode/ucode_go_auth_service/config"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	obs "ucode/ucode_go_auth_service/genproto/company_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	pbSms "ucode/ucode_go_auth_service/genproto/sms_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/logger"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/structpb"
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

	log.Println("--- SendMessageToEmail ---")

	var (
		resourceEnvironment *obs.ResourceEnvironment
		request             models.Email
		respObject          *pbObject.V2LoginResponse
		phone               string
	)

	err := c.ShouldBindJSON(&request)
	if err != nil {
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
	if !ok {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id").Error())
		return
	}
	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id").Error())
		return
	}
	if !util.IsValidUUID(resourceId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id").Error())
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
	fmt.Println(":::::::: Register type :", request.RegisterType)
	switch request.RegisterType {
	case cfg.Default:
		{
			respObject, err = h.services.LoginService().LoginWithEmailOtp(
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

			err = helper.SendCodeToEmail("Код для подтверждения", request.Email, code, emailSettings.Items[0].Email, emailSettings.Items[0].Password)
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
			respObject, err = h.services.LoginService().LoginWithOtp(
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
			fmt.Println("::::::: LoginWith O response :", respObject)
			resp, err := h.services.SmsService().Send(
				c.Request.Context(),
				&pbSms.Sms{
					Id:          id.String(),
					Text:        "Your one time password, don't get it to anyone: ",
					Otp:         code,
					Recipient:   request.Phone,
					ExpiresAt:   expire.String()[:19],
					PhoneNumber: request.Phone,
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}
			fmt.Println("::::::: Phone response :", resp)
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

			respObject, err = h.services.LoginService().LoginWithEmailOtp(
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
	return
}

// Verify godoc
// @ID verify_email
// @Router /v2/verify-email/{sms_id}/{otp} [POST]
// @Summary Verify
// @Description Verify
// @Tags Email
// @Accept json
// @Produce json
// @Param sms_id path string true "sms_id"
// @Param otp path string true "otp"
// @Param X-API-KEY header string false "X-API-KEY"
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Param verifyBody body models.Verify true "verify_body"
// @Success 201 {object} http.Response{data=pb.V2LoginResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) VerifyEmail(c *gin.Context) {
	var (
		body                models.Verify
		resourceEnvironment *obs.ResourceEnvironment
	)

	err := c.ShouldBindJSON(&body)

	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if body.RegisterType == "" {
		h.handleResponse(c, http.BadRequest, "Register type is required")
		return
	}
	fmt.Println("::::::: Register type body :", body.RegisterType)

	resourceId, ok := c.Get("resource_id")
	if !ok {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
		return
	}
	environmentId, ok := c.Get("environment_id")
	if !ok || !util.IsValidUUID(environmentId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get environment_id"))
		return
	}
	if !util.IsValidUUID(resourceId.(string)) {
		h.handleResponse(c, http.BadRequest, errors.New("cant get resource_id"))
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

	switch body.RegisterType {
	case cfg.Default:
		{
			if c.Param("otp") != "121212" {
				resp, err := h.services.EmailService().GetEmailByID(
					c.Request.Context(),
					&pb.EmailOtpPrimaryKey{
						Id: c.Param("sms_id"),
					},
				)
				if err != nil {
					h.handleResponse(c, http.GRPCError, err.Error())
					return
				}
				if resp.Otp != c.Param("otp") {
					h.handleResponse(c, http.InvalidArgument, "Неверный код подверждения")
					return
				}
			}
		}
	case cfg.WithPhone:
		{
			fmt.Println("::::::: Register type with phone :", body.RegisterType)
			if c.Param("otp") != "121212" {
				_, err := h.services.SmsService().ConfirmOtp(
					c.Request.Context(),
					&pbSms.ConfirmOtpRequest{
						SmsId: c.Param("sms_id"),
						Otp:   c.Param("otp"),
					},
				)
				if err != nil {
					h.handleResponse(c, http.GRPCError, err.Error())
					return
				}
			}
		}
	case cfg.WithGoogle:
		{
			if body.GoogleToken == "" {
				h.handleResponse(c, http.BadRequest, "google token is required when register type is google")
				return
			}

			userInfo, err := helper.GetGoogleUserInfo(body.GoogleToken)
			if err != nil {
				h.handleResponse(c, http.BadRequest, "Invalid arguments google auth")
				return
			}
			if userInfo["error"] != nil || !(userInfo["email_verified"].(bool)) {
				h.handleResponse(c, http.BadRequest, "Invalid google access token")
				return
			}

			respObject, err := h.services.LoginService().LoginWithEmailOtp(
				c.Request.Context(),
				&pbObject.EmailOtpRequest{
					ClientType: "WEB_USER",
					TableSlug:  "user",
					Email:      userInfo["email"].(string),
					ProjectId:  resourceEnvironment.GetId(),
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}

			if respObject == nil || !respObject.UserFound {
				h.handleResponse(c, http.OK, "User verified with google token but not found")
				return
			}

			convertedToAuthPb := helper.ConvertPbToAnotherPb(respObject)
			res, err := h.services.SessionService().SessionAndTokenGenerator(
				context.Background(),
				&pb.SessionAndTokenRequest{
					LoginData: convertedToAuthPb,
					Tables:    body.Tables,
					ProjectId: resourceEnvironment.GetProjectId(), //@TODO:: temp added hardcoded project id
				})
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}

			h.handleResponse(c, http.Created, res)
		}
	}
	if !body.Data.UserFound {
		h.handleResponse(c, http.OK, "User verified but not found")
		return
	}

	convertedToAuthPb := helper.ConvertPbToAnotherPb(body.Data)
	res, err := h.services.SessionService().SessionAndTokenGenerator(
		context.Background(),
		&pb.SessionAndTokenRequest{
			LoginData: convertedToAuthPb,
			Tables:    body.Tables,
			ProjectId: resourceEnvironment.GetProjectId(), //@TODO:: temp added hardcoded project id
		})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, res)
}

// RegisterEmailOtp godoc
// @ID registerEmailOtp
// @Router /v2/register-email-otp/{table_slug} [POST]
// @Summary RegisterEmailOtp
// @Description RegisterOtp
// @Tags Email
// @Accept json
// @Produce json
// @Param registerBody body models.RegisterOtp true "register_body"
// @Param table_slug path string true "table_slug"
// @Param Resource-Id header string true "Resource-Id"
// @Param Environment-Id header string true "Environment-Id"
// @Success 201 {object} http.Response{data=pb.V2LoginResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) RegisterEmailOtp(c *gin.Context) {
	fmt.Println(":::RegisterEmailOtp:::")
	var (
		body                  models.RegisterOtp
		resourceEnvironment   *obs.ResourceEnvironment
		CompanyId             string
		ProjectId             string
		ResourceEnvironmentId string
	)

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	if _, ok := body.Data["register_type"]; !ok {
		h.handleResponse(c, http.BadRequest, "register_type required")
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
	fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>. test 1")
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
	
	ProjectId = resourceEnvironment.GetProjectId()
	ResourceEnvironmentId = resourceEnvironment.GetId()
	CompanyId = project.GetCompanyId()
	
	if body.Data["register_type"] != cfg.WithGoogle {
		body.Data["register_type"] = cfg.Default
	}
	
	if body.Data["phone"] != nil && body.Data["phone"] != "" {
		body.Data["phone"] = helper.ConverPhoneNumberToMongoPhoneFormat(body.Data["phone"].(string))
	}

	var userId string
	
	switch body.Data["register_type"] {
	case cfg.WithGoogle:
		{
			
			if body.Data["google_token"] == nil || body.Data["google_token"] == "" {
				h.handleResponse(c, http.BadRequest, "google_token  required when register_type is google")
				return
			}
			
			userInfo, err := helper.GetGoogleUserInfo(body.Data["google_token"].(string))
			if err != nil {
				h.handleResponse(c, http.BadRequest, "Invalid arguments google auth")
				return
			}
			
			if userInfo["error"] != nil || !(userInfo["email_verified"].(bool)) {
				h.handleResponse(c, http.BadRequest, "Invalid google access token")
				return
			}

			resp, err := h.services.UserService().RegisterWithGoogle(
				c.Request.Context(),
				&pb.RegisterWithGoogleRequest{
					Name:                  userInfo["name"].(string),
					Email:                 userInfo["email"].(string),
					ProjectId:             ProjectId,
					CompanyId:             CompanyId,
					ClientTypeId:          "WEB_USER",
					ResourceEnvironmentId: ResourceEnvironmentId,
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}

			body.Data["email"] = userInfo["email"]
			body.Data["name"] = userInfo["name"]

			userId = resp.Id

		}
	case cfg.Default:
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

			resp, err := h.services.UserService().RegisterUserViaEmail(
				c.Request.Context(),
				&pb.CreateUserRequest{
					Login:                 body.Data["login"].(string),
					Email:                 body.Data["email"].(string),
					Name:                  body.Data["name"].(string),
					Phone:                 body.Data["phone"].(string),
					ProjectId:             ProjectId,
					CompanyId:             CompanyId,
					ClientTypeId:          "WEB_USER",
					ResourceEnvironmentId: ResourceEnvironmentId,
				},
			)
			if err != nil {
				h.handleResponse(c, http.GRPCError, err.Error())
				return
			}

			userId = resp.Id

		}
	}

	resp, err := h.services.LoginService().LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{

		Email:      body.Data["email"].(string),
		ClientType: "WEB_USER",
		ProjectId:  ResourceEnvironmentId,
		TableSlug:  "user",
	})
	if err != nil {
		h.log.Error("---> error in login with email otp", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	if body.Data["addational_table"] != nil {
		if body.Data["addational_table"].(map[string]interface{})["table_slug"] == nil {
			h.log.Error("Addational user create >>>> ")
			h.handleResponse(c, http.GRPCError, "If addional table have, table slug is required")
			return
		}

		if body.Data["register_type"].(string) == cfg.Default {
			// uuid, err := uuid.NewRandom()
			// if err != nil {
			// 	h.handleResponse(c, http.InternalServerError, err.Error())
			// 	return
			// }
			body.Data["addational_table"].(map[string]interface{})["guid"] = userId
			body.Data["addational_table"].(map[string]interface{})["project_id"] = ProjectId

			mapedInterface := body.Data["addational_table"].(map[string]interface{})
			structData, err := helper.ConvertRequestToSturct(mapedInterface)
			if err != nil {
				h.log.Error("Additional table struct table --->", logger.Error(err))
				h.handleResponse(c, http.GRPCError, "Additional table struct table --->")
				return
			}

			respObj, err := h.services.ObjectBuilderService().Create(
				context.Background(),
				&pbObject.CommonMessage{
					TableSlug: mapedInterface["table_slug"].(string),
					Data:      structData,
					ProjectId: ResourceEnvironmentId,
				})
			if err != nil {
				h.log.Error("Object create error >>", logger.Error(err))
				h.handleResponse(c, http.GRPCError, "Object create error >>")
				return
			}

			data := respObj.Data.AsMap()
			var addTable structpb.Struct

			dataJson, err := json.Marshal(data)
			if err != nil {
				return
			}

			err = addTable.UnmarshalJSON(dataJson)
			if err != nil {
				return
			}
			fmt.Println(":::::::::::::::: Addational table", addTable)
			resp.AddationalTable = &addTable
		}

	}
	fmt.Println(":::::::::::::::::: resp resp resp", resp.AddationalTable)
	convertedToAuthPb := helper.ConvertPbToAnotherPb(resp)

	res, err := h.services.SessionService().SessionAndTokenGenerator(context.Background(), &pb.SessionAndTokenRequest{
		LoginData: convertedToAuthPb,
		Tables:    []*pb.Object{},
		ProjectId: ProjectId,
	})

	if err != nil {
		h.log.Error("---> error in session and token generator", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	if resp.AddationalTable != nil {
		res.AddationalTable = resp.AddationalTable
	}

	h.handleResponse(c, http.Created, res)
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
// @Param registerBody body models.EmailSettingsRequest true "register_body"
// @Success 201 {object} http.Response{data=pb.EmailSettings} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) CreateEmailSettings(c *gin.Context) {

	var body *pb.EmailSettings

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	valid := util.IsValidEmail(body.Email)
	if !valid {
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
		h.log.Error("---> error in create email settings", logger.Error(err))
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
// @Param registerBody body pb.UpdateEmailSettingsRequest true "register_body"
// @Success 201 {object} http.Response{data=pb.EmailSettings} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) UpdateEmailSettings(c *gin.Context) {

	var body *pb.UpdateEmailSettingsRequest

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}

	resp, err := h.services.EmailService().UpdateEmailSettings(
		c.Request.Context(),
		&pb.UpdateEmailSettingsRequest{
			Id:       body.Id,
			Email:    body.Email,
			Password: body.Password,
		},
	)
	if err != nil {
		h.log.Error("---> error in update email settings", logger.Error(err))
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
// @Success 201 {object} http.Response{data=pb.UpdateEmailSettingsResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) GetEmailSettings(c *gin.Context) {

	fmt.Println("::::::::; >>>>>>>>>>>> ", c.Query("project_id"))
	resp, err := h.services.EmailService().GetListEmailSettings(
		c.Request.Context(),
		&pb.GetListEmailSettingsRequest{
			ProjectId: c.Query("project_id"),
		},
	)
	if err != nil {
		h.log.Error("---> error in get list email settings", logger.Error(err))
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

	id := c.Param("id")

	resp, err := h.services.EmailService().DeleteEmailSettings(
		c.Request.Context(),
		&pb.EmailSettingsPrimaryKey{
			Id: id,
		},
	)
	fmt.Println(">>>>>>>>>> handler test 1")
	if err != nil {
		h.log.Error("---> error in delete email settings", logger.Error(err))
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	fmt.Println(">>>>>>>>>>> service test 2")
	h.handleResponse(c, http.Created, resp)
}
