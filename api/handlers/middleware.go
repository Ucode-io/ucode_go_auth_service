package handlers

import (
	"errors"
	"fmt"
	nethttp "net/http"
	"strings"
	"ucode/ucode_go_auth_service/api/http"
	"ucode/ucode_go_auth_service/genproto/auth_service"
	"ucode/ucode_go_auth_service/genproto/company_service"
	"ucode/ucode_go_auth_service/pkg/helper"
	"ucode/ucode_go_auth_service/pkg/logger"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *Handler) LoginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceId := c.GetHeader("Resource-Id")
		environmentId := c.GetHeader("Environment-Id")
		projectId := c.DefaultQuery("project-id", "")

		c.Set("resource_id", resourceId)
		c.Set("environment_id", environmentId)
		c.Set("project_id", projectId)
		fmt.Println("Project id->", projectId)
		//c.Set("namespace", h.cfg.UcodeNamespace)
		c.Next()
	}
}

func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceId := c.GetHeader("Resource-Id")
		environmentId := c.GetHeader("Environment-Id")
		projectId := c.DefaultQuery("project-id", "")
		bearerToken := c.GetHeader("Authorization")
		if len(bearerToken) == 0 {
			c.Set("resource_id", resourceId)
			c.Set("environment_id", environmentId)
			c.Set("project_id", projectId)
			//c.Set("namespace", h.cfg.UcodeNamespace)
			c.Next()
			return
		}

		strArr := strings.Split(bearerToken, " ")
		if len(strArr) < 1 && (strArr[0] != "Bearer" && strArr[0] != "API-KEY") {
			h.log.Error("---ERR->Unexpected token format")
			_ = c.AbortWithError(nethttp.StatusForbidden, errors.New("token error: wrong format"))
			return
		}

		switch strArr[0] {
		case "Bearer":
			res, ok := h.hasAccess(c)
			if !ok {
				h.log.Error("---ERR->AuthMiddleware->hasNotAccess-->")
				c.Abort()
				return
			}
			//}
			resourceId := c.GetHeader("Resource-Id")
			environmentId := c.GetHeader("Environment-Id")
			projectId := c.Query("Project-Id")

			if res.ProjectId != "" {
				projectId = res.ProjectId
			}
			if res.EnvId != "" {
				environmentId = res.EnvId
			}

			c.Set("resource_id", resourceId)
			c.Set("environment_id", environmentId)
			c.Set("project_id", projectId)
			c.Set("user_id", res.UserId)
		case "API-KEY":
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
		default:
			if !strings.Contains(c.Request.URL.Path, "api") {
				err := errors.New("error invalid authorization method")
				h.log.Error("--AuthMiddleware--", logger.Error(err))
				h.handleResponse(c, http.BadRequest, err.Error())
				c.Abort()
			} else {
				err := errors.New("error invalid authorization method")
				h.log.Error("--AuthMiddleware--", logger.Error(err))
				c.JSON(401, struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				}{
					Code:    401,
					Message: "The request requires an user authentication.",
				})
				c.Abort()
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

func (h *Handler) hasAccess(c *gin.Context) (*auth_service.V2HasAccessUserRes, bool) {
	bearerToken := c.GetHeader("Authorization")
	// projectId := c.DefaultQuery("project_id", "")

	strArr := strings.Split(bearerToken, " ")

	if len(strArr) != 2 || strArr[0] != "Bearer" {
		h.log.Error("---ERR->HasAccess->Unexpected token format")
		h.handleResponse(c, http.Forbidden, "token error: wrong format")
		return nil, false
	}
	accessToken := strArr[1]
	resp, err := h.services.SessionService().V2HasAccessUser(
		c.Request.Context(),
		&auth_service.V2HasAccessUserReq{
			AccessToken: accessToken,
			Path:        helper.GetURLWithTableSlug(c),
			Method:      c.Request.Method,
		},
	)
	if err != nil {
		errr := status.Error(codes.PermissionDenied, "Permission denied")
		if errr.Error() == err.Error() {
			h.log.Error("---ERR->HasAccess->Permission--->", logger.Error(err))
			h.handleResponse(c, http.BadRequest, err.Error())
			return nil, false
		}
		errr = status.Error(codes.InvalidArgument, "User has been expired")
		if errr.Error() == err.Error() {
			h.log.Error("---ERR->HasAccess->User Expired-->")
			h.handleResponse(c, http.Forbidden, err.Error())
			return nil, false
		}
		h.log.Error("---ERR->HasAccess->Session->V2HasAccessUser--->", logger.Error(err))
		h.handleResponse(c, http.Unauthorized, err.Error())
		return nil, false
	}

	return resp, true
}
