package handlers

import (
	"strings"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//var (
		//	res *auth_service.V2HasAccessUserRes
		//	ok  bool
		//)
		//host := c.Request.Host
		//
		//if strings.Contains(host, CLIENT_HOST) {
		//	res, ok = h.hasAccess(c)
		//	if !ok {
		//		c.Abort()
		//		return
		//	}
		//}
		resourceId := c.GetHeader("Resource-Id")
		environmentId := c.GetHeader("Environment-Id")
		projectId := c.DefaultQuery("project-id", "")
		bearerToken := c.GetHeader("Authorization")
		if bearerToken != "" {
			strArr := strings.Split(bearerToken, " ")

			if strArr[0] == "API-KEY" {
				app_id := c.GetHeader("X-API-KEY")
				apikeys, err := h.services.ApiKeysService().GetEnvID(
					c.Request.Context(),
					&auth_service.GetReq{
						Id: app_id,
					},
				)
				if err != nil {
					h.handleResponse(c, http.BadRequest, err.Error())
					c.Abort()
					return
				}

				resource, err := h.services.ResourceService().GetResourceByEnvID(
					c.Request.Context(),
					&company_service.GetResourceByEnvIDRequest{
						EnvId: apikeys.GetEnvironmentId(),
					},
				)
				if err != nil {
					h.handleResponse(c, http.BadRequest, err.Error())
					c.Abort()
					return
				}

				resourceId = resource.GetResource().GetId()
				environmentId = apikeys.GetEnvironmentId()
				projectId = apikeys.GetProjectId()
			}
		}

		//c.Set("Auth", res)
		c.Set("resource_id", resourceId)
		c.Set("environment_id", environmentId)
		c.Set("project_id", projectId)
		//c.Set("namespace", h.cfg.UcodeNamespace)
		c.Next()
	}
}
