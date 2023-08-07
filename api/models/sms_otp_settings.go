package models

type CreateSmsOtpSettingsRequest struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	NumberOfOtp int32  `json:"number_of_otp"`
	DefaultOtp  string `json:"default_otp"`
}

type UpdateSmsOtpSettingsRequest struct {
	Id          string `json:"id"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	NumberOfOtp int32  `json:"number_of_otp"`
	DefaultOtp  string `json:"default_otp"`
}
