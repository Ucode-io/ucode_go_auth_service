package models

import "ucode/ucode_go_auth_service/genproto/auth_service"

type LoginMiddlewareReq struct {
	Data      map[string]string      `json:"data"`
	Tables    []*auth_service.Object `json:"tables"`
	NodeType  string                 `json:"node_type"`
	ClientId  string                 `json:"client_id"`
	ClientIp  string                 `json:"client_ip"`
	UserAgent string                 `json:"user_agent"`
}
