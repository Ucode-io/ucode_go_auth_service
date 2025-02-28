package models

import (
	pb "ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/object_builder_service"
)

type Sms struct {
	Text       string `json:"text"`
	Recipient  string `json:"recipient"`
	ClientType string `json:"client_type"`
}

type SendCodeResponse struct {
	SmsId       string                                  `json:"sms_id"`
	GoogleAcces bool                                    `json:"google_acces"`
	Data        *object_builder_service.V2LoginResponse `json:"data"`
}

type Verify struct {
	Data         *object_builder_service.V2LoginResponse `json:"data"`
	Tables       []*pb.Object                            `json:"tables"`
	Provider     string                                  `json:"provider"`
	GoogleToken  string                                  `json:"google_token"`
	AppleCode    string                                  `json:"apple_code"`
	Otp          string                                  `json:"otp"`
	RegisterType string                                  `json:"register_type"`
}

type RegisterOtp struct {
	Data map[string]any `json:"data"`
}

type VerifyEmail struct {
	RegisterType string `json:"register_type"`
	SmsId        string `json:"sms_id"`
	Otp          string `json:"otp"`
}

type Email struct {
	Email        string `json:"email"`
	ClientType   string `json:"client_type"`
	RegisterType string `json:"register_type"`
	GoogleToken  string `json:"google_token"`
	Phone        string `json:"phone"`
	Text         string `json:"text"`
}

type EmailSettingsRequest struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	ProjectId string `json:"project_id"`
}

type V2SendCodeRequest struct {
	Text          string `json:"text"`
	Recipient     string `json:"recipient"`
	Type          string `json:"type"`
	SmsTemplateId string `json:"sms_template_id"`
}

type V2SendCodeResponse struct {
	SmsId       string `json:"sms_id"`
	GoogleAcces bool   `json:"google_acces"`
	UserFound   bool   `json:"user_found"`
}
