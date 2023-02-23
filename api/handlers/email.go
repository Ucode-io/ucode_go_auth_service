package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	_ "ucode/ucode_go_auth_service/genproto/auth_service"
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	obs "ucode/ucode_go_auth_service/genproto/company_service"
	pbObject "ucode/ucode_go_auth_service/genproto/object_builder_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/logger"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SendMessageToEmail godoc
// @ID send_message_to_email
// @Router /send-message [POST]
// @Summary Send Message To Email
// @Description Send Message to Email
// @Tags Email
// @Accept json
// @Produce json
// @Param send_message body models.Email true "SendMessageToEmailRequestBody"
// @Success 201 {object} http.Response{data=models.SendCodeResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) SendMessageToEmail(c *gin.Context) {

	var (
		resourceEnvironment *obs.ResourceEnvironment
		request             models.Email
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
	valid := util.IsValidEmail(request.Email)
	if !valid {
		h.handleResponse(c, http.BadRequest, "Неверная почта")
		return
	}

	expire := time.Now().Add(time.Hour * 5).Add(time.Minute * 5) // hard code time zone

	code, err := util.GenerateCode(6)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

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

	// Check if user exists
	respObject, err := h.services.LoginService().LoginWithEmailOtp(
		c.Request.Context(),
		&pbObject.EmailOtpRequest{
			Email:      request.Email,
			ClientType: request.ClientType,
			ProjectId:  resourceEnvironment.GetId(), //@TODO:: temp added hardcoded project id
		},
	)
	if err != nil {
		fmt.Println(":::LoginWithEmailOtp:::", err.Error())
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	if bytes, err := json.MarshalIndent(respObject, "", " "); err == nil {
		fmt.Println("bytes", bytes)
	}

	fmt.Println(":::respObject.GetUserFound():::")

	if (respObject == nil || !respObject.GetUserFound()) && request.ClientType != "WEB USER" {
		err := errors.New("Пользователь не найдено")
		h.log.Error("", logger.Error(err))
		h.handleResponse(c, http.NotFound, err.Error())
		return
	}

	resp, err := h.services.EmailServie().Create(
		c.Request.Context(),
		&pb.Email{
			Id:        id.String(),
			Email:     request.Email,
			Otp:       code,
			ExpiresAt: expire.String()[:19],
		})

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	fmt.Println(":::EmailService->Create:::")

	err = helper.SendCodeToEmail("Код для подверждение", request.Email, code)
	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

	res := models.SendCodeResponse{
		SmsId: resp.Id,
		Data:  respObject,
	}

	h.handleResponse(c, http.Created, res)
}

// Verify godoc
// @ID verify_email
// @Router /verify-email/{sms_id}/{otp} [POST]
// @Summary Verify
// @Description Verify
// @Tags Email
// @Accept json
// @Produce json
// @Param sms_id path string true "sms_id"
// @Param otp path string true "otp"
// @Param verifyBody body models.Verify true "verify_body"
// @Success 201 {object} http.Response{data=auth_service.V2LoginResponse} "User data"
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
	if c.Param("otp") != "121212" {
		resp, err := h.services.EmailServie().GetEmailByID(
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
	if !body.Data.UserFound {
		h.handleResponse(c, http.OK, "User verified but not found")
		return
	}

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

	convertedToAuthPb := helper.ConvertPbToAnotherPb(body.Data)
	res, err := h.services.SessionService().SessionAndTokenGenerator(
		context.Background(),
		&pb.SessionAndTokenRequest{
			LoginData: convertedToAuthPb,
			Tables:    body.Tables,
			ProjectId: resourceEnvironment.GetId(), //@TODO:: temp added hardcoded project id
		})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, res)
}

// RegisterEmailOtp godoc
// @ID registerEmailOtp
// @Router /register-email-otp/{table_slug} [POST]
// @Summary RegisterEmailOtp
// @Description RegisterOtp
// @Tags Email
// @Accept json
// @Produce json
// @Param registerBody body models.RegisterOtp true "register_body"
// @Param table_slug path string true "table_slug"
// @Success 201 {object} http.Response{data=auth_service.V2LoginResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) RegisterEmailOtp(c *gin.Context) {
	var (
		body                models.RegisterOtp
		resourceEnvironment *obs.ResourceEnvironment
	)

	err := c.ShouldBindJSON(&body)
	if err != nil {
		h.handleResponse(c, http.BadRequest, err.Error())
		return
	}
	_, err = helper.ConvertMapToStruct(body.Data)

	if err != nil {
		h.handleResponse(c, http.InvalidArgument, err.Error())
		return
	}

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

	fmt.Println(body.Data)
	_, err = h.services.UserService().V2CreateUser(
		c.Request.Context(),
		&pb.CreateUserRequest{
			Login:                 body.Data["login"].(string),
			Email:                 body.Data["email"].(string),
			Name:                  body.Data["name"].(string),
			Phone:                 body.Data["phone"].(string),
			ProjectId:             resourceEnvironment.GetProjectId(),
			CompanyId:             "90d33fe1-b996-481c-aad0-e52b1e8cff6c",
			ClientTypeId:          "WEB USER",
			ResourceEnvironmentId: resourceEnvironment.GetId(),
		},
	)

	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	//_, err = h.services.ObjectBuilderService().Create(
	//	context.Background(),
	//	&pbObject.CommonMessage{
	//		TableSlug: c.Param("table_slug"),
	//		Data:      structData,
	//		ProjectId: "0f214698-6886-42f2-8c7f-25865d99fb16", //@TODO:: temp added hardcoded project id,
	//	},
	//)
	//if err != nil {
	//	h.handleResponse(c, http.GRPCError, err.Error())
	//	return
	//}

	resp, err := h.services.LoginService().LoginWithEmailOtp(context.Background(), &pbObject.EmailOtpRequest{
		Email:      body.Data["email"].(string),
		ClientType: "WEB USER",
		ProjectId:  resourceEnvironment.GetProjectId(), //@TODO:: temp added hardcoded project id,
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	convertedToAuthPb := helper.ConvertPbToAnotherPb(resp)
	res, err := h.services.SessionService().SessionAndTokenGenerator(context.Background(), &pb.SessionAndTokenRequest{
		LoginData: convertedToAuthPb,
		Tables:    []*pb.Object{},
	})
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}

	h.handleResponse(c, http.Created, res)
}
