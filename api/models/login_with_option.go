package models

import "ucode/ucode_go_auth_service/genproto/auth_service"

type LoginMiddlewareReq struct {
	Data   map[string]string
	Tables []*auth_service.Object
}
