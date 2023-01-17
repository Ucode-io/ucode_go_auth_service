package handlers

import (
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

		//c.Set("Auth", res)
		c.Set("resource_id", resourceId)
		c.Set("environment_id", environmentId)
		//c.Set("namespace", h.cfg.UcodeNamespace)
		c.Next()
	}
}
