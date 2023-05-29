package handlers

import (
	"time"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/api/models"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	pbSms "ucode/ucode_go_auth_service/genproto/sms_service"
	"ucode/ucode_go_auth_service/pkg/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// V2SendCode godoc
// @ID v2SendCode
// @Router /v2/send-code [POST]
// @Summary SendCode
// @Description SendCode type must be one of the following values ["EMAIL", "PHONE"]
// @Tags v2_register
// @Accept json
// @Produce json
// @Param login body models.V2SendCodeRequest true "SendCode"
// @Success 201 {object} http.Response{data=models.V2SendCodeResponse} "User data"
// @Response 400 {object} http.Response{data=string} "Bad Request"
// @Failure 500 {object} http.Response{data=string} "Server Error"
func (h *Handler) V2SendCode(c *gin.Context) {

	var request models.V2SendCodeRequest

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

	expire := time.Now().Add(time.Minute * 5) // todo dont write expire time here

	code, err := util.GenerateCode(4)
	if err != nil {
		h.handleResponse(c, http.InternalServerError, err.Error())
		return
	}

	resp, err := h.services.SmsService().Send(
		c.Request.Context(),
		&pbSms.Sms{
			Id:        id.String(),
			Text:      request.Text,
			Otp:       code,
			Recipient: request.Recipient,
			ExpiresAt: expire.String()[:19],
			Type:      request.Type,
		},
	)
	if err != nil {
		h.handleResponse(c, http.GRPCError, err.Error())
		return
	}
	res := models.SendCodeResponse{
		SmsId: resp.SmsId,
	}

	h.handleResponse(c, http.Created, res)
}
